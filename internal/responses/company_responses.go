package responses

import (
	"encoding/json"

	"github.com/google/uuid"
)

type CompanyResponse struct {
	CompanyID uuid.UUID       `json:"company_id"`
	Name      string          `json:"name"`
	Email     string          `json:"email"`
	Tel       string          `json:"tel"`
	Address   json.RawMessage `json:"address"`
	TaxID     string          `json:"tax_id"`
	IsNew     bool            `json:"-"`
}

// Example API response structures
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type CompanyAPIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Data    CompanyResponse `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}
