package requests

import (
	"github.com/google/uuid"
)

type CreateGeneralCostRequest struct {
	BOQID      uuid.UUID `json:"boq_id" validate:"required"`
	TypeName   string    `json:"type_name" validate:"required"`
	ActualCost float64   `json:"actual_cost" validate:"required,gte=0"`
}

type UpdateGeneralCostRequest struct {
	EstimatedCost float64 `json:"estimated_cost" validate:"required,gte=0"`
}

type UpdateActualGeneralCostRequest struct {
	ActualCost float64 `json:"actual_cost" validate:"required,gte=0"`
}
