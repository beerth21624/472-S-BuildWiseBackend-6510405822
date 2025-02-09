package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type QuotationStatus string

const (
	QuotationStatusDraft    QuotationStatus = "draft"
	QuotationStatusApproved QuotationStatus = "approved"
)

type Quotation struct {
	QuotationID   uuid.UUID       `db:"quotation_id"`
	ProjectID     uuid.UUID       `db:"project_id"`
	ValidDate     sql.NullTime    `db:"valid_date"`
	Status        QuotationStatus `db:"status"`
	FinalAmount   sql.NullFloat64 `db:"final_amount"`
	TaxPercentage sql.NullFloat64 `db:"tax_percentage"`
}

type QuotationJob struct {
	QuotationID        uuid.UUID       `db:"quotation_id"`
	Status             string          `db:"status"`
	ValidDate          time.Time       `db:"valid_date"`
	TaxPercentage      sql.NullFloat64 `db:"tax_percentage"`
	SellingGeneralCost sql.NullFloat64 `db:"selling_general_cost"`
	JobID              uuid.UUID       `db:"job_id"`
	JobName            string          `db:"name"`
	Unit               string          `db:"unit"`
	Quantity           float64         `db:"quantity"`
	LaborCost          float64         `db:"labor_cost"`
	TotalMaterialPrice sql.NullFloat64 `db:"total_material_price"`
	OverallCost        sql.NullFloat64 `db:"overall_cost"`
	SellingPrice       sql.NullFloat64 `db:"selling_price"`
	Total              sql.NullFloat64 `db:"total"`
	TotalSellingPrice  sql.NullFloat64 `db:"total_selling_price"`
}

type QuotationGeneralCost struct {
	BoqID              uuid.UUID  `db:"boq_id"`
	SellingGeneralCost float64    `db:"selling_general_cost"`
	TaxPercentage      float64    `db:"tax_percentage"`
	GID                *uuid.UUID `db:"g_id"`
	TypeName           *string    `db:"type_name"`
	EstimatedCost      *float64   `db:"estimated_cost"`
}

type QuotationExportData struct {
	ProjectID     uuid.UUID       `db:"project_id"`
	ProjectName   string          `db:"name"`
	Description   string          `db:"description"`
	Address       json.RawMessage `db:"address"`
	ClientName    string          `db:"client_name"`
	ClientAddress json.RawMessage `db:"client_address"`
	ClientEmail   string          `db:"client_email"`
	ClientTel     string          `db:"client_tel"`
	ClientTaxID   string          `db:"client_tax_id"`
	QuotationID   uuid.UUID       `db:"quotation_id"`
	ValidDate     time.Time       `db:"valid_date"`
	TaxPercentage float64         `db:"tax_percentage"`
	FinalAmount   sql.NullFloat64 `db:"final_amount"`
	Status        string          `db:"status"`
	JobDetails    []JobDetail     `db:"-"`
}

type JobDetail struct {
	Name         string          `db:"name"`
	Unit         string          `db:"unit"`
	Quantity     float64         `db:"quantity"`
	SellingPrice sql.NullFloat64 `db:"selling_price"` // Changed to handle NULL
	Amount       sql.NullFloat64 `db:"amount"`        // Changed to handle NULL
}
