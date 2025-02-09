package postgres

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"boonkosang/internal/requests"
	"boonkosang/internal/responses"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type quotationRepository struct {
	db *sqlx.DB
}

func NewQuotationRepository(db *sqlx.DB) repositories.QuotationRepository {
	return &quotationRepository{db: db}
}

func (r *quotationRepository) CheckBOQStatus(ctx context.Context, projectID uuid.UUID) (string, error) {
	var status string
	query := `
        SELECT b.status 
        FROM boq b 
        WHERE b.project_id = $1`

	err := r.db.GetContext(ctx, &status, query, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("BOQ not found")
		}
		return "", fmt.Errorf("failed to check BOQ status: %w", err)
	}

	return status, nil
}

func (r *quotationRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*models.Quotation, error) {
	var quotation models.Quotation
	query := `SELECT * FROM quotation WHERE project_id = $1`

	err := r.db.GetContext(ctx, &quotation, query, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No quotation exists
		}
		return nil, fmt.Errorf("failed to get quotation: %w", err)
	}

	return &quotation, nil
}

func (r *quotationRepository) Create(ctx context.Context, projectID uuid.UUID) (*models.Quotation, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	quotation := &models.Quotation{
		QuotationID:   uuid.New(),
		ProjectID:     projectID,
		Status:        "draft",
		ValidDate:     sql.NullTime{Time: time.Now().AddDate(0, 1, 0), Valid: true}, // Default validity: 1 month
		TaxPercentage: sql.NullFloat64{Float64: 7, Valid: true},                     // Default tax percentage
	}

	query := `
        INSERT INTO quotation (
            quotation_id, project_id, status, valid_date, 
            final_amount, tax_percentage
        ) VALUES (
            :quotation_id, :project_id, :status, :valid_date, 
            :final_amount, :tax_percentage
        ) RETURNING *`

	_, err = tx.NamedExecContext(ctx, query, quotation)
	if err != nil {
		return nil, fmt.Errorf("failed to create quotation: %w", err)
	}

	err = tx.GetContext(ctx, quotation, "SELECT * FROM quotation WHERE quotation_id = $1", quotation.QuotationID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created quotation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return quotation, nil
}
func (r *quotationRepository) GetQuotationJobs(ctx context.Context, projectID uuid.UUID) ([]models.QuotationJob, error) {
	query := `
WITH MaterialTotals AS (
    SELECT 
        job_id, 
        boq_id, 
        SUM(estimated_price * quantity) as total_material_price 
    FROM material_price_log 
    GROUP BY job_id, boq_id
)
SELECT 
    q.quotation_id, 
    q.status, 
    q.valid_date, 
    q.tax_percentage, 
	 b.selling_general_cost,
	j.job_id,
    j.name, 
    j.unit, 
    bj.quantity, 
    bj.labor_cost, 
    mt.total_material_price,
    (mt.total_material_price + bj.labor_cost) as overall_cost,
    bj.selling_price,
    (mt.total_material_price + bj.labor_cost) * bj.quantity as total,
    (bj.selling_price * bj.quantity) as total_selling_price
FROM project p
LEFT JOIN quotation q ON q.project_id = p.project_id
JOIN boq b ON b.project_id = p.project_id
JOIN boq_job bj ON bj.boq_id = b.boq_id
JOIN job j ON j.job_id = bj.job_id
LEFT JOIN MaterialTotals mt ON mt.job_id = j.job_id AND mt.boq_id = b.boq_id
WHERE p.project_id = $1
GROUP BY 
    q.quotation_id, 
    q.status, 
    q.valid_date, 
    q.tax_percentage,
	 b.selling_general_cost,
	j.job_id,
    j.name, 
    j.unit, 
    bj.quantity, 
    bj.labor_cost, 
    bj.selling_price, 
    mt.total_material_price`

	var jobs []models.QuotationJob
	err := r.db.SelectContext(ctx, &jobs, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quotation jobs: %w", err)
	}

	return jobs, nil
}

func (r *quotationRepository) GetQuotationGeneralCosts(ctx context.Context, projectID uuid.UUID) ([]models.QuotationGeneralCost, error) {
	query := `
        SELECT 
            b.boq_id, gc.g_id, gc.type_name, gc.estimated_cost
        FROM project p
        JOIN boq b ON b.project_id = p.project_id
        LEFT JOIN general_cost gc ON gc.boq_id = b.boq_id
        LEFT JOIN "type" t ON t.type_name = gc.type_name
        WHERE p.project_id = $1`

	var costs []models.QuotationGeneralCost
	err := r.db.SelectContext(ctx, &costs, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quotation general costs: %w", err)
	}

	return costs, nil
}

func (r *quotationRepository) ValidateApproval(ctx context.Context, projectID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check BOQ status
	var boqStatus string
	boqQuery := `SELECT status FROM boq WHERE project_id = $1`
	err = tx.GetContext(ctx, &boqStatus, boqQuery, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("BOQ not found")
		}
		return fmt.Errorf("failed to get BOQ status: %w", err)
	}

	if boqStatus != "approved" {
		return errors.New("BOQ must be approved before approving quotation")
	}

	// Check quotation status
	var quotationStatus string
	quotationQuery := `SELECT status FROM quotation WHERE project_id = $1`
	err = tx.GetContext(ctx, &quotationStatus, quotationQuery, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("quotation not found")
		}
		return fmt.Errorf("failed to get quotation status: %w", err)
	}

	if quotationStatus != "draft" {
		return errors.New("only draft quotations can be approved")
	}

	return tx.Commit()
}

