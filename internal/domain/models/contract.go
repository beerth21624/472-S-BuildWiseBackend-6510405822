package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
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

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	case string:
		return json.Unmarshal([]byte(v), a)
	default:
		return fmt.Errorf("unsupported type for StringArray: %T", value)
	}
}

type Contract struct {
	ContractID          uuid.UUID       `db:"contract_id"`
	ProjectID           uuid.UUID       `db:"project_id"`
	ProjectDescription  sql.NullString  `db:"project_description"`
	AreaSize            sql.NullFloat64 `db:"area_size"`
	StartDate           sql.NullTime    `db:"start_date"`
	EndDate             sql.NullTime    `db:"end_date"`
	ForceMajeure        sql.NullString  `db:"force_majeure"`
	BreachOfContract    sql.NullString  `db:"breach_of_contract"`
	EndOfContract       sql.NullString  `db:"end_of_contract"`
	TerminationContract sql.NullString  `db:"termination_of_contract"`
	Amendment           sql.NullString  `db:"amendment"`
	GuaranteeWithin     sql.NullInt32   `db:"guarantee_within"`
	RetentionMoney      sql.NullFloat64 `db:"retention_money"`
	PayWithin           sql.NullInt32   `db:"pay_within"`
	ValidateWithin      sql.NullInt32   `db:"validate_within"`
	Format              StringArray     `db:"format"`
	CreatedAt           time.Time       `db:"created_at"`
	UpdatedAt           sql.NullTime    `db:"updated_at"`
	Periods             []Period        `db:"-"`
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
