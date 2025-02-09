package responses

import "github.com/google/uuid"

type MaterialResponse struct {
	MaterialID string `json:"material_id"`
	Name       string `json:"name"`
	Unit       string `json:"unit"`
}

type MaterialListResponse struct {
	Materials []MaterialResponse `json:"materials"`
}

type MaterialPriceListResponse struct {
	Materials []MaterialPriceDetail `json:"materials"`
}

type MaterialPriceDetail struct {
	MaterialID     string  `json:"material_id"`
	Name           string  `json:"name"`
	TotalQuantity  float64 `json:"total_quantity"`
	Unit           string  `json:"unit"`
	EstimatedPrice float64 `json:"estimated_price"`
	AvgActualPrice float64 `json:"avg_actual_price"`
	ActualPrice    float64 `json:"actual_price"`
	SupplierID     string  `json:"supplier_id"`
	SupplierName   string  `json:"supplier_name"`
}

type MaterialActualPriceResponse struct {
	MaterialID  string    `json:"material_id"`
	ActualPrice float64   `json:"actual_price"`
	SupplierID  uuid.UUID `json:"supplier_id"`
}