func (r *quotationRepository) ApproveQuotation(ctx context.Context, projectID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update quotation status
	query := `
        UPDATE quotation 
        SET status = 'approved'
        WHERE project_id = $1 AND status = 'draft'
        RETURNING quotation_id`

	var quotationID uuid.UUID
	err = tx.GetContext(ctx, &quotationID, query, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("no draft quotation found to approve")
		}
		return fmt.Errorf("failed to approve quotation: %w", err)
	}

	return tx.Commit()
}

func (r *quotationRepository) GetQuotationStatus(ctx context.Context, projectID uuid.UUID) (string, error) {
	var status string
	query := `SELECT status FROM quotation WHERE project_id = $1`

	err := r.db.GetContext(ctx, &status, query, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("quotation not found")
		}
		return "", fmt.Errorf("failed to get quotation status: %w", err)
	}

	return status, nil
}

func (r *quotationRepository) GetExportData(ctx context.Context, projectID uuid.UUID) (*responses.QuotationExportData, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get main quotation data
	query := `
        SELECT 
            p.project_id,
            p.name,
            p.description,
            p.address,
            c.name as client_name,
            c.address as client_address,
            c.email as client_email,
            c.tel as client_tel,
            c.tax_id as client_tax_id,
            q.quotation_id,
            q.valid_date,
            q.tax_percentage,
            q.final_amount,
            q.status
        FROM project p
        LEFT JOIN client c ON c.client_id = p.client_id
        LEFT JOIN quotation q ON q.project_id = p.project_id
        WHERE p.project_id = $1
            AND q.status = 'approved'
        LIMIT 1`

	var data responses.QuotationExportData
	err = tx.GetContext(ctx, &data, query, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("approved quotation not found")
		}
		return nil, fmt.Errorf("failed to get quotation data: %w", err)
	}

	// Get detailed job and cost information using the new query structure
	detailQuery := `
        WITH MaterialTotals AS (
            SELECT job_id, boq_id, SUM(estimated_price * quantity) as total_material_price 
            FROM material_price_log 
            GROUP BY job_id, boq_id
        )
        SELECT 
            b.selling_general_cost,
            j.name,
            j.description,
            j.unit,
            bj.quantity,
            bj.selling_price,
            (bj.selling_price * bj.quantity) as amount
        FROM project p
        JOIN boq b ON b.project_id = p.project_id
        JOIN boq_job bj ON bj.boq_id = b.boq_id
        JOIN job j ON j.job_id = bj.job_id
        LEFT JOIN MaterialTotals mt ON mt.job_id = j.job_id AND mt.boq_id = b.boq_id
        WHERE p.project_id = $1
        GROUP BY b.selling_general_cost, j.name, j.description, j.unit, bj.quantity, bj.selling_price`

	type jobDetailResult struct {
		SellingGeneralCost float64         `db:"selling_general_cost"`
		Name               string          `db:"name"`
		Description        string          `db:"description"`
		Unit               string          `db:"unit"`
		Quantity           float64         `db:"quantity"`
		SellingPrice       sql.NullFloat64 `db:"selling_price"`
		Amount             sql.NullFloat64 `db:"amount"`
	}

	var detailResults []jobDetailResult
	err = tx.SelectContext(ctx, &detailResults, detailQuery, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job details: %w", err)
	}

	// Process job details and calculate totals
	data.JobDetails = make([]responses.JobDetail, len(detailResults))
	var totalSellingPrice float64

	// Set selling general cost from the first result
	if len(detailResults) > 0 {
		data.SellingGeneralCost = detailResults[0].SellingGeneralCost
	}

	for i, result := range detailResults {
		data.JobDetails[i] = responses.JobDetail{
			Name:         result.Name,
			Description:  result.Description,
			Unit:         result.Unit,
			Quantity:     result.Quantity,
			SellingPrice: result.SellingPrice,
			Amount:       result.Amount,
		}

		if result.Amount.Valid {
			totalSellingPrice += result.Amount.Float64
		}
	}

	// Calculate subtotal including selling general cost and tax
	data.SubTotal = data.SellingGeneralCost + totalSellingPrice

	// Calculate tax amount if tax percentage exists
	if data.TaxPercentage > 0 {
		data.TaxAmount = data.SubTotal * data.TaxPercentage / 100

		// Update final amount if not already set
		if !data.FinalAmount.Valid {
			data.FinalAmount = sql.NullFloat64{
				Float64: data.SubTotal + data.TaxAmount,
				Valid:   true,
			}
		}
	}

	// Format all nullable fields
	data.FormatFinalAmount()
	for i := range data.JobDetails {
		if data.JobDetails[i].SellingPrice.Valid {
			value := data.JobDetails[i].SellingPrice.Float64
			data.JobDetails[i].FormattedSellingPrice = &value
		}
		if data.JobDetails[i].Amount.Valid {
			value := data.JobDetails[i].Amount.Float64
			data.JobDetails[i].FormattedAmount = &value
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &data, nil
}

func (r *quotationRepository) UpdateProjectSellingPrice(ctx context.Context, req requests.UpdateProjectSellingPriceRequest) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update quotation tax percentage
	query := `UPDATE quotation SET tax_percentage = $1 WHERE project_id = $2`
	_, err = tx.ExecContext(ctx, query, req.TaxPercentage, req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to update tax percentage: %w", err)
	}

	query = `UPDATE boq SET selling_general_cost = $1 WHERE project_id = $2`
	_, err = tx.ExecContext(ctx, query, req.SellingGeneralCost, req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to update selling general cost: %w", err)
	}

	// Get BOQ ID
	var boqID uuid.UUID
	query = `SELECT boq_id FROM boq WHERE project_id = $1`
	err = tx.GetContext(ctx, &boqID, query, req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get BOQ ID: %w", err)
	}

	// Update job selling prices
	for _, job := range req.JobSellingPrices {
		query = `UPDATE boq_job SET selling_price = $1 WHERE boq_id = $2 AND job_id = $3`
		_, err = tx.ExecContext(ctx, query, job.SellingPrice, boqID, job.JobID)
		if err != nil {
			return fmt.Errorf("failed to update job selling price: %w", err)
		}
	}

	// Update final amount
	query = `
        WITH ProjectCostData AS (
            SELECT 
                q.tax_percentage, 
                SUM(bj.selling_price * bj.quantity) + b.selling_general_cost as total_selling_price 
            FROM project p 
            JOIN boq b ON b.project_id = p.project_id 
            JOIN quotation q ON q.project_id = p.project_id 
            LEFT JOIN boq_job bj ON bj.boq_id = b.boq_id 
            WHERE p.project_id = $1 
            GROUP BY bj.boq_id, q.tax_percentage , b.selling_general_cost
        )
        UPDATE quotation 
        SET final_amount = (
            SELECT (ProjectCostData.tax_percentage * ProjectCostData.total_selling_price / 100) + ProjectCostData.total_selling_price 
            FROM ProjectCostData
        ) 
        WHERE project_id = $1`

	_, err = tx.ExecContext(ctx, query, req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to update final amount: %w", err)
	}

	return tx.Commit()
}
