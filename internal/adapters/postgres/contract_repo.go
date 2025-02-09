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

type contractRepository struct {
	db *sqlx.DB
}

func NewContractRepository(db *sqlx.DB) repositories.ContractRepository {
	return &contractRepository{db: db}
}

// Create ...
func (r *contractRepository) Create(ctx context.Context, contract *models.Contract) error {
	query := `INSERT INTO contract (
		contract_id, project_id, created_at
	) VALUES (
		:contract_id, :project_id, :created_at
	)`
	params := map[string]interface{}{
		"contract_id": uuid.New(),
		"project_id":  contract.ProjectID,
		"created_at":  time.Now(),
	}
	_, err := r.db.NamedExecContext(ctx, query, params)
	return err
}

// Update ...
func (r *contractRepository) Update(ctx context.Context, contract *models.Contract) error {
	query := `
		UPDATE contract SET
			project_description = :project_description,
			area_size = :area_size,
			start_date = :start_date,
			end_date = :end_date,
			force_majeure = :force_majeure,
			breach_of_contract = :breach_of_contract,
			end_of_contract = :end_of_contract,
			termination_of_contract = :termination_of_contract,
			amendment = :amendment,
			guarantee_within = :guarantee_within,
			retention_money = :retention_money,
			pay_within = :pay_within,
			validate_within = :validate_within,
			format = :format,
			updated_at = :updated_at
		WHERE contract_id = :contract_id`

	params := map[string]interface{}{
		"contract_id":             contract.ContractID,
		"project_description":     contract.ProjectDescription,
		"area_size":               contract.AreaSize,
		"start_date":              contract.StartDate,
		"end_date":                contract.EndDate,
		"force_majeure":           contract.ForceMajeure,
		"breach_of_contract":      contract.BreachOfContract,
		"end_of_contract":         contract.EndOfContract,
		"termination_of_contract": contract.TerminationContract,
		"amendment":               contract.Amendment,
		"guarantee_within":        contract.GuaranteeWithin,
		"retention_money":         contract.RetentionMoney,
		"pay_within":              contract.PayWithin,
		"validate_within":         contract.ValidateWithin,
		"format":                  contract.Format,
		"updated_at":              contract.UpdatedAt,
	}

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return errors.New("contract with this ID already exists")
		}
		return fmt.Errorf("failed to update contract: %w", err)
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
