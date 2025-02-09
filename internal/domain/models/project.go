package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ProjectStatus string

const (
	ProjectStatusPlanning   ProjectStatus = "planning"
	ProjectStatusInProgress ProjectStatus = "in_progress"
	ProjectStatusCompleted  ProjectStatus = "completed"
	ProjectStatusCancelled  ProjectStatus = "cancelled"
)

type Project struct {
	ProjectID   uuid.UUID       `db:"project_id"`
	Name        string          `db:"name"`
	Description string          `db:"description"`
	Address     json.RawMessage `db:"address"`
	Status      ProjectStatus   `db:"status"`
	ClientID    uuid.UUID       `db:"client_id"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   sql.NullTime    `db:"updated_at"`
}

type ProjectStatusCheck struct {
	ProjectStatus   string         `db:"project_status"`
	BOQStatus       sql.NullString `db:"boq_status"`
	QuotationStatus sql.NullString `db:"quotation_status"`
}

type ProjectOverview struct {
	QuotationID       uuid.UUID       `db:"quotation_id"`
	BOQID             uuid.UUID       `db:"boq_id"`
	TotalOverallCost  sql.NullFloat64 `db:"total_overall_cost"`
	TotalSellingPrice sql.NullFloat64 `db:"total_selling_price"`
	TaxPercentage     sql.NullFloat64 `db:"tax_percentage"`
	TotalActualCost   sql.NullFloat64 `db:"total_actual_cost"`
}

type ProjectSummary struct {
	ProjectOverview
	Jobs []JobSummary
}
