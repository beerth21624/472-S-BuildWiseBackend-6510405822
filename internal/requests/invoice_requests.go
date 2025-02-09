package requests

import (
	"github.com/google/uuid"
)

type CreateInvoiceRequest struct {
	FileURL string `json:"file_url" validate:"required,url"`
}

type DeleteInvoiceRequest struct {
	InvoiceID uuid.UUID `json:"invoice_id" validate:"required"`
}
