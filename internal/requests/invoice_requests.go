package requests

import (
	"github.com/google/uuid"
)

type CreateInvoiceRequest struct {
	ContractID  uuid.UUID `json:"contract_id" validate:"required"`
	PaymentTerm string    `json:"payment_term"`
}

type DeleteInvoiceRequest struct {
	InvoiceID uuid.UUID `json:"invoice_id" validate:"required"`
}

type UpdateInvoiceStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending processing paid canceled"`
}

type UpdateInvoicePaidRequest struct {
	PaidDate string `json:"paid_date" validate:"required,datetime=2006-01-02"`
}

type UpdateInvoiceRequest struct {
	InvoiceDate    *string  `json:"invoice_date" validate:"omitempty,datetime=2006-01-02"`
	PaymentDueDate *string  `json:"payment_due_date" validate:"omitempty,datetime=2006-01-02"`
	PaidDate       *string  `json:"paid_date" validate:"omitempty,datetime=2006-01-02"`
	PaymentTerm    *string  `json:"payment_term"`
	Remarks        *string  `json:"remarks"`
	Retention      *float64 `json:"retention"`
}
