package responses

import (
	"time"

	"github.com/google/uuid"
)

type ContractResponse struct {
	ContractID          uuid.UUID        `json:"contract_id"`
	ProjectID           uuid.UUID        `json:"project_id"`
	ProjectDescription  string           `json:"project_description"`
	AreaSize            float64          `json:"area_size"`
	StartDate           time.Time        `json:"start_date"`
	EndDate             time.Time        `json:"end_date"`
	ForceMajeure        string           `json:"force_majeure"`
	BreachOfContract    string           `json:"breach_of_contract"`
	EndOfContract       string           `json:"end_of_contract"`
	TerminationContract string           `json:"termination_of_contract"`
	Amendment           string           `json:"amendment"`
	GuaranteeWithin     int              `json:"guarantee_within"`
	RetentionMoney      float64          `json:"retention_money"`
	PayWithin           int              `json:"pay_within"`
	ValidateWithin      int              `json:"validate_within"`
	Format              []string         `json:"format"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
	Periods             []PeriodResponse `json:"periods"`
}

type PeriodResponse struct {
	PeriodID        uuid.UUID           `json:"period_id"`
	PeriodNumber    int                 `json:"period_number"`
	AmountPeriod    float64             `json:"amount_period"`
	DeliveredWithin int                 `json:"delivered_within"`
	Jobs            []JobPeriodResponse `json:"jobs"`
}

type JobPeriodResponse struct {
	JobID     uuid.UUID   `json:"job_id"`
	JobAmount float64     `json:"job_amount"`
	Job       JobResponse `json:"job"`
}
