package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Invoice struct {
	InvoiceID      uuid.UUID       `db:"invoice_id"`
	ProjectID      uuid.UUID       `db:"project_id"`
	PeriodID       uuid.UUID       `db:"period_id"`
	InvoiceDate    sql.NullTime    `db:"invoice_date"`
	PaymentDueDate sql.NullTime    `db:"payment_due_date"`
	PaidDate       sql.NullTime    `db:"paid_date"`
	PaymentTerm    sql.NullString  `db:"payment_term"`
	Remarks        sql.NullString  `db:"remarks"`
	Status         sql.NullString  `db:"status"`
	CreatedAt      time.Time       `db:"created_at"`
	UpdatedAt      sql.NullTime    `db:"updated_at"`
	Retention      sql.NullFloat64 `db:"retention"`

	// Related data (not stored in database)
	Period Period `db:"-"`
}

// PeriodInvoiceStatus provides status information about a period's invoice
type PeriodInvoiceStatus struct {
	PeriodID         uuid.UUID      `db:"period_id"`
	PeriodNumber     int            `db:"period_number"`
	AmountPeriod     float64        `db:"amount_period"`
	InvoiceID        uuid.NullUUID  `db:"invoice_id"`
	InvoiceStatus    sql.NullString `db:"invoice_status"`
	InvoiceCreatedAt sql.NullTime   `db:"invoice_created_at"`
	PaidDate         sql.NullTime   `db:"paid_date"`
}
