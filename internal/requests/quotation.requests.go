package requests

import "github.com/google/uuid"

type UpdateProjectSellingPriceRequest struct {
	ProjectID          uuid.UUID         `json:"project_id" validate:"required"`
	TaxPercentage      float64           `json:"tax_percentage" validate:"required,gt=0"`
	SellingGeneralCost float64           `json:"selling_general_cost" validate:"required,gt=0"`
	JobSellingPrices   []JobSellingPrice `json:"job_selling_prices" validate:"required,dive"`
}

type JobSellingPrice struct {
	JobID        uuid.UUID `json:"job_id" validate:"required"`
	SellingPrice float64   `json:"selling_price" validate:"required,gt=0"`
}
