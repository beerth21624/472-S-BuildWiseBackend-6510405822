package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, a)
}

type Contract struct {
	ContractID          uuid.UUID   `db:"contract_id"`
	ProjectID           uuid.UUID   `db:"project_id"`
	ProjectDescription  string      `db:"project_description"`
	AreaSize            float64     `db:"area_size"`
	StartDate           time.Time   `db:"start_date"`
	EndDate             time.Time   `db:"end_date"`
	ForceMajeure        string      `db:"force_majeure"`
	BreachOfContract    string      `db:"breach_of_contract"`
	EndOfContract       string      `db:"end_of_contract"`
	TerminationContract string      `db:"termination_of_contract"`
	Amendment           string      `db:"amendment"`
	GuaranteeWithin     int         `db:"guarantee_within"`
	RetentionMoney      float64     `db:"retention_money"`
	PayWithin           int         `db:"pay_within"`
	ValidateWithin      int         `db:"validate_within"`
	Format              StringArray `db:"format"`
	CreatedAt           time.Time   `db:"created_at"`
	UpdatedAt           *time.Time  `db:"updated_at"`
	Periods             []Period    `db:"-"`
}

type Period struct {
	PeriodID        uuid.UUID   `db:"period_id"`
	ContractID      uuid.UUID   `db:"contract_id"`
	PeriodNumber    int         `db:"period_number"`
	AmountPeriod    float64     `db:"amount_period"`
	DeliveredWithin int         `db:"delivered_within"`
	Jobs            []JobPeriod `db:"-"`
}

type JobPeriod struct {
	JobID     uuid.UUID `db:"job_id"`
	PeriodID  uuid.UUID `db:"period_id"`
	JobAmount float64   `db:"job_amount"`
	Job       Job       `db:"-"`
}
