// repository/job_repository.go
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
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type jobRepository struct {
	db *sqlx.DB
}

func NewJobRepository(db *sqlx.DB) repositories.JobRepository {
	return &jobRepository{db: db}
}
func (r *jobRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Job, error) {
	var job models.Job
	query := `SELECT * FROM Job WHERE job_id = $1`
	err := r.db.GetContext(ctx, &job, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("job not found")
		}
		return nil, fmt.Errorf("failed to get job by ID: %w", err)
	}
	return &job, nil
}

func (r *jobRepository) GetJobMaterialByID(ctx context.Context, id uuid.UUID) (responses.JobMaterialResponse, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return responses.JobMaterialResponse{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	type JobQuery struct {
		JobID       uuid.UUID      `db:"job_id"`
		Name        string         `db:"name"`
		Description sql.NullString `db:"description"`
		Unit        string         `db:"unit"`
	}

	// Get job details
	var jobQuery JobQuery
	jobQueryString := `
        SELECT 
            j.job_id,
            j.name,
            j.description,
            j.unit
        FROM Job j
        WHERE j.job_id = $1`

	err = tx.GetContext(ctx, &jobQuery, jobQueryString, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return responses.JobMaterialResponse{}, errors.New("job not found")
		}
		return responses.JobMaterialResponse{}, fmt.Errorf("failed to get job: %w", err)
	}

	// Get materials for the job
	materialsQuery := `
        SELECT 
            m.material_id,
            m.name,
            m.unit,
            jm.quantity
        FROM Material m
        JOIN Job_material jm ON m.material_id = jm.material_id
        WHERE jm.job_id = $1`

	var materials []responses.JobMaterialItem
	err = tx.SelectContext(ctx, &materials, materialsQuery, id)
	if err != nil {
		return responses.JobMaterialResponse{}, fmt.Errorf("failed to get job materials: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return responses.JobMaterialResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	var materialsForResponse []responses.JobMaterialItem
	for _, material := range materials {
		materialsForResponse = append(materialsForResponse, responses.JobMaterialItem{
			MaterialID: material.MaterialID,
			Name:       material.Name,
			Unit:       material.Unit,
			Quantity:   material.Quantity,
		})
	}

	job := responses.JobMaterialResponse{
		JobID:       jobQuery.JobID,
		Name:        jobQuery.Name,
		Description: jobQuery.Description.String,
		Unit:        jobQuery.Unit,
		Materials:   materialsForResponse,
	}

	return job, nil
}

// List retrieves all jobs
func (r *jobRepository) List(ctx context.Context) (*responses.JobListResponse, error) {
	var jobs []models.Job
	query := `SELECT * FROM Job`
	err := r.db.SelectContext(ctx, &jobs, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	var jobList []responses.JobResponse
	for _, job := range jobs {
		jobList = append(jobList, responses.JobResponse{
			JobID:       job.JobID,
			Name:        job.Name,
			Description: job.Description.String,
			Unit:        job.Unit,
		})
	}

	return &responses.JobListResponse{Jobs: jobList}, nil
}

// CreateJob creates a new job without materials
func (r *jobRepository) Create(ctx context.Context, req requests.CreateJobRequest) (*responses.JobResponse, error) {
	job := &models.Job{
		JobID:       uuid.New(),
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		Unit:        req.Unit,
	}

	query := `
        INSERT INTO Job (
            job_id, name, description, unit
        ) VALUES (
            :job_id, :name, :description, :unit
        ) RETURNING *`

	rows, err := r.db.NamedQueryContext(ctx, query, job)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.StructScan(job)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		return &responses.JobResponse{
			JobID:       job.JobID,
			Name:        job.Name,
			Description: job.Description.String,
			Unit:        job.Unit,
		}, nil
	}

	return nil, errors.New("failed to create job: no rows returned")
}

func (r *jobRepository) Update(ctx context.Context, id uuid.UUID, req requests.UpdateJobRequest) error {
	query := `
        UPDATE Job SET 
            name = :name,
            description = :description,
            unit = :unit
        WHERE job_id = :job_id`

	params := map[string]interface{}{
		"job_id":      id,
		"name":        req.Name,
		"description": req.Description,
		"unit":        req.Unit,
	}

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("job not found")
	}

	return nil
}

func (r *jobRepository) Delete(ctx context.Context, jobID uuid.UUID) error {
	// 5. Check if job is used in any BOQs
	query := `
        SELECT DISTINCT 
            p.project_id,
            p.name as project_name,
            b.boq_id,
            b.status as boq_status
        FROM boq_job bj 
        JOIN boq b ON b.boq_id = bj.boq_id 
        JOIN project p ON p.project_id = b.project_id 
        WHERE bj.job_id = $1`

	var projects []responses.ProjectUsage
	err := r.db.SelectContext(ctx, &projects, query, jobID)
	if err != nil {
		return fmt.Errorf("failed to get associated projects: %w", err)
	}

	if len(projects) > 0 {
		var projectNames []string
		for _, project := range projects {
			projectNames = append(projectNames, project.ProjectName)
		}
		return fmt.Errorf("job is used in projects: %s", strings.Join(projectNames, ", "))
	}

	// 7. If job is not used, proceed with deletion
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete job materials first due to foreign key constraint
	deleteMaterialsQuery := `DELETE FROM Job_material WHERE job_id = $1`
	_, err = tx.ExecContext(ctx, deleteMaterialsQuery, jobID)
	if err != nil {
		return fmt.Errorf("failed to delete job materials: %w", err)
	}
	// Delete material price logs
	deletePriceLogsQuery := `
		DELETE FROM Material_price_log
		WHERE job_id = $1`
	_, err = tx.ExecContext(ctx, deletePriceLogsQuery, jobID)
	if err != nil {
		return fmt.Errorf("failed to delete material price logs: %w", err)
	}

	// Delete the job
	deleteJobQuery := `DELETE FROM Job WHERE job_id = $1`
	result, err := tx.ExecContext(ctx, deleteJobQuery, jobID)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("job not found")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *jobRepository) AddJobMaterial(ctx context.Context, jobID uuid.UUID, req requests.AddJobMaterialRequest) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertJobMaterialQuery := `
		INSERT INTO Job_material (
			job_id, material_id, quantity
		) VALUES (
			:job_id, :material_id, :quantity
		) ON CONFLICT (job_id, material_id) 
		DO UPDATE SET quantity = Job_material.quantity + EXCLUDED.quantity`

	getProjectsQuery := `
		SELECT DISTINCT b.boq_id, b.status
		FROM boq_job bj 
		JOIN boq b ON b.boq_id = bj.boq_id 
		JOIN project p ON p.project_id = b.project_id 
		WHERE bj.job_id = $1`

	type BOQInfo struct {
		BOQID  uuid.UUID `db:"boq_id"`
		Status string    `db:"status"`
	}
	var boqs []BOQInfo

	err = tx.SelectContext(ctx, &boqs, getProjectsQuery, jobID)
	if err != nil {
		return fmt.Errorf("failed to get associated projects: %w", err)
	}

	// Insert job materials
	for _, material := range req.Materials {
		params := map[string]interface{}{
			"job_id":      jobID,
			"material_id": material.MaterialID,
			"quantity":    material.Quantity,
		}

		_, err := tx.NamedExecContext(ctx, insertJobMaterialQuery, params)
		fmt.Print(err)

		if err != nil {

			return fmt.Errorf("failed to add material: %w", err)
		}

		// 14.4 For each draft BOQ, create material price log entries
		for _, boq := range boqs {
			if boq.Status == "draft" {
				insertPriceLogQuery := `
					INSERT INTO Material_price_log (
						material_id, boq_id, supplier_id, actual_price, estimated_price, job_id, quantity, updated_at
					) VALUES (
						$1, $2, NULL, NULL, NULL, $3, $4, CURRENT_TIMESTAMP
					)`

				_, err = tx.ExecContext(ctx, insertPriceLogQuery, material.MaterialID, boq.BOQID, jobID, material.Quantity)
				if err != nil {
					return fmt.Errorf("failed to create material price log: %w", err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *jobRepository) DeleteJobMaterial(ctx context.Context, jobID uuid.UUID, materialID string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	deleteJobMaterialQuery := `
		DELETE FROM Job_material 
		WHERE job_id = $1 AND material_id = $2`

	result, err := tx.ExecContext(ctx, deleteJobMaterialQuery, jobID, materialID)
	if err != nil {
		return fmt.Errorf("failed to delete job material: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("job material not found")
	}

	deletePriceLogsQuery := `
		DELETE FROM material_price_log 
		WHERE job_id = $1 
		AND material_id = $2 
		AND boq_id IN (
			SELECT b.boq_id 
			FROM boq b 
			JOIN boq_job bj ON b.boq_id = bj.boq_id 
			WHERE bj.job_id = $1 
			AND b.status = 'draft'
		)`

	_, err = tx.ExecContext(ctx, deletePriceLogsQuery, jobID, materialID)
	if err != nil {
		return fmt.Errorf("failed to delete material price logs: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *jobRepository) UpdateJobMaterialQuantity(ctx context.Context, jobID uuid.UUID, req requests.UpdateJobMaterialQuantityRequest) error {

	updateJobMaterialQuery := `
		UPDATE job_material 
		SET quantity = :quantity
		WHERE job_id = :job_id AND material_id = :material_id`

	params := map[string]interface{}{
		"job_id":      jobID,
		"material_id": req.MaterialID,
		"quantity":    req.Quantity,
	}

	result, err := r.db.NamedExecContext(ctx, updateJobMaterialQuery, params)
	if err != nil {
		return fmt.Errorf("failed to update material quantity: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("job material not found")
	}

	updateMaterialPriceLogQuery := `
		UPDATE material_price_log mpl 
		SET quantity = :quantity
		WHERE mpl.job_id = :job_id 
		AND mpl.material_id = :material_id 
		AND boq_id IN (
			SELECT b.boq_id 
			FROM boq b 
			JOIN boq_job bj ON b.boq_id = bj.boq_id 
			JOIN job_material jm ON jm.job_id = bj.job_id 
			WHERE jm.job_id = :job_id 
			AND jm.material_id = :material_id 
			AND b.status = 'draft'
		)`

	_, err = r.db.NamedExecContext(ctx, updateMaterialPriceLogQuery, params)
	if err != nil {
		return fmt.Errorf("failed to update material price log: %w", err)
	}

	return nil
}

type UpdateJobMaterialQuantityRequest struct {
	MaterialID uuid.UUID `json:"material_id" validate:"required"`
	Quantity   int       `json:"quantity" validate:"required,gt=0"`
}
