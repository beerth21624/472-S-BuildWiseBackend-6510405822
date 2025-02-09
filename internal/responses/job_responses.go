package responses

import (
	"github.com/google/uuid"
)

type JobResponse struct {
	JobID       uuid.UUID `json:"job_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Unit        string    `json:"unit"`
	Quantity    float64   `json:"quantity"`
	LaborCost   float64   `json:"labor_cost"`
}

type JobMaterialResponse struct {
	JobID       uuid.UUID         `json:"job_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Unit        string            `json:"unit"`
	Materials   []JobMaterialItem `json:"materials"`
}
type JobUsage struct {
	IsUsed   bool           `json:"is_used"`
	Projects []ProjectUsage `json:"projects,omitempty"`
}

type ProjectUsage struct {
	ProjectID   uuid.UUID `db:"project_id"`
	ProjectName string    `db:"project_name"`
	BOQID       uuid.UUID `db:"boq_id"`
	BOQStatus   string    `db:"boq_status"`
}
type JobMaterialItem struct {
	MaterialID string  `json:"material_id" db:"material_id"`
	Name       string  `json:"name" db:"name"`
	Unit       string  `json:"unit" db:"unit"`
	Quantity   float64 `json:"quantity" db:"quantity"`
}

type JobListResponse struct {
	Jobs []JobResponse `json:"jobs"`
}

type PaginationResponse struct {
	CurrentPage  int `json:"current_page"`
	PageSize     int `json:"page_size"`
	TotalPages   int `json:"total_pages"`
	TotalRecords int `json:"total_records"`
}

// ApiResponse represents a standard API response structure
type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
