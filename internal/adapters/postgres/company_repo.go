package postgres

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type companyRepository struct {
	db *sqlx.DB
}

func NewCompanyRepository(db *sqlx.DB) repositories.CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) GetOrCreateCompanyByUserID(ctx context.Context, userID uuid.UUID) (*models.Company, error) {
	// Start transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// First, check if user exists and get their info
	var user struct {
		CompanyID *uuid.UUID `db:"company_id"`
		FirstName string     `db:"first_name"`
		LastName  string     `db:"last_name"`
		Email     string     `db:"email"`
	}

	userQuery := `
        SELECT company_id, first_name, last_name, email 
        FROM "User" 
        WHERE user_id = $1
        FOR UPDATE`

	err = tx.GetContext(ctx, &user, userQuery, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// If user has company_id, get the company
	if user.CompanyID != nil {
		var company models.Company
		companyQuery := `
            SELECT * FROM company 
            WHERE company_id = $1`

		err = tx.GetContext(ctx, &company, companyQuery, *user.CompanyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get company: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		return &company, nil
	}

	// If user doesn't have a company, create one
	// Create default address
	defaultAddress := map[string]string{
		"house_number": "",
		"soi":          "",
		"moo":          "",
		"road":         "",
		"province":     "",
		"district":     "",
		"sub_district": "",
		"postal_code":  "",
	}
	addressJSON, err := json.Marshal(defaultAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create default address: %w", err)
	}

	newCompany := &models.Company{
		CompanyID: uuid.New(),
		Name:      fmt.Sprintf("%s %s Company", user.FirstName, user.LastName),
		Email:     user.Email,
		Tel:       "",
		Address:   addressJSON,
		TaxID:     "",
	}

	// Insert new company
	insertCompanyQuery := `
        INSERT INTO company (company_id, name, email, tel, address, tax_id)
        VALUES (:company_id, :name, :email, :tel, :address, :tax_id)
        RETURNING *`

	rows, err := r.db.NamedQueryContext(ctx, insertCompanyQuery, newCompany)
	if err != nil {
		return nil, fmt.Errorf("failed to create company: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("failed to create company: no rows returned")
	}

	// Update user with new company_id
	updateUserQuery := `
        UPDATE "User" 
        SET company_id = $1 
        WHERE user_id = $2`

	if _, err := tx.ExecContext(ctx, updateUserQuery, newCompany.CompanyID, userID); err != nil {
		return nil, fmt.Errorf("failed to update user company: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newCompany, nil
}

func (r *companyRepository) UpdateCompany(ctx context.Context, company *models.Company) error {

	query := `
        UPDATE company 
        SET name = :name,
            email = :email,
            tel = :tel,
            address = :address,
            tax_id = :tax_id
        WHERE company_id = :company_id`

	result, err := r.db.NamedExecContext(ctx, query, company)
	if err != nil {
		return fmt.Errorf("failed to update company: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("company not found")
	}

	return nil
}
