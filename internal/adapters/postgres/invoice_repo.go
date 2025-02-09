package postgres

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"context"
	"database/sql"
	"errors"
	"fmt"

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
            q.status as quotation_status
        FROM project p
        LEFT JOIN boq b ON b.project_id = p.project_id
        LEFT JOIN quotation q ON q.project_id = p.project_id
        WHERE p.project_id = $1`

	type StatusCheck struct {
		ProjectStatus   string         `db:"project_status"`
		BOQStatus       sql.NullString `db:"boq_status"`
		QuotationStatus sql.NullString `db:"quotation_status"`
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

	return nil
}

func (r *invoiceRepository) Create(ctx context.Context, projectID uuid.UUID, fileURL string) error {
	if err := r.ValidateProjectStatus(ctx, projectID); err != nil {
		return err
	}

	query := `
        INSERT INTO invoice (
            invoice_id,
            project_id,
            file_url,
            created_at,
            updated_at
        ) VALUES (
            $1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
        )`

	_, err := r.db.ExecContext(ctx, query, uuid.New(), projectID, fileURL)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	return nil
}

func (r *invoiceRepository) Delete(ctx context.Context, invoiceID uuid.UUID) error {
	query := `DELETE FROM invoice WHERE invoice_id = $1`
	result, err := r.db.ExecContext(ctx, query, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
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
	return &invoice, nil
}

func (r *invoiceRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.Invoice, error) {
	var invoices []models.Invoice
	query := `SELECT * FROM invoice WHERE project_id = $1`
	err := r.db.SelectContext(ctx, &invoices, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project invoices: %w", err)
	}
	return invoices, nil
}
