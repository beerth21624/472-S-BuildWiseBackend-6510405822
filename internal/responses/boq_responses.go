package responses

import (
	"boonkosang/internal/domain/models"
	"encoding/json"

	"github.com/google/uuid"
)

type BOQResponse struct {
	ID                 uuid.UUID        `json:"id"`
	ProjectID          uuid.UUID        `json:"project_id"`
	Status             models.BOQStatus `json:"status"`
	SellingGeneralCost float64          `json:"selling_general_cost"`
	Jobs               []JobResponse    `json:"jobs"`
}

type BOQListResponse struct {
	BOQs  []BOQResponse `json:"boqs"`
	Total int64         `json:"total"`
}

type BOQSummaryResponse struct {
	ProjectInfo    ProjectInfo      `json:"project_info"`
	GeneralCosts   []GeneralCostDTO `json:"general_costs"`
	Details        []BOQDetailDTO   `json:"jobs"`
	SummaryMetrics SummaryMetrics   `json:"summary_metrics"`
}

type ProjectInfo struct {
	ProjectName    string          `json:"project_name"`
	ProjectAddress json.RawMessage `json:"project_address"`
}

type GeneralCostDTO struct {
	TypeName      string  `json:"type_name"`
	EstimatedCost float64 `json:"estimated_cost"`
}

type BOQDetailDTO struct {
	JobID               uuid.UUID     `json:"job_id"`
	JobName             string        `json:"job_name"`
	Description         string        `json:"description"`
	Quantity            int           `json:"quantity"`
	Unit                string        `json:"unit"`
	LaborCost           float64       `json:"labor_cost"`
	EstimatedPrice      float64       `json:"estimated_price"`
	TotalEstimatedPrice float64       `json:"total_estimated_price"`
	TotalLaborCost      float64       `json:"total_labor_cost"`
	Total               float64       `json:"total"`
	Materials           []MaterialDTO `json:"materials"`
}

type MaterialDTO struct {
	JobID          uuid.UUID `json:"job_id"`
	JobName        string    `json:"job_name"`
	MaterialName   string    `json:"material_name"`
	Quantity       float64   `json:"quantity"`
	Unit           string    `json:"unit"`
	EstimatedPrice float64   `json:"estimated_price"`
	Total          float64   `json:"total"`
}

type SummaryMetrics struct {
	TotalGeneralCost    float64 `json:"total_general_cost"`
	TotalMaterialCost   float64 `json:"total_material_cost"`
	TotalLaborCost      float64 `json:"total_labor_cost"`
	TotalEstimatedPrice float64 `json:"total_estimated_price"`
	TotalAmount         float64 `json:"total_amount"`
	GrandTotal          float64 `json:"grand_total"`
}
