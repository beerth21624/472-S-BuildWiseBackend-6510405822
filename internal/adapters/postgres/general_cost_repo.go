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

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type generalCostRepository struct {
	db *sqlx.DB
}

func NewGeneralCostRepository(db *sqlx.DB) repositories.GeneralCostRepository {
	return &generalCostRepository{db: db}
}

func (r *generalCostRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*responses.GeneralCostListResponse, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get BOQ ID for the project
	var boqID uuid.UUID
	boqQuery := `SELECT boq_id FROM boq WHERE project_id = $1`
	err = tx.GetContext(ctx, &boqID, boqQuery, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("BOQ not found for this project")
		}
		return nil, fmt.Errorf("failed to get BOQ: %w", err)
	}

	// Get all available types
	var types []string
	typeQuery := `SELECT type_name FROM type`
	err = tx.SelectContext(ctx, &types, typeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get types: %w", err)
	}

	// Get existing general costs for this BOQ
	var existingCosts []models.GeneralCost
	existingQuery := `
        SELECT 
            g_id,
            boq_id,
            type_name,
            actual_cost,
            estimated_cost
        FROM general_cost
        WHERE boq_id = $1`

	err = tx.SelectContext(ctx, &existingCosts, existingQuery, boqID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing general costs: %w", err)
	}

	// Create a map of existing types for easy lookup
	existingTypes := make(map[string]bool)
	for _, cost := range existingCosts {
		existingTypes[cost.TypeName] = true
	}

	// Create general costs for missing types
	for _, typeName := range types {
		if !existingTypes[typeName] {
			// Create new general cost for this type
			newGID := uuid.New()
			insertQuery := `
                INSERT INTO general_cost (
                    g_id, boq_id, type_name, actual_cost, estimated_cost
                ) VALUES (
                    $1, $2, $3, $4, $5
                )`

			_, err = tx.ExecContext(ctx, insertQuery,
				newGID,
				boqID,
				typeName,
				0, // Default actual_cost
				0, // Default estimated_cost
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create general cost for type %s: %w", typeName, err)
			}
		}
	}

	// Commit the transaction if all operations are successful
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get all general costs after creation of missing ones
	var allGeneralCosts []models.GeneralCost
	finalQuery := `
        SELECT 
            gc.g_id,
            gc.boq_id,
            gc.type_name,
            gc.actual_cost,
            gc.estimated_cost
        FROM general_cost gc
        JOIN boq b ON gc.boq_id = b.boq_id
        WHERE b.project_id = $1
        ORDER BY gc.type_name`

	err = r.db.SelectContext(ctx, &allGeneralCosts, finalQuery, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get final general costs: %w", err)
	}

	// Convert to response format
	var response []responses.GeneralCostResponse
	for _, gc := range allGeneralCosts {
		response = append(response, responses.GeneralCostResponse{
			GID:           gc.GID,
			BOQID:         gc.BOQID,
			TypeName:      gc.TypeName,
			ActualCost:    gc.ActualCost.Float64,
			EstimatedCost: gc.EstimatedCost.Float64,
		})
	}

	return &responses.GeneralCostListResponse{
		GeneralCosts: response,
	}, nil
}

// Get general cost by ID
func (r *generalCostRepository) GetByID(ctx context.Context, gID uuid.UUID) (*models.GeneralCost, error) {
	query := `
       SELECT 
           g_id,
           boq_id,
           type_name,
           actual_cost,
           estimated_cost
       FROM general_cost
       WHERE g_id = $1`

	var generalCost models.GeneralCost
	err := r.db.GetContext(ctx, &generalCost, query, gID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("general cost not found")
		}
		return nil, fmt.Errorf("failed to get general cost: %w", err)
	}

	return &generalCost, nil
}

