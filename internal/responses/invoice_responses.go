package responses

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceResponse struct {
	InvoiceID uuid.UUID `json:"invoice_id"`
	ProjectID uuid.UUID `json:"project_id"`
	FileURL   string    `json:"file_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type InvoiceListResponse struct {
	Invoices []InvoiceResponse `json:"invoices"`
}
