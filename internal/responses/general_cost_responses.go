package responses

import (
	"github.com/google/uuid"
)

type GeneralCostResponse struct {
	GID           uuid.UUID `json:"g_id"`
	BOQID         uuid.UUID `json:"boq_id"`
	TypeName      string    `json:"type_name"`
	ActualCost    float64   `json:"actual_cost"`
	EstimatedCost float64   `json:"estimated_cost"`
}

type GeneralCostListResponse struct {
	GeneralCosts []GeneralCostResponse `json:"general_costs"`
}

type GeneralCostUpdateResponse struct {
	Message string    `json:"message"`
	GID     uuid.UUID `json:"g_id"`
}
