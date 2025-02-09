package postgres

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"boonkosang/internal/requests"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type clientRepository struct {
	db *sqlx.DB
}

func NewClientRepository(db *sqlx.DB) repositories.ClientRepository {
	return &clientRepository{
		db: db,
	}
}

func (r *clientRepository) Create(ctx context.Context, req requests.CreateClientRequest) (*models.Client, error) {
	client := &models.Client{
		ClientID: uuid.New(),
		Name:     req.Name,
		Email:    req.Email,
		Tel:      req.Tel,
		Address:  req.Address,
		TaxID:    req.TaxID,
	}

	query := `
        INSERT INTO Client (
            client_id, name, email, tel, address, tax_id
          
        ) VALUES (
            :client_id, :name, :email, :tel, :address, :tax_id
        ) RETURNING *`

	rows, err := r.db.NamedQueryContext(ctx, query, client)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.New("client with this email already exists")
		}
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.StructScan(client)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		return client, nil
	}
	return nil, errors.New("failed to create client: no rows returned")
}

func (r *clientRepository) Update(ctx context.Context, id uuid.UUID, req requests.UpdateClientRequest) error {
	query := `
        UPDATE Client SET 
            name = :name,
            email = :email,
            tel = :tel,
            address = :address,
            tax_id = :tax_id
        WHERE client_id = :client_id`

	params := map[string]interface{}{
		"client_id": id,
		"name":      req.Name,
		"email":     req.Email,
		"tel":       req.Tel,
		"address":   req.Address,
		"tax_id":    req.TaxID,
	}

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return errors.New("client with this email already exists")
		}
		return fmt.Errorf("failed to update client: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("client not found")
	}

	return nil
}

func (r *clientRepository) Delete(ctx context.Context, id uuid.UUID) error {

	checkUsageQuery := `
		SELECT p.name as project_name
		FROM Client c 
		JOIN Project p ON c.client_id = p.client_id 
		WHERE c.client_id = $1`

	type ProjectUsage struct {
		ProjectName string `db:"project_name"`
	}
	var projects []ProjectUsage

	err := r.db.SelectContext(ctx, &projects, checkUsageQuery, id)
	if err != nil {
		return fmt.Errorf("failed to check client usage: %w", err)
	}

	if len(projects) > 0 {
		var projectNames []string
		for _, project := range projects {
			projectNames = append(projectNames, project.ProjectName)
		}
		return fmt.Errorf("client is currently used in following projects: %s. Please remove client from these projects before deletion",
			strings.Join(projectNames, ", "))
	}

	query := `DELETE FROM Client WHERE client_id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("client not found")
	}

	return nil
}

func (r *clientRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Client, error) {
	client := &models.Client{}
	query := `SELECT * FROM Client WHERE client_id = $1`

	err := r.db.GetContext(ctx, client, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("client not found")
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return client, nil
}

func (r *clientRepository) GetByEmail(ctx context.Context, email string) (*models.Client, error) {
	client := &models.Client{}
	query := `SELECT * FROM Client WHERE email = $1`

	err := r.db.GetContext(ctx, client, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("client not found")
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return client, nil
}

func (r *clientRepository) List(ctx context.Context, limit, offset int) ([]models.Client, int64, error) {
	var clients []models.Client
	var total int64

	countQuery := `SELECT COUNT(*) FROM Client`
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	query := `
        SELECT * FROM Client 
        LIMIT $1 OFFSET $2`

	err = r.db.SelectContext(ctx, &clients, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list clients: %w", err)
	}

	return clients, total, nil
}
