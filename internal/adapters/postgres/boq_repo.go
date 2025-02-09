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

type boqRepository struct {
	db *sqlx.DB
}

func NewBOQRepository(db *sqlx.DB) repositories.BOQRepository {
	return &boqRepository{
		db: db,
	}
}

func (r *boqRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.BOQ, error) {
	var boq models.BOQ
	query := `SELECT * FROM boq WHERE boq_id = $1`
	err := r.db.GetContext(ctx, &boq, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get BOQ: %w", err)
	}

	return &boq, nil
}

func (r *boqRepository) Approve(ctx context.Context, boqID uuid.UUID) error {
	// Start transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)

	}

	// Check BOQ status
	var status string
	checkStatusQuery := `SELECT status FROM boq WHERE boq_id = $1`
	err = tx.GetContext(ctx, &status, checkStatusQuery, boqID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("boq not found")
		}
		return fmt.Errorf("failed to get BOQ status: %w", err)
	}

	if status != "draft" {
		return errors.New("can only approve BOQ in draft status")

	}

	// Update BOQ status
	updateQuery := `UPDATE boq SET status = 'approved' WHERE boq_id = $1`
	_, err = tx.ExecContext(ctx, updateQuery, boqID)
	if err != nil {
		return fmt.Errorf("failed to update BOQ status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *boqRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*models.BOQ, error) {
	var boq models.BOQ
	query := `SELECT * FROM boq WHERE project_id = $1`
	err := r.db.GetContext(ctx, &boq, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get BOQ: %w", err)
	}

	return &boq, nil
}

func (r *boqRepository) GetBoqWithProject(ctx context.Context, projectID uuid.UUID) (*responses.BOQResponse, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var data models.BOQ

	boqQuery := `
        SELECT  boq_id, project_id, status, selling_general_cost
		FROM Boq
		WHERE project_id = $1`

	err = tx.GetContext(ctx, &data, boqQuery, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create new BOQ if it doesn't exist
			createBOQQuery := `
                INSERT INTO Boq (project_id, status, selling_general_cost) 
                VALUES (:project_id, 'draft', NULL) 
                RETURNING boq_id, project_id, status, selling_general_cost`

			row, err := r.db.NamedQueryContext(ctx, createBOQQuery, map[string]interface{}{
				"project_id": projectID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create new BOQ: %w", err)
			}

			if row.Next() {
				err = row.StructScan(&data)
				if err != nil {
					return nil, fmt.Errorf("failed to scan BOQ: %w", err)
				}
			}

			if err := row.Close(); err != nil {
				return nil, fmt.Errorf("failed to close row: %w", err)
			}

		} else {
			return nil, fmt.Errorf("failed to check BOQ existence: %w", err)
		}
	}

	// Convert to response struct
	response := &responses.BOQResponse{
		ID:                 data.BOQID, // Assuming the correct field name is BOQID
		ProjectID:          data.ProjectID,
		Status:             data.Status, // Assuming the correct field name is Status
		SellingGeneralCost: data.SellingGeneralCost.Float64,
	}

	jobsQuery := `
   SELECT DISTINCT
	j.*, bj.quantity, bj.labor_cost
FROM job j
JOIN boq_job bj ON j.job_id = bj.job_id
WHERE bj.boq_id = $1
`

	type BoqJobData struct {
		JobID       uuid.UUID      `db:"job_id"`
		Name        string         `db:"name"`
		Description sql.NullString `db:"description"`
		Unit        string         `db:"unit"`
		Quantity    float64        `db:"quantity"`
		LaborCost   float64        `db:"labor_cost"`
	}

	var jobs []BoqJobData

	err = tx.SelectContext(ctx, &jobs, jobsQuery, data.BOQID)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}

	var jobForResponse []responses.JobResponse
	for _, job := range jobs {
		jobForResponse = append(jobForResponse, responses.JobResponse{
			JobID:       job.JobID,
			Name:        job.Name,
			Description: job.Description.String,
			Unit:        job.Unit,
			Quantity:    job.Quantity,
			LaborCost:   job.LaborCost,
		})
	}

	response.Jobs = jobForResponse

	return response, nil
}

