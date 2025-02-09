package requests

import "github.com/google/uuid"

type CreateMaterialRequest struct {
	Name string `json:"name" validate:"required"`
	Unit string `json:"unit" validate:"required"`
}

type UpdateMaterialRequest struct {
	Name string `json:"name" validate:"required"`
	Unit string `json:"unit" validate:"required"`
}

type UpdateMaterialEstimatedPriceRequest struct {
	MaterialID     string  `json:"material_id" validate:"required"`
	EstimatedPrice float64 `json:"estimated_price" validate:"required,gt=0"`
}

type UpdateMaterialActualPriceRequest struct {
	MaterialID  string    `json:"material_id" validate:"required"`
	ActualPrice float64   `json:"actual_price" validate:"required,gt=0"`
	SupplierID  uuid.UUID `json:"supplier_id" validate:"required"`
}
