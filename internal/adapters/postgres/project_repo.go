package postgres

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"boonkosang/internal/requests"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type projectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(db *sqlx.DB) repositories.ProjectRepository {
	return &projectRepository{
		db: db,
	}
}

func (r *projectRepository) Create(ctx context.Context, req requests.CreateProjectRequest) (*models.Project, error) {
	project := &models.Project{
		ProjectID:   uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Status:      models.ProjectStatusPlanning,
		ClientID:    req.ClientID,
		CreatedAt:   time.Now(),
	}

	query := `
        INSERT INTO Project (
            project_id, name, description, address, status, 
            client_id, created_at
        ) VALUES (
            :project_id, :name, :description, :address, :status,
            :client_id, :created_at
        ) RETURNING *`

	rows, err := r.db.NamedQueryContext(ctx, query, project)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.StructScan(project)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		return project, nil
	}
	return nil, errors.New("failed to create project: no rows returned")
}

func (r *projectRepository) Update(ctx context.Context, id uuid.UUID, req requests.UpdateProjectRequest) error {
	query := `
        UPDATE Project SET 
            name = :name,
            description = :description,
            address = :address,
			client_id = :client_id,
            updated_at = :updated_at
        WHERE project_id = :project_id`

	params := map[string]interface{}{
		"project_id":  id,
		"name":        req.Name,
		"description": req.Description,
		"address":     req.Address,
		"client_id":   req.ClientID,
		"updated_at":  time.Now(),
	}

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("project not found")
	}

	return nil
}

func (r *projectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM Project WHERE project_id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return errors.New("project not found")
	}
	return nil
}

func (r *projectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	project := &models.Project{}
	query := `SELECT * FROM Project WHERE project_id = $1`

	err := r.db.GetContext(ctx, project, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

func (r *projectRepository) GetByIDWithClient(ctx context.Context, id uuid.UUID) (*models.Project, *models.Client, error) {
	project := &models.Project{}
	client := &models.Client{}

	query := `
        SELECT 
            p.*,
            c.client_id as "client.client_id",
            c.name as "client.name",
            c.email as "client.email",
            c.tel as "client.tel",
            c.address as "client.address",
            c.tax_id as "client.tax_id"
        FROM Project p
        LEFT JOIN Client c ON p.client_id = c.client_id
        WHERE p.project_id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ProjectID, &project.Name, &project.Description,
		&project.Address, &project.Status, &project.ClientID,
		&project.CreatedAt, &project.UpdatedAt,
		&client.ClientID, &client.Name, &client.Email,
		&client.Tel, &client.Address, &client.TaxID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.New("project not found")
		}
		return nil, nil, fmt.Errorf("failed to get project with client: %w", err)
	}

	return project, client, nil
}

func (r *projectRepository) List(ctx context.Context) ([]models.Project, error) {
	var projects []models.Project

	query := `
		SELECT * FROM Project ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &projects, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return projects, nil
}

func (r *projectRepository) Cancel(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE Project SET 
			status = 'cancelled'
		WHERE project_id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to cancel project: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("project not found")
	}

	return nil
}

func (r *projectRepository) GetProjectStatus(ctx context.Context, projectID uuid.UUID) (*models.ProjectStatusCheck, error) {
	query := `
        SELECT 
            p.status as project_status,
            b.status as boq_status,
            q.status as quotation_status
        FROM project p
        LEFT JOIN boq b ON b.project_id = p.project_id
        LEFT JOIN quotation q ON q.project_id = p.project_id
        WHERE p.project_id = $1`

	var status models.ProjectStatusCheck
	err := r.db.GetContext(ctx, &status, query, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("project not found")
		}
		return nil, fmt.Errorf("failed to get project status: %w", err)
	}

	return &status, nil
}

