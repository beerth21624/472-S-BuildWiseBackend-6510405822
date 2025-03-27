package postgres

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type invoiceRepository struct {
	db *sqlx.DB
}

func NewInvoiceRepository(db *sqlx.DB) repositories.InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) ValidateProjectStatus(ctx context.Context, projectID uuid.UUID) error {
	query := `
        SELECT 
            p.status as project_status,
            b.status as boq_status,
            q.status as quotation_status,
            c.contract_id as contract_id
        FROM project p
        LEFT JOIN boq b ON b.project_id = p.project_id
        LEFT JOIN quotation q ON q.project_id = p.project_id
        LEFT JOIN contract c ON c.project_id = p.project_id
        WHERE p.project_id = $1`

	type StatusCheck struct {
		ProjectStatus   string         `db:"project_status"`
		BOQStatus       sql.NullString `db:"boq_status"`
		QuotationStatus sql.NullString `db:"quotation_status"`
		ContractID      uuid.NullUUID  `db:"contract_id"`
	}

	var status StatusCheck
	err := r.db.GetContext(ctx, &status, query, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project status: %w", err)
	}

	if status.ProjectStatus == "completed" {
		return errors.New("project is already completed")
	}

	if !status.BOQStatus.Valid || status.BOQStatus.String != "approved" {
		return errors.New("BOQ must be approved")
	}

	if !status.QuotationStatus.Valid || status.QuotationStatus.String != "approved" {
		return errors.New("quotation must be approved")
	}

	if !status.ContractID.Valid {
		return errors.New("contract must exist for the project")
	}

	return nil
}

func (r *invoiceRepository) CreateForAllPeriods(ctx context.Context, projectID uuid.UUID, contractID uuid.UUID, paymentTerm string) error {
	if err := r.ValidateProjectStatus(ctx, projectID); err != nil {
		return err
	}

	// Validate contract belongs to the project
	query := `
		SELECT c.contract_id
		FROM contract c
		WHERE c.project_id = $1 AND c.contract_id = $2
	`

	var result uuid.UUID
	err := r.db.GetContext(ctx, &result, query, projectID, contractID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("contract not found for this project")
		}
		return fmt.Errorf("failed to validate contract: %w", err)
	}

	// Get all periods for this contract that don't have invoices yet
	periodsQuery := `
		SELECT p.period_id, p.period_number, c.pay_within
		FROM period p
		JOIN contract c ON p.contract_id = c.contract_id
		WHERE p.contract_id = $1
		AND NOT EXISTS (
			SELECT 1 FROM invoice i WHERE i.period_id = p.period_id
		)
		ORDER BY p.period_number
	`

	type PeriodInfo struct {
		PeriodID     uuid.UUID `db:"period_id"`
		PeriodNumber int       `db:"period_number"`
		PayWithin    int       `db:"pay_within"`
	}

	var periods []PeriodInfo
	err = r.db.SelectContext(ctx, &periods, periodsQuery, contractID)
	if err != nil {
		return fmt.Errorf("failed to get contract periods: %w", err)
	}

	if len(periods) == 0 {
		return errors.New("no available periods found for invoicing in this contract")
	}

	// Begin a transaction for creating multiple invoices
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Insert invoice for each period
	insertQuery := `
        INSERT INTO invoice (
            invoice_id,
            project_id,
            period_id,
            invoice_date,
            payment_due_date,
            payment_term,
            status,
            created_at,
            updated_at
        ) VALUES (
            $1, $2, $3, CURRENT_DATE, 
            CURRENT_DATE + INTERVAL '1 day' * $4, 
            $5, 'draft', 
            CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
        )`

	for _, period := range periods {
		invoiceID := uuid.New()
		_, err = tx.ExecContext(ctx, insertQuery,
			invoiceID, projectID, period.PeriodID, period.PayWithin, paymentTerm)

		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create invoice for period %d: %w", period.PeriodNumber, err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *invoiceRepository) GetByID(ctx context.Context, invoiceID uuid.UUID) (*models.Invoice, error) {
	var invoice models.Invoice
	query := `SELECT * FROM invoice WHERE invoice_id = $1`
	err := r.db.GetContext(ctx, &invoice, query, invoiceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Get period details
	periodQuery := `SELECT * FROM period WHERE period_id = $1`
	err = r.db.GetContext(ctx, &invoice.Period, periodQuery, invoice.PeriodID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get period details: %w", err)
	}

	return &invoice, nil
}

func (r *invoiceRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.Invoice, error) {
	var invoices []models.Invoice
	query := `SELECT * FROM invoice WHERE project_id = $1 ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &invoices, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project invoices: %w", err)
	}

	// Get period details for each invoice
	for i := range invoices {
		periodQuery := `SELECT * FROM period WHERE period_id = $1`
		err = r.db.GetContext(ctx, &invoices[i].Period, periodQuery, invoices[i].PeriodID)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get period details: %w", err)
		}
	}

	return invoices, nil
}

func (r *invoiceRepository) UpdateStatus(ctx context.Context, invoiceID uuid.UUID, status string) error {
	query := `UPDATE invoice SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE invoice_id = $2`
	result, err := r.db.ExecContext(ctx, query, status, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("invoice not found")
	}

	return nil
}

func (r *invoiceRepository) Update(ctx context.Context, invoiceID uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return errors.New("no fields to update")
	}
	query := "UPDATE invoice SET "

	fields := []string{}
	values := []interface{}{}
	paramCount := 1

	for field, value := range updates {
		fields = append(fields, fmt.Sprintf("%s = $%d", field, paramCount))
		values = append(values, value)
		paramCount++
	}

	// Add updated_at field
	fields = append(fields, fmt.Sprintf("updated_at = $%d", paramCount))
	values = append(values, time.Now())
	paramCount++

	query += strings.Join(fields, ", ")

	query += fmt.Sprintf(" WHERE invoice_id = $%d", paramCount)
	values = append(values, invoiceID)

	result, err := r.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("invoice not found")
	}

	return nil
}
