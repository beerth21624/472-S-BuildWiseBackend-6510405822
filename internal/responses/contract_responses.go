package responses

import (
	"boonkosang/internal/domain/models"
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
	Status              string           `json:"status"`
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
	Job       JobResponse `json:"job_detail"`
}

func ToContractResponse(c *models.Contract) *ContractResponse {
	response := &ContractResponse{
		ContractID: c.ContractID,
		ProjectID:  c.ProjectID,
		Format:     []string(c.Format),
		CreatedAt:  c.CreatedAt,
	}

	if c.ProjectDescription.Valid {
		response.ProjectDescription = c.ProjectDescription.String
	}
	if c.AreaSize.Valid {
		response.AreaSize = c.AreaSize.Float64
	}
	if c.StartDate.Valid {
		response.StartDate = c.StartDate.Time
	}
	if c.EndDate.Valid {
		response.EndDate = c.EndDate.Time
	}
	if c.ForceMajeure.Valid {
		response.ForceMajeure = c.ForceMajeure.String
	}
	if c.BreachOfContract.Valid {
		response.BreachOfContract = c.BreachOfContract.String
	}
	if c.EndOfContract.Valid {
		response.EndOfContract = c.EndOfContract.String
	}
	if c.TerminationContract.Valid {
		response.TerminationContract = c.TerminationContract.String
	}
	if c.Amendment.Valid {
		response.Amendment = c.Amendment.String
	}
	if c.GuaranteeWithin.Valid {
		response.GuaranteeWithin = int(c.GuaranteeWithin.Int32)
	}
	if c.RetentionMoney.Valid {
		response.RetentionMoney = c.RetentionMoney.Float64
	}
	if c.PayWithin.Valid {
		response.PayWithin = int(c.PayWithin.Int32)
	}
	if c.ValidateWithin.Valid {
		response.ValidateWithin = int(c.ValidateWithin.Int32)
	}
	if c.UpdatedAt.Valid {
		response.UpdatedAt = c.UpdatedAt.Time
	}

	return response
}