func (r *boqRepository) AddBOQJob(ctx context.Context, boqID uuid.UUID, req requests.BOQJobRequest) error {
	// Start transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check BOQ status
	var status string
	checkStatusQuery := `SELECT status FROM boq WHERE boq_id = $1`
	err = tx.GetContext(ctx, &status, checkStatusQuery, boqID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("boq not found")
		}
		return fmt.Errorf("failed to get BOQ status: %w", err)
	}

	if status != "draft" {
		return errors.New("can only add jobs to BOQ in draft status")
	}

	// Validate input
	if req.Quantity <= 0 || req.LaborCost <= 0 {
		return errors.New("quantity and labor cost must be positive numbers")
	}

	// Check if job already exists in BOQ
	var exists bool
	checkJobQuery := `
        SELECT EXISTS (
            SELECT 1 FROM boq_job 
            WHERE boq_id = $1 AND job_id = $2
        )`
	err = tx.GetContext(ctx, &exists, checkJobQuery, boqID, req.JobID)
	if err != nil {
		return fmt.Errorf("failed to check job existence: %w", err)
	}
	if exists {
		return errors.New("job already exists in this BOQ")
	}

	// Insert into boq_job
	insertBOQJobQuery := `
        INSERT INTO boq_job (
            boq_id, job_id, quantity, labor_cost
        ) VALUES (
            $1, $2, $3, $4
        )`

	_, err = tx.ExecContext(ctx, insertBOQJobQuery,
		boqID,
		req.JobID,
		req.Quantity,
		req.LaborCost,
	)
	if err != nil {
		return fmt.Errorf("failed to add job to BOQ: %w", err)
	}

	// Get all materials for the job
	materialQuery := `
        SELECT material_id, quantity 
        FROM job_material 
        WHERE job_id = $1`

	type JobMaterial struct {
		MaterialID string  `db:"material_id"`
		Quantity   float64 `db:"quantity"`
	}
	var materials []JobMaterial

	err = tx.SelectContext(ctx, &materials, materialQuery, req.JobID)
	if err != nil {
		return fmt.Errorf("failed to get job materials: %w", err)
	}

	// Get existing materials in BOQ with their estimated prices
	type ExistingMaterial struct {
		MaterialID     sql.NullString  `db:"material_id"`
		EstimatedPrice sql.NullFloat64 `db:"estimated_price"`
	}

	// แก้ไข query ให้ชัดเจนขึ้นว่าต้องการข้อมูลอะไร และจัดการกับ NULL
	existingMaterialsQuery := `
    SELECT DISTINCT 
        mpl.material_id, 
        mpl.estimated_price 
    FROM boq_job bj 
    INNER JOIN material_price_log mpl 
        ON mpl.boq_id = bj.boq_id
        AND mpl.job_id = bj.job_id 
    WHERE bj.boq_id = $1
    AND mpl.material_id IS NOT NULL`

	var existingMaterials []ExistingMaterial
	err = tx.SelectContext(ctx, &existingMaterials, existingMaterialsQuery, boqID)
	if err != nil {
		return fmt.Errorf("failed to get existing materials: %w", err)
	}

	// Create map for quick lookup of estimated prices
	estimatedPrices := make(map[string]sql.NullFloat64)
	for _, em := range existingMaterials {
		if em.MaterialID.Valid {
			estimatedPrices[em.MaterialID.String] = em.EstimatedPrice
		}
	}
	// Add material_price_log entries
	for _, material := range materials {
		insertPriceLogQuery := `
            INSERT INTO material_price_log (
                material_id, boq_id, job_id, quantity, estimated_price, updated_at
            ) VALUES (
                $1, $2, $3, $4, $5, CURRENT_TIMESTAMP
            )`

		estimatedPrice := estimatedPrices[material.MaterialID]

		_, err = tx.ExecContext(ctx, insertPriceLogQuery,
			material.MaterialID,
			boqID,
			req.JobID,
			material.Quantity,
			estimatedPrice,
		)
		if err != nil {
			return fmt.Errorf("failed to create material price log: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *boqRepository) UpdateBOQJob(ctx context.Context, boqID uuid.UUID, req requests.BOQJobRequest) error {
	jobID := req.JobID
	// Start transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check BOQ status
	var status string
	checkStatusQuery := `SELECT status FROM boq WHERE boq_id = $1`
	err = tx.GetContext(ctx, &status, checkStatusQuery, boqID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("boq not found")
		}
		return fmt.Errorf("failed to get BOQ status: %w", err)
	}

	if status != "draft" {
		return errors.New("can only update jobs in BOQ in draft status")
	}

	// Update BOQ job
	updateBOQJobQuery := `
		UPDATE boq_job
		SET quantity = $1, labor_cost = $2
		WHERE boq_id = $3 AND job_id = $4`

	_, err = tx.ExecContext(ctx, updateBOQJobQuery, req.Quantity, req.LaborCost, boqID, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job in BOQ: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *boqRepository) DeleteBOQJob(ctx context.Context, boqID uuid.UUID, jobID uuid.UUID) error {
	// Start transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check BOQ status
	var status string
	checkStatusQuery := `SELECT status FROM boq WHERE boq_id = $1`
	err = tx.GetContext(ctx, &status, checkStatusQuery, boqID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("boq not found")
		}
		return fmt.Errorf("failed to get BOQ status: %w", err)
	}

	if status != "draft" {
		return errors.New("can only delete jobs from BOQ in draft status")
	}

	// Delete related material price logs first (foreign key constraint)
	deleteMaterialPriceLogQuery := `
        DELETE FROM material_price_log 
        WHERE boq_id = $1 
        AND job_id = $2`

	_, err = tx.ExecContext(ctx, deleteMaterialPriceLogQuery, boqID, jobID)
	if err != nil {
		return fmt.Errorf("failed to delete material price logs: %w", err)
	}

	// Then delete the BOQ job
	deleteBOQJobQuery := `
        DELETE FROM boq_job 
        WHERE boq_id = $1 
        AND job_id = $2`

	result, err := tx.ExecContext(ctx, deleteBOQJobQuery, boqID, jobID)
	if err != nil {
		return fmt.Errorf("failed to delete job from BOQ: %w", err)
	}

	// Check if the BOQ job was actually deleted
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("job not found in BOQ")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *boqRepository) GetBOQGeneralCosts(ctx context.Context, boqID uuid.UUID) ([]models.BOQGeneralCost, error) {
	query := `
        SELECT b.boq_id, gc.type_name, gc.estimated_cost 
        FROM boq b 
        JOIN general_cost gc ON gc.boq_id = b.boq_id 
        JOIN "type" t ON t.type_name = gc.type_name 
        WHERE b.boq_id = $1`

	var costs []models.BOQGeneralCost
	err := r.db.SelectContext(ctx, &costs, query, boqID)
	if err != nil {
		return nil, err
	}

	return costs, nil
}
func (r *boqRepository) GetBOQDetails(ctx context.Context, projectID uuid.UUID) ([]models.BOQDetails, error) {
	query := `
        WITH MaterialTotals AS (
            SELECT 
                job_id, 
                boq_id, 
                COALESCE(SUM(COALESCE(estimated_price, 0) * COALESCE(quantity, 0)), 0) as total_material_price
            FROM material_price_log
            GROUP BY job_id, boq_id
        )
        SELECT 
            p.name, 
            p.address, 
			j.job_id,
            j.name as job_name, 
            j.description, 
            bj.quantity, 
            j.unit, 
            COALESCE(bj.labor_cost, 0) as labor_cost,
            mt.total_material_price as estimated_price,
            (mt.total_material_price * bj.quantity) as total_estimated_price,
            (COALESCE(bj.labor_cost, 0) * bj.quantity) as total_labour_cost,
            ((mt.total_material_price * bj.quantity) + (COALESCE(bj.labor_cost, 0) * bj.quantity)) as total
        FROM project p 
        JOIN boq b ON b.project_id = p.project_id 
        LEFT JOIN client c ON c.client_id = p.project_id
        JOIN boq_job bj ON bj.boq_id = b.boq_id 
        JOIN job j ON j.job_id = bj.job_id 
        LEFT JOIN MaterialTotals mt ON mt.job_id = bj.job_id AND mt.boq_id = bj.boq_id 
        WHERE p.project_id = $1 
        GROUP BY 
            p.name, p.address, j.job_id, j.name, j.description, 
            bj.quantity, j.unit, bj.labor_cost, mt.total_material_price`

	var details []models.BOQDetails
	err := r.db.SelectContext(ctx, &details, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get BOQ details: %w", err)
	}

	return details, nil
}

func (r *boqRepository) GetBOQMaterialDetails(ctx context.Context, projectID uuid.UUID) ([]models.BOQMaterialDetails, error) {
	query := `
        SELECT 
		    j.job_id,
            j.name, 
            m.name as material_name,
            mpl.quantity, 
            m.unit, 
            mpl.estimated_price, 
            COALESCE(mpl.quantity, 0) * COALESCE(mpl.estimated_price, 0) as total
        FROM project p 
        JOIN boq b ON b.project_id = p.project_id 
        LEFT JOIN client c ON c.client_id = p.project_id 
        JOIN boq_job bj ON bj.boq_id = b.boq_id 
        JOIN job j ON j.job_id = bj.job_id 
        LEFT JOIN material_price_log mpl ON mpl.job_id = bj.job_id AND mpl.boq_id = bj.boq_id 
        JOIN material m ON m.material_id = mpl.material_id 
        WHERE p.project_id = $1`

	var details []models.BOQMaterialDetails
	err := r.db.SelectContext(ctx, &details, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get material details: %w", err)
	}

	return details, nil
}
