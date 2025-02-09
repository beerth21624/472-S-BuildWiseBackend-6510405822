package models

import (
	"database/sql"

	"github.com/google/uuid"
)

type Job struct {
	JobID       uuid.UUID      `db:"job_id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Unit        string         `db:"unit"`
}
type JobSummary struct {
	QuotationID        uuid.UUID    `db:"quotation_id"`
	QuotationStatus    string       `db:"status"`
	ValidDate          sql.NullTime `db:"valid_date"` // Changed to sql.NullTime
	TaxPercentage      float64      `db:"tax_percentage"`
	JobName            string       `db:"name"`
	Unit               string       `db:"unit"`
	Quantity           int          `db:"quantity"`
	LaborCost          float64      `db:"labor_cost"`
	MaterialCost       float64      `db:"total_material_price"`
	OverallCost        float64      `db:"overall_cost"`
	SellingPrice       float64      `db:"selling_price"`
	EstimatedProfit    float64      `db:"estimated_profit"`
	ActualOverallCost  float64      `db:"overall_actual_price"`
	ActualProfit       float64      `db:"job_profit"`
	TotalProfit        float64      `db:"total_profit"`
	SellingGeneralCost float64      `db:"selling_general_cost"`
}
