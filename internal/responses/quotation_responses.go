package responses

import (
	"boonkosang/internal/domain/models"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type QuotationResponse struct {
	QuotationID        uuid.UUID            `json:"quotation_id"`
	Status             string               `json:"status"`
	ValidDate          time.Time            `json:"valid_date"`
	TaxPercentage      float64              `json:"tax_percentage"`
	SellingGeneralCost float64              `json:"selling_general_cost"`
	Jobs               []QuotationJobDetail `json:"jobs"`
	Costs              []GeneralCostDetail  `json:"general_costs"`
}

type QuotationJobDetail struct {
	ID uuid.UUID `json:"id"`

	Name               string  `json:"name"`
	Unit               string  `json:"unit"`
	Quantity           float64 `json:"quantity"`
	LaborCost          float64 `json:"labor_cost"`
	SellingPrice       float64 `json:"selling_price"`
	TotalMaterialPrice float64 `json:"total_material_price"`
	Total              float64 `json:"total"`
	OverallCost        float64 `json:"overall_cost"`
	TotalSellingPrice  float64 `json:"total_selling_price"`
}

type GeneralCostDetail struct {
	TypeName      string  `json:"type_name"`
	EstimatedCost float64 `json:"estimated_cost"`
}

type QuotationExportData struct {
	ProjectID   uuid.UUID       `json:"project_id" db:"project_id"`
	ProjectName string          `json:"project_name" db:"name"`
	Description string          `json:"description" db:"description"`
	Address     json.RawMessage `json:"address" db:"address"`

	ClientName    string          `json:"client_name" db:"client_name"`
	ClientAddress json.RawMessage `json:"client_address" db:"client_address"`
	ClientEmail   string          `json:"client_email" db:"client_email"`
	ClientTel     string          `json:"client_tel" db:"client_tel"`
	ClientTaxID   string          `json:"client_tax_id" db:"client_tax_id"`

	QuotationID   uuid.UUID              `json:"quotation_id" db:"quotation_id"`
	ValidDate     time.Time              `json:"valid_date" db:"valid_date"`
	TaxPercentage float64                `json:"tax_percentage" db:"tax_percentage"`
	FinalAmount   sql.NullFloat64        `json:"-" db:"final_amount"`
	Status        models.QuotationStatus `json:"status" db:"status"`

	SubTotal  float64 `json:"sub_total"`
	TaxAmount float64 `json:"tax_amount"`

	JobDetails []JobDetail `json:"jobs"`

	SellingGeneralCost   float64  `json:"selling_general_cost"`
	FormattedFinalAmount *float64 `json:"final_amount"`
}

func (q *QuotationExportData) FormatFinalAmount() {
	if q.FinalAmount.Valid {
		q.FormattedFinalAmount = &q.FinalAmount.Float64
	} else {
		q.FormattedFinalAmount = nil
	}
}

type JobDetail struct {
	Name         string          `json:"name" db:"name"`
	Description  string          `json:"description" db:"description"`
	Unit         string          `json:"unit" db:"unit"`
	Quantity     float64         `json:"quantity" db:"quantity"`
	SellingPrice sql.NullFloat64 `json:"-" db:"selling_price"`
	Amount       sql.NullFloat64 `json:"-" db:"amount"`

	FormattedSellingPrice *float64 `json:"selling_price"`
	FormattedAmount       *float64 `json:"amount"`
}

func (j *JobDetail) FormatSellingPrice() {
	if j.SellingPrice.Valid {
		j.FormattedSellingPrice = &j.SellingPrice.Float64
	} else {
		j.FormattedSellingPrice = nil
	}
}

func (j *JobDetail) FormatAmount() {
	if j.Amount.Valid {
		j.FormattedAmount = &j.Amount.Float64
	} else {
		j.FormattedAmount = nil
	}
}