// Update general cost
func (r *generalCostRepository) Update(ctx context.Context, gID uuid.UUID, req requests.UpdateGeneralCostRequest) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check BOQ status first
	var boqStatus string
	statusQuery := `
       SELECT b.status
       FROM general_cost gc
       JOIN boq b ON b.boq_id = gc.boq_id
       WHERE gc.g_id = $1`

	err = tx.GetContext(ctx, &boqStatus, statusQuery, gID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("general cost not found")
		}
		return fmt.Errorf("failed to check BOQ status: %w", err)
	}

	if boqStatus != "draft" {
		return errors.New("can only update general cost for BOQ in draft status")
	}

	// Validate estimated cost
	if req.EstimatedCost < 0 {
		return errors.New("estimated cost must be positive")
	}

	// Update general cost
	updateQuery := `
       UPDATE general_cost 
       SET estimated_cost = $1
       WHERE g_id = $2`

	result, err := tx.ExecContext(ctx, updateQuery, req.EstimatedCost, gID)
	if err != nil {
		return fmt.Errorf("failed to update general cost: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("general cost not found")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *generalCostRepository) GetType(ctx context.Context) ([]models.Type, error) {
	query := `SELECT * FROM Type`

	var types []models.Type
	err := r.db.SelectContext(ctx, &types, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get types: %w", err)
	}

	return types, nil
}

func (r *generalCostRepository) UpdateActualCost(ctx context.Context, gID uuid.UUID, req requests.UpdateActualGeneralCostRequest) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get project status and validate
	var projectStatus struct {
		ProjectStatus   string    `db:"project_status"`
		BOQStatus       string    `db:"boq_status"`
		QuotationStatus string    `db:"quotation_status"`
		BOQID           uuid.UUID `db:"boq_id"`
	}

	query := `
        SELECT 
            p.status as project_status,
            b.status as boq_status,
            q.status as quotation_status,
            b.boq_id
        FROM general_cost gc
        JOIN boq b ON b.boq_id = gc.boq_id
        JOIN project p ON p.project_id = b.project_id
        LEFT JOIN quotation q ON q.project_id = p.project_id
        WHERE gc.g_id = $1`

	err = tx.GetContext(ctx, &projectStatus, query, gID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("general cost not found")
		}
		return fmt.Errorf("failed to get project status: %w", err)
	}

	// Validate project status
	if projectStatus.ProjectStatus == "completed" {
		return errors.New("cannot update actual cost for completed project")
	}

	// Validate BOQ and Quotation status
	if projectStatus.BOQStatus != "approved" {
		return errors.New("BOQ must be approved to update actual cost")
	}
	if projectStatus.QuotationStatus != "approved" {
		return errors.New("quotation must be approved to update actual cost")
	}

	// Update actual cost
	updateQuery := `
        UPDATE general_cost 
        SET actual_cost = $1
        WHERE g_id = $2`

	result, err := tx.ExecContext(ctx, updateQuery, req.ActualCost, gID)
	if err != nil {
		return fmt.Errorf("failed to update actual cost: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("general cost not found")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *generalCostRepository) ValidateProjectStatus(ctx context.Context, projectID uuid.UUID) error {
	query := `
        SELECT 
            p.status as project_status,
            b.status as boq_status,
            q.status as quotation_status
        FROM project p
        LEFT JOIN boq b ON b.project_id = p.project_id
        LEFT JOIN quotation q ON q.project_id = p.project_id
        WHERE p.project_id = $1`

	var status struct {
		ProjectStatus   string `db:"project_status"`
		BOQStatus       string `db:"boq_status"`
		QuotationStatus string `db:"quotation_status"`
	}

	err := r.db.GetContext(ctx, &status, query, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project status: %w", err)
	}

	if status.ProjectStatus == "completed" {
		return errors.New("cannot update actual cost for completed project")
	}
	if status.BOQStatus != "approved" {
		return errors.New("BOQ must be approved to update actual cost")
	}
	if status.QuotationStatus != "approved" {
		return errors.New("quotation must be approved to update actual cost")
	}

	return nil
}
