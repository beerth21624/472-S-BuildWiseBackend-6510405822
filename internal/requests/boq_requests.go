package requests

import (
	"github.com/google/uuid"
)

type CreateBOQRequest struct {
	ProjectID uuid.UUID `json:"project_id" validate:"required"`
}

type UpdateBOQRequest struct {
	Status             string  `json:"status" validate:"required,oneof=draft approved"`
	SellingGeneralCost float64 `json:"selling_general_cost" validate:"required"`
}
type BOQJobRequest struct {
	JobID     uuid.UUID `json:"job_id" validate:"required"`
	Quantity  float64   `json:"quantity" validate:"required,gt=0"`
	LaborCost float64   `json:"labor_cost" validate:"required,gt=0"`
}
