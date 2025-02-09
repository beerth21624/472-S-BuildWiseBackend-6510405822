package requests

import "github.com/google/uuid"

type CreateJobRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Unit        string `json:"unit" validate:"required"`
}

type UpdateJobRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Unit        string `json:"unit" validate:"required"`
}

type AddJobMaterialRequest struct {
	Materials []JobMaterialItem `json:"materials" validate:"required,dive"`
}

type JobMaterialItem struct {
	MaterialID string  `json:"material_id" validate:"required"`
	Quantity   float64 `json:"quantity" validate:"required,gt=0"`
}

type DeleteJobMaterialRequest struct {
	JobID      uuid.UUID `json:"job_id" validate:"required"`
	MaterialID string    `json:"material_id" validate:"required"`
}

type UpdateJobMaterialQuantityRequest struct {
	JobID      uuid.UUID `json:"job_id" validate:"required"`
	MaterialID string    `json:"material_id" validate:"required"`
	Quantity   float64   `json:"quantity" validate:"required,gt=0"`
}
