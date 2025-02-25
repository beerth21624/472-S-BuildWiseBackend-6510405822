package responses

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceResponse struct {
	InvoiceID      uuid.UUID      `json:"invoice_id"`
	ProjectID      uuid.UUID      `json:"project_id"`
	PeriodID       uuid.UUID      `json:"period_id"`
	InvoiceDate    time.Time      `json:"invoice_date"`
	PaymentDueDate time.Time      `json:"payment_due_date"`
	PaidDate       time.Time      `json:"paid_date"`
	PaymentTerm    string         `json:"payment_term"`
	Remarks        string         `json:"remarks"`
	Status         string         `json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	Period         PeriodResponse `json:"period"`
}

type InvoiceListResponse struct {
	Invoices []InvoiceResponse `json:"invoices"`
}

type AvailablePeriodsResponse struct {
	Periods []PeriodResponse `json:"periods"`
}

type PeriodInvoiceStatusResponse struct {
	PeriodID         uuid.UUID `json:"period_id"`
	PeriodNumber     int       `json:"period_number"`
	AmountPeriod     float64   `json:"amount_period"`
	InvoiceID        uuid.UUID `json:"invoice_id"`
	InvoiceStatus    string    `json:"invoice_status"`
	InvoiceCreatedAt time.Time `json:"invoice_created_at"`
	PaidDate         time.Time `json:"paid_date"`
	IsInvoiced       bool      `json:"is_invoiced"`
	IsPaid           bool      `json:"is_paid"`
}

type ContractInvoiceStatusResponse struct {
	ContractID uuid.UUID                     `json:"contract_id"`
	Periods    []PeriodInvoiceStatusResponse `json:"periods"`
	Progress   struct {
		TotalPeriods    int     `json:"total_periods"`
		InvoicedPeriods int     `json:"invoiced_periods"`
		PaidPeriods     int     `json:"paid_periods"`
		TotalAmount     float64 `json:"total_amount"`
		InvoicedAmount  float64 `json:"invoiced_amount"`
		PaidAmount      float64 `json:"paid_amount"`
		PercentInvoiced float64 `json:"percent_invoiced"`
		PercentPaid     float64 `json:"percent_paid"`
	} `json:"progress"`
}
