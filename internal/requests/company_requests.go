package requests

import "encoding/json"

type CreateCompanyRequest struct {
	Name    string          `json:"name"`
	Email   string          `json:"email"`
	Tel     string          `json:"tel"`
	Address json.RawMessage `json:"address"`
	TaxID   string          `json:"tax_id"`
}

type UpdateCompanyRequest struct {
	Name    string          `json:"name"`
	Email   string          `json:"email"`
	Tel     string          `json:"tel"`
	Address json.RawMessage `json:"address"`
	TaxID   string          `json:"tax_id"`
}
