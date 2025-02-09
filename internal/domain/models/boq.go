package models

import (
	"database/sql"

	"github.com/google/uuid"
)

type BOQStatus string

const (
	BOQStatusDraft    BOQStatus = "draft"
	BOQStatusApproved BOQStatus = "approved"
)

type BOQ struct {
	BOQID              uuid.UUID       `db:"boq_id"`
	ProjectID          uuid.UUID       `db:"project_id"`
	Status             BOQStatus       `db:"status"`
	SellingGeneralCost sql.NullFloat64 `db:"selling_general_cost"`
}

type BOQDetails struct {
	ProjectName         string          `db:"name"`
	ProjectAddress      sql.NullString  `db:"address"`
	JobID               uuid.UUID       `db:"job_id"`
	JobName             string          `db:"job_name"`
	Description         sql.NullString  `db:"description"`
	Quantity            int             `db:"quantity"`
	Unit                string          `db:"unit"`
	LaborCost           float64         `db:"labor_cost"`
	EstimatedPrice      sql.NullFloat64 `db:"estimated_price"`
	TotalEstimatedPrice sql.NullFloat64 `db:"total_estimated_price"`
	TotalLaborCost      float64         `db:"total_labour_cost"`
	Total               sql.NullFloat64 `db:"total"`
}

type BOQMaterialDetails struct {
	JobID          uuid.UUID       `db:"job_id"`
	JobName        string          `db:"name"`
	MaterialName   string          `db:"material_name"`
	Quantity       sql.NullFloat64 `db:"quantity"` // Changed to handle NULL
	Unit           string          `db:"unit"`
	EstimatedPrice sql.NullFloat64 `db:"estimated_price"` // Changed to handle NULL
	Total          sql.NullFloat64 `db:"total"`           // Changed to handle NULL
}

type BOQGeneralCost struct {
	BOQID         uuid.UUID `db:"boq_id"`
	TypeName      string    `db:"type_name"`
	EstimatedCost float64   `db:"estimated_cost"`
}
