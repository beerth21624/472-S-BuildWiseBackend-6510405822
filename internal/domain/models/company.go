package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Company struct {
	CompanyID uuid.UUID       `db:"company_id" json:"company_id"`
	Name      string          `db:"name" json:"name" validate:"required"`
	Email     string          `db:"email" json:"email" validate:"required,email"`
	Tel       string          `db:"tel" json:"tel" validate:"required,len=10"`
	Address   json.RawMessage `db:"address" json:"address"`
	TaxID     string          `db:"tax_id" json:"tax_id" validate:"required,len=13,numeric"`
}
