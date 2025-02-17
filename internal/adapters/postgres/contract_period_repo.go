package postgres

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type periodRepository struct {
	db *sqlx.DB
}

func NewPeriodRepository(db *sqlx.DB) repositories.PeriodRepository {
	return &periodRepository{db: db}
}

func (r *periodRepository) CreatePeriod(ctx context.Context, contractID uuid.UUID, period *models.Period) error {
	query := `
		INSERT INTO period (
			period_id,
			contract_id,
			period_number,
			amount_period,
			delivered_within
		) VALUES (
			:period_id,
			:contract_id,
			:period_number,
			:amount_period,
			:delivered_within
		)`

	period.PeriodID = uuid.New()
	period.ContractID = contractID

	_, err := r.db.NamedExecContext(ctx, query, period)
	if err != nil {
		return fmt.Errorf("failed to create period: %w", err)
	}

	// Create associated job periods
	for _, job := range period.Jobs {
		if err := r.createJobPeriod(ctx, period.PeriodID, &job); err != nil {
			return err
		}
	}

	return nil
}

func (r *periodRepository) createJobPeriod(ctx context.Context, periodID uuid.UUID, jobPeriod *models.JobPeriod) error {
	query := `
		INSERT INTO job_period (
			job_id,
			period_id,
			job_amount
		) VALUES (
			:job_id,
			:period_id,
			:job_amount
		)`

	jobPeriod.PeriodID = periodID

	_, err := r.db.NamedExecContext(ctx, query, jobPeriod)
	if err != nil {
		return fmt.Errorf("failed to create job period: %w", err)
	}

	return nil
}

func (r *periodRepository) UpdatePeriod(ctx context.Context, period *models.Period) error {
	query := `
		UPDATE period 
		SET amount_period = :amount_period,
			delivered_within = :delivered_within
		WHERE period_id = :period_id AND contract_id = :contract_id`

	result, err := r.db.NamedExecContext(ctx, query, period)
	if err != nil {
		return fmt.Errorf("failed to update period: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("period not found")
	}

	// Update job periods
	for _, job := range period.Jobs {
		if err := r.updateJobPeriod(ctx, period.PeriodID, &job); err != nil {
			return err
		}
	}

	return nil
}

func (r *periodRepository) updateJobPeriod(ctx context.Context, periodID uuid.UUID, jobPeriod *models.JobPeriod) error {
	query := `
		UPDATE job_period 
		SET job_amount = :job_amount
		WHERE job_id = :job_id AND period_id = :period_id`

	result, err := r.db.NamedExecContext(ctx, query, jobPeriod)
	if err != nil {
		return fmt.Errorf("failed to update job period: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("job period not found")
	}

	return nil
}

func (r *periodRepository) DeletePeriodsByContractID(ctx context.Context, contractID uuid.UUID) error {
	// First delete all associated job periods
	jobPeriodsQuery := `
		DELETE FROM job_period 
		WHERE period_id IN (
			SELECT period_id 
			FROM period 
			WHERE contract_id = $1
		)`

	_, err := r.db.ExecContext(ctx, jobPeriodsQuery, contractID)
	if err != nil {
		return fmt.Errorf("failed to delete job periods: %w", err)
	}

	// Then delete the periods
	periodsQuery := `DELETE FROM period WHERE contract_id = $1`
	_, err = r.db.ExecContext(ctx, periodsQuery, contractID)
	if err != nil {
		return fmt.Errorf("failed to delete periods: %w", err)
	}

	return nil
}

func (r *periodRepository) GetPeriodsByContractID(ctx context.Context, contractID uuid.UUID) ([]models.Period, error) {
	periods := []models.Period{}
	periodsQuery := `SELECT * FROM period WHERE contract_id = $1 ORDER BY period_number`

	err := r.db.SelectContext(ctx, &periods, periodsQuery, contractID)
	if err != nil {
		return nil, fmt.Errorf("failed to get periods: %w", err)
	}

	for i := range periods {
		jobPeriodsQuery := `
            SELECT 
                jp.job_id,
                jp.period_id,
                jp.job_amount,
                j.job_id as "job.job_id",
                j.name as "job.name",
                j.description as "job.description",
                j.unit as "job.unit"
            FROM job_period jp
            LEFT JOIN job j ON jp.job_id = j.job_id
            WHERE jp.period_id = $1`

		type JobPeriodWithDetail struct {
			models.JobPeriod
			Job models.Job `db:"job"`
		}

		jobPeriods := []JobPeriodWithDetail{}
		err := r.db.SelectContext(ctx, &jobPeriods, jobPeriodsQuery, periods[i].PeriodID)
		if err != nil {
			return nil, fmt.Errorf("failed to get job periods: %w", err)
		}

		periods[i].Jobs = make([]models.JobPeriod, len(jobPeriods))
		for j, jp := range jobPeriods {
			periods[i].Jobs[j] = models.JobPeriod{
				JobID:     jp.JobID,
				PeriodID:  jp.PeriodID,
				JobAmount: jp.JobAmount,
				JobDetail: jp.Job,
			}
		}
	}

	return periods, nil
}