func (r *projectRepository) ValidateStatusTransition(ctx context.Context, projectID uuid.UUID, newStatus models.ProjectStatus) error {
	status, err := r.GetProjectStatus(ctx, projectID)
	if err != nil {
		return err
	}

	// Validate BOQ and Quotation status
	if !status.BOQStatus.Valid || status.BOQStatus.String != "approved" {
		return errors.New("BOQ must be approved")
	}
	if !status.QuotationStatus.Valid || status.QuotationStatus.String != "approved" {
		return errors.New("quotation must be approved")
	}

	// Validate status transitions
	switch newStatus {
	case models.ProjectStatusInProgress:
		if status.ProjectStatus != string(models.ProjectStatusPlanning) {
			return errors.New("project must be in planning status to move to in_progress")
		}
	case models.ProjectStatusCompleted:
		if status.ProjectStatus != string(models.ProjectStatusInProgress) {
			return errors.New("project must be in in_progress status to move to completed")
		}
	}

	return nil
}

func (r *projectRepository) UpdateStatus(ctx context.Context, projectID uuid.UUID, status models.ProjectStatus) error {
	if err := r.ValidateStatusTransition(ctx, projectID, status); err != nil {
		return err
	}

	query := `
        UPDATE project 
        SET status = $1, updated_at = CURRENT_TIMESTAMP 
        WHERE project_id = $2`

	result, err := r.db.ExecContext(ctx, query, status, projectID)
	if err != nil {
		return fmt.Errorf("failed to update project status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("project not found")
	}

	return nil
}

func (r *projectRepository) ValidateProjectData(ctx context.Context, projectID uuid.UUID) error {
	query := `
        SELECT 
            COUNT(*) as total_materials,
            COUNT(CASE WHEN estimated_price IS NOT NULL AND actual_price IS NOT NULL THEN 1 END) as filled_materials
        FROM material_price_log mpl
        JOIN boq b ON b.boq_id = mpl.boq_id
        WHERE b.project_id = $1`

	type ValidationResult struct {
		TotalMaterials  int `db:"total_materials"`
		FilledMaterials int `db:"filled_materials"`
	}

	var result ValidationResult
	if err := r.db.GetContext(ctx, &result, query, projectID); err != nil {
		return fmt.Errorf("failed to validate project data: %w", err)
	}

	return nil
}

func (r *projectRepository) GetProjectOverview(ctx context.Context, projectID uuid.UUID) (*models.ProjectOverview, error) {
	if err := r.ValidateProjectData(ctx, projectID); err != nil {
		return nil, err
	}

	query := `
        WITH MaterialTotals AS (
            SELECT 
                job_id,
                boq_id,
                SUM(estimated_price * quantity) as total_material_price
            FROM material_price_log 
            GROUP BY job_id, boq_id
        ), GeneralCost AS (
            SELECT 
                b.boq_id, 
                SUM(gc.estimated_cost) as total_estimated_cost, 
                SUM(gc.actual_cost) as total_actual_cost 
            FROM boq b 
            LEFT JOIN general_cost gc ON gc.boq_id = b.boq_id 
            WHERE b.project_id = $1 
            GROUP BY b.boq_id
        ), JobTotals AS (
            SELECT 
                b.boq_id, 
                SUM(bj.selling_price*bj.quantity) as total_selling_price_exclude_gc_cost 
            FROM boq b 
            LEFT JOIN boq_job bj ON bj.boq_id = b.boq_id 
            WHERE b.project_id = $1 
            GROUP BY b.boq_id
        ), ActualPriceTotal AS (
            SELECT 
                job_id,
                boq_id,
                SUM(actual_price * quantity) as total_actual_price
            FROM material_price_log 
            GROUP BY job_id, boq_id
        )
        SELECT 
            q.quotation_id, 
            b.boq_id, 
            SUM((mt.total_material_price + bj.labor_cost) * bj.quantity) + gc.total_estimated_cost AS total_overall_cost,
            (jt.total_selling_price_exclude_gc_cost + b.selling_general_cost) as total_selling_price,
            q.tax_percentage,
            SUM((apt.total_actual_price + bj.labor_cost) * bj.quantity) + gc.total_actual_cost AS total_actual_cost
        FROM project p 
        LEFT JOIN quotation q ON q.project_id = p.project_id 
        LEFT JOIN boq b ON b.project_id = p.project_id 
        LEFT JOIN boq_job bj ON bj.boq_id = b.boq_id 
        LEFT JOIN MaterialTotals mt ON mt.job_id = bj.job_id AND mt.boq_id = bj.boq_id 
        LEFT JOIN GeneralCost gc ON gc.boq_id = b.boq_id 
        LEFT JOIN JobTotals jt ON jt.boq_id = b.boq_id 
        LEFT JOIN ActualPriceTotal apt ON apt.boq_id = bj.boq_id AND apt.job_id = bj.job_id 
        WHERE p.project_id = $1
        GROUP BY bj.boq_id, gc.total_estimated_cost, jt.total_selling_price_exclude_gc_cost, 
                q.tax_percentage, gc.total_actual_cost, q.quotation_id, b.boq_id`

	var overview models.ProjectOverview
	err := r.db.GetContext(ctx, &overview, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project overview: %w", err)
	}

	return &overview, nil
}

func (r *projectRepository) ValidateProjectStatus(ctx context.Context, projectID uuid.UUID) error {
	var status string
	query := `SELECT status FROM project WHERE project_id = $1`

	err := r.db.GetContext(ctx, &status, query, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project status: %w", err)
	}

	if status != "completed" {
		return errors.New("project must be completed to view summary")
	}

	return nil
}

func (r *projectRepository) GetProjectSummary(ctx context.Context, projectID uuid.UUID) (*models.ProjectSummary, error) {
	if err := r.ValidateProjectStatus(ctx, projectID); err != nil {
		return nil, err
	}

	// Get project overview first
	overview, err := r.GetProjectOverview(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Get job-level details
	jobs, err := r.getJobDetails(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &models.ProjectSummary{
		ProjectOverview: *overview,
		Jobs:            jobs,
	}, nil
}

func (r *projectRepository) getJobDetails(ctx context.Context, projectID uuid.UUID) ([]models.JobSummary, error) {
	query := `
        WITH MaterialTotals AS (
            SELECT 
                job_id,  
                boq_id, 
                COALESCE(SUM(estimated_price * quantity), 0) as total_material_price, 
                COALESCE(SUM(actual_price * quantity), 0) as total_actual_price 
            FROM material_price_log 
            GROUP BY job_id, boq_id
        )
        SELECT 
            COALESCE(b.selling_general_cost, 0) as selling_general_cost, 
            q.quotation_id, 
            q.status, 
            q.valid_date, 
            COALESCE(q.tax_percentage, 0) as tax_percentage, 
            j.name, 
            j.unit, 
            bj.quantity, 
            COALESCE(bj.labor_cost, 0) as labor_cost, 
            COALESCE(mt.total_material_price, 0) as total_material_price, 
            COALESCE(mt.total_material_price + bj.labor_cost, 0) as overall_cost, 
            COALESCE(bj.selling_price, 0) as selling_price,
            COALESCE(bj.selling_price - (mt.total_material_price + bj.labor_cost), 0) as estimated_profit, 
            COALESCE(mt.total_actual_price + bj.labor_cost, 0) as overall_actual_price, 
            COALESCE(bj.selling_price - (mt.total_actual_price + bj.labor_cost), 0) as job_profit, 
            COALESCE((bj.selling_price - (mt.total_actual_price + bj.labor_cost)) * bj.quantity, 0) as total_profit 
        FROM project p 
        LEFT JOIN quotation q ON q.project_id = p.project_id 
        JOIN boq b ON b.project_id = p.project_id 
        JOIN boq_job bj ON bj.boq_id = b.boq_id 
        JOIN job j ON j.job_id = bj.job_id 
        LEFT JOIN MaterialTotals mt ON mt.job_id = j.job_id AND mt.boq_id = b.boq_id 
        WHERE p.project_id = $1
        GROUP BY 
            q.quotation_id, q.status, q.valid_date, q.tax_percentage,
            j.name, j.unit, bj.quantity, bj.labor_cost, bj.selling_price, 
            mt.total_material_price, b.selling_general_cost, mt.total_actual_price`

	var jobs []models.JobSummary
	err := r.db.SelectContext(ctx, &jobs, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job details: %w", err)
	}

	return jobs, nil
}
