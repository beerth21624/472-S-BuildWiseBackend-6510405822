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

type materialRepository struct {
	db *sqlx.DB
}

func NewMaterialRepository(db *sqlx.DB) repositories.MaterialRepository {
	return &materialRepository{
		db: db,
	}
}

func (r *materialRepository) Create(ctx context.Context, req requests.CreateMaterialRequest) (*models.Material, error) {
	material := &models.Material{
		MaterialID: uuid.New().String(),
		Name:       req.Name,
		Unit:       req.Unit,
	}

	query := `
        INSERT INTO Material (
            material_id, name, unit
        ) VALUES (
            :material_id, :name, :unit
        ) RETURNING *`

	rows, err := r.db.NamedQueryContext(ctx, query, material)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.New("material ID already exists")
		}
		return nil, fmt.Errorf("failed to create material: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.StructScan(material)
		if err != nil {
			return nil, fmt.Errorf("failed to scan material: %w", err)
		}
		return material, nil
	}
	return nil, errors.New("failed to create material: no rows returned")
}

func (r *materialRepository) Update(ctx context.Context, materialID string, req requests.UpdateMaterialRequest) error {
	query := `
        UPDATE Material SET 
            name = :name,
            unit = :unit
        WHERE material_id = :material_id`

	params := map[string]interface{}{
		"material_id": materialID,
		"name":        req.Name,
		"unit":        req.Unit,
	}

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to update material: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("material not found")
	}

	return nil
}

func (r *materialRepository) Delete(ctx context.Context, materialID string) error {
	// Start transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check material usage in projects
	type ProjectUsage struct {
		ProjectID   uuid.UUID `db:"project_id"`
		ProjectName string    `db:"name"`
		Status      string    `db:"status"`
	}

	checkUsageQuery := `
       SELECT DISTINCT 
           p.project_id,
           p.name,
           b.status
       FROM job_material jm 
       JOIN boq_job bj ON bj.job_id = jm.job_id 
       JOIN boq b ON b.boq_id = bj.boq_id 
       JOIN project p ON p.project_id = b.project_id 
       WHERE jm.material_id = $1`

	var usages []ProjectUsage
	err = tx.SelectContext(ctx, &usages, checkUsageQuery, materialID)
	if err != nil {
		return fmt.Errorf("failed to check material usage: %w", err)
	}

	// If material is being used, return error with project names
	if len(usages) > 0 {
		var projectNames []string
		for _, usage := range usages {
			projectNames = append(projectNames, usage.ProjectName)
		}
		return fmt.Errorf("material is used in following projects: %s", strings.Join(projectNames, ", "))
	}

	// Delete from material_price_log first
	deletePriceLogQuery := `
       DELETE FROM material_price_log 
       WHERE material_id = $1`

	_, err = tx.ExecContext(ctx, deletePriceLogQuery, materialID)
	if err != nil {
		return fmt.Errorf("failed to delete material price logs: %w", err)
	}

	// Delete from job_material
	deleteJobMaterialQuery := `
       DELETE FROM job_material 
       WHERE material_id = $1`

	_, err = tx.ExecContext(ctx, deleteJobMaterialQuery, materialID)
	if err != nil {
		return fmt.Errorf("failed to delete job materials: %w", err)
	}

	// Finally delete the material
	deleteMaterialQuery := `
       DELETE FROM Material 
       WHERE material_id = $1`

	result, err := tx.ExecContext(ctx, deleteMaterialQuery, materialID)
	if err != nil {
		return fmt.Errorf("failed to delete material: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("material not found")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *materialRepository) GetByID(ctx context.Context, materialID string) (*models.Material, error) {
	material := &models.Material{}
	query := `SELECT * FROM Material WHERE material_id = $1`

	err := r.db.GetContext(ctx, material, query, materialID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("material not found")
		}
		return nil, fmt.Errorf("failed to get material: %w", err)
	}

	return material, nil
}

func (r *materialRepository) List(ctx context.Context) ([]models.Material, error) {
	var materials []models.Material
	var args []interface{}

	query := `
		SELECT * FROM Material
	   `

	err := r.db.SelectContext(ctx, &materials, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list materials: %w", err)
	}

	return materials, nil
}

func (r *materialRepository) GetMaterialPricesByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.MaterialPriceInfo, error) {
	query := `
        WITH LatestMaterialPrices AS (
            SELECT 
                mpl.material_id, 
                mpl.boq_id,
                mpl.actual_price, 
                mpl.updated_at,
                ROW_NUMBER() OVER (PARTITION BY mpl.material_id, mpl.boq_id ORDER BY mpl.updated_at DESC) AS row_num
            FROM material_price_log mpl
        ),
        FilteredPrices AS (
            SELECT 
                material_id,
                boq_id, 
                actual_price 
            FROM LatestMaterialPrices 
            WHERE row_num <= 3
            GROUP BY boq_id, material_id, actual_price
        ),
        FinalAvg AS (
            SELECT  
                material_id, 
                AVG(actual_price) AS avg_actual_price 
            FROM FilteredPrices
            GROUP BY material_id
        )
        SELECT 
            m.material_id, 
            m.name, 
            SUM(mpl.quantity) * SUM(bj.quantity) as qty_all_material_in_all_job,
            m.unit, 
            mpl.estimated_price,
            fa.avg_actual_price,
            mpl.actual_price,
			s.supplier_id,
            s.name as supplier_name
        FROM project p 
        JOIN boq b ON b.project_id = p.project_id 
        JOIN material_price_log mpl ON mpl.boq_id = b.boq_id
        JOIN job j ON j.job_id = mpl.job_id 
        JOIN material m ON m.material_id = mpl.material_id 
        JOIN FinalAvg fa ON fa.material_id = m.material_id
        JOIN boq_job bj ON bj.boq_id = mpl.boq_id AND bj.job_id = mpl.job_id 
        LEFT JOIN supplier s ON s.supplier_id = mpl.supplier_id
        WHERE p.project_id = $1
        GROUP BY 
            mpl.material_id, 
            m.material_id, 
            m.name, 
            mpl.estimated_price,
            mpl.actual_price, 
            fa.avg_actual_price, 
			s.supplier_id,
            s.name`

	var materials []models.MaterialPriceInfo
	err := r.db.SelectContext(ctx, &materials, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get material prices: %w", err)
	}

	return materials, nil
}

