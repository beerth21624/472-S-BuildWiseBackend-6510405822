package postgres

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type contractRepository struct {
	db *sqlx.DB
}

func NewContractRepository(db *sqlx.DB) repositories.ContractRepository {
	return &contractRepository{db: db}
}

// Create ...
func (r *contractRepository) Create(ctx context.Context, contract *models.Contract) error {
	query := `
		INSERT INTO contract (
			contract_id, 
			project_id, 
			format,
			created_at
		) VALUES (
			:contract_id, 
			:project_id, 
			:format,
			:created_at
		)`

	formatJSON, err := json.Marshal(contract.Format)
	if err != nil {
		return fmt.Errorf("failed to marshal format: %w", err)
	}

	params := map[string]interface{}{
		"contract_id": uuid.New(),
		"project_id":  contract.ProjectID,
		"format":      string(formatJSON),
		"created_at":  time.Now(),
	}

	_, err = r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to create contract: %w", err)
	}

	return nil
}

func (r *contractRepository) Update(ctx context.Context, contract *models.Contract) error {
	// Build dynamic query and params based on which fields are being updated
	var setFields []string
	params := make(map[string]interface{})

	// Helper function to add field to update
	addField := func(fieldName, paramName string, value interface{}, valid bool) {
		if valid {
			setFields = append(setFields, fieldName+" = :"+paramName)
			params[paramName] = value
		}
	}

	addField("project_description", "project_description", contract.ProjectDescription.String, contract.ProjectDescription.Valid)
	addField("area_size", "area_size", contract.AreaSize.Float64, contract.AreaSize.Valid)
	addField("start_date", "start_date", contract.StartDate.Time, contract.StartDate.Valid)
	addField("end_date", "end_date", contract.EndDate.Time, contract.EndDate.Valid)
	addField("force_majeure", "force_majeure", contract.ForceMajeure.String, contract.ForceMajeure.Valid)
	addField("breach_of_contract", "breach_of_contract", contract.BreachOfContract.String, contract.BreachOfContract.Valid)
	addField("end_of_contract", "end_of_contract", contract.EndOfContract.String, contract.EndOfContract.Valid)
	addField("termination_of_contract", "termination_of_contract", contract.TerminationContract.String, contract.TerminationContract.Valid)
	addField("amendment", "amendment", contract.Amendment.String, contract.Amendment.Valid)
	addField("guarantee_within", "guarantee_within", contract.GuaranteeWithin.Int32, contract.GuaranteeWithin.Valid)
	addField("retention_money", "retention_money", contract.RetentionMoney.Float64, contract.RetentionMoney.Valid)
	addField("pay_within", "pay_within", contract.PayWithin.Int32, contract.PayWithin.Valid)
	addField("validate_within", "validate_within", contract.ValidateWithin.Int32, contract.ValidateWithin.Valid)

	// Handle format field specially
	if len(contract.Format) > 0 {
		formatJSON, err := json.Marshal(contract.Format)
		if err != nil {
			return fmt.Errorf("failed to marshal format: %w", err)
		}
		addField("format", "format", string(formatJSON), true)
	}

	// Always update updated_at
	now := time.Now()
	addField("updated_at", "updated_at", now, true)

	// Add contract_id for WHERE clause
	params["contract_id"] = contract.ContractID

	// If no fields to update, return early
	if len(setFields) == 0 {
		return nil
	}

	// Construct the final query
	query := fmt.Sprintf(`
		UPDATE contract 
		SET %s 
		WHERE contract_id = :contract_id
	`, strings.Join(setFields, ", "))

	// Execute the query
	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return errors.New("contract with this ID already exists")
		}
		return fmt.Errorf("failed to update contract: %w", err)
	}

	// Check if any rows were affected
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("contract not found")
	}

	return nil
}

// Delete ...
func (r *contractRepository) Delete(ctx context.Context, projectID uuid.UUID) error {
	query := `DELETE FROM contract WHERE project_id = $1`
	result, err := r.db.ExecContext(ctx, query, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete contract: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("contract not found")
	}

	return nil
}

// GetByProjectID ...
func (r *contractRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*models.Contract, error) {
	var contract models.Contract
	query := `SELECT * FROM contract WHERE project_id = $1`

	err := r.db.GetContext(ctx, &contract, query, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("contract not found")
		}
		return nil, fmt.Errorf("failed to get contract: %w", err)
	}

	return &contract, nil
}

// ValidateProjectStatus ...
func (r *contractRepository) ValidateProjectStatus(ctx context.Context, projectID uuid.UUID) error {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM contract WHERE project_id = $1)`
	err := r.db.GetContext(ctx, &exists, query, projectID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("project does not exist")
	}
	return nil
}
