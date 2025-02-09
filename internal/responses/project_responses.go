package responses

import (
	"boonkosang/internal/domain/models"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ProjectResponse struct {
	ID          uuid.UUID            `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Address     json.RawMessage      `json:"address"`
	Status      models.ProjectStatus `json:"status"`
	ClientID    uuid.UUID            `json:"client_id"`
	Client      *ClientResponse      `json:"client,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type ProjectListResponse struct {
	Projects []ProjectResponse `json:"projects"`
	Total    int64             `json:"total"`
}

type ProjectOverviewResponse struct {
	QuotationID       string  `json:"quotation_id"`
	BOQID             string  `json:"boq_id"`
	TotalOverallCost  float64 `json:"total_overall_cost"`
	TotalSellingPrice float64 `json:"total_selling_price"`
	TotalActualCost   float64 `json:"total_actual_cost"`
	TaxAmount         float64 `json:"tax_amount"`
	TotalWithTax      float64 `json:"total_with_tax"`
	EstimatedProfit   float64 `json:"estimated_profit"`
	EstimatedMargin   float64 `json:"estimated_margin_percentage"`
	ActualProfit      float64 `json:"actual_profit"`
	ActualMargin      float64 `json:"actual_margin_percentage"`
}

type ProjectSummaryResponse struct {
	ProjectID   string                  `json:"project_id"`
	ProjectName string                  `json:"project_name"`
	Overview    ProjectOverviewResponse `json:"overview"`
	Jobs        []JobSummaryResponse    `json:"jobs"`
	TotalStats  TotalStatsResponse      `json:"total_stats"`
}

type JobSummaryResponse struct {
	JobName           string  `json:"job_name"`
	Unit              string  `json:"unit"`
	Quantity          int     `json:"quantity"`
	ValidDate         *string `json:"valid_date,omitempty"` // Changed to pointer to string
	LaborCost         float64 `json:"labor_cost"`
	MaterialCost      float64 `json:"material_cost"`
	OverallCost       float64 `json:"overall_cost"`
	SellingPrice      float64 `json:"selling_price"`
	EstimatedProfit   float64 `json:"estimated_profit"`
	EstimatedMargin   float64 `json:"estimated_margin_percentage"`
	ActualOverallCost float64 `json:"actual_overall_cost"`
	ActualProfit      float64 `json:"actual_profit"`
	ActualMargin      float64 `json:"actual_margin_percentage"`
	TotalProfit       float64 `json:"total_profit"`
	QuotationStatus   string  `json:"quotation_status"`
	TaxPercentage     float64 `json:"tax_percentage"`
}
type TotalStatsResponse struct {
	TotalEstimatedCost   float64 `json:"total_estimated_cost"`
	TotalActualCost      float64 `json:"total_actual_cost"`
	TotalSellingPrice    float64 `json:"total_selling_price"`
	TotalEstimatedProfit float64 `json:"total_estimated_profit"`
	TotalActualProfit    float64 `json:"total_actual_profit"`
	EstimatedMargin      float64 `json:"estimated_margin_percentage"`
	ActualMargin         float64 `json:"actual_margin_percentage"`
	CostVariance         float64 `json:"cost_variance"`
	CostVariancePercent  float64 `json:"cost_variance_percentage"`
}