func (r *materialRepository) UpdateEstimatedPrices(ctx context.Context, boqID uuid.UUID, materialID string, estimatedPrice float64) error {
	query := `
        UPDATE material_price_log 
        SET estimated_price = $1
        WHERE material_id = $2 AND boq_id = $3`

	result, err := r.db.ExecContext(ctx, query, estimatedPrice, materialID, boqID)
	if err != nil {
		return fmt.Errorf("failed to update estimated prices: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("no material price records found to update")
	}

	return nil
}

func (r *materialRepository) GetBOQStatus(ctx context.Context, boqID uuid.UUID) (string, error) {
	var status string
	query := `SELECT status FROM boq WHERE boq_id = $1`

	err := r.db.GetContext(ctx, &status, query, boqID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("BOQ not found")
		}
		return "", fmt.Errorf("failed to get BOQ status: %w", err)
	}

	return status, nil
}

func (r *materialRepository) UpdateActualPrice(ctx context.Context, boqID uuid.UUID, req requests.UpdateMaterialActualPriceRequest) error {
	query := `
        UPDATE material_price_log 
        SET actual_price = :actual_price, 
            supplier_id = :supplier_id,
            updated_at = CURRENT_TIMESTAMP
        WHERE material_id = :material_id 
        AND boq_id = :boq_id`

	params := map[string]interface{}{
		"material_id":  req.MaterialID,
		"boq_id":       boqID,
		"actual_price": req.ActualPrice,
		"supplier_id":  req.SupplierID,
	}

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to update actual price: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("no material price records found to update")
	}

	return nil
}

func (r *materialRepository) GetProjectStatus(ctx context.Context, BOQId uuid.UUID) (string, error) {

	var projectID uuid.UUID
	queryProjectId := `SELECT project_id FROM boq WHERE boq_id = $1`

	err := r.db.GetContext(ctx, &projectID, queryProjectId, BOQId)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("BOQ not found")
		}
		return "", fmt.Errorf("failed to get project ID: %w", err)
	}

	var status string
	query := `SELECT status FROM project WHERE project_id = $1`

	err = r.db.GetContext(ctx, &status, query, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("project not found")
		}
		return "", fmt.Errorf("failed to get project status: %w", err)
	}

	return status, nil
}

func (r *materialRepository) GetQuotationStatus(ctx context.Context, BOQId uuid.UUID) (string, error) {
	var projectID uuid.UUID
	queryProjectId := `SELECT project_id FROM boq WHERE boq_id = $1`

	err := r.db.GetContext(ctx, &projectID, queryProjectId, BOQId)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("BOQ not found")
		}
		return "", fmt.Errorf("failed to get project ID: %w", err)
	}
	var status string
	query := `
        SELECT status 
        FROM quotation 
        WHERE project_id = $1 
        LIMIT 1`

	err = r.db.GetContext(ctx, &status, query, projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("quotation not found")
		}
		return "", fmt.Errorf("failed to get quotation status: %w", err)
	}

	return status, nil
}
