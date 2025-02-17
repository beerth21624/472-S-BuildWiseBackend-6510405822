package usecase

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"boonkosang/internal/requests"
	"boonkosang/internal/responses"
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type ContractUseCase interface {
	Create(ctx context.Context, req *requests.CreateContractRequest) error
	Update(ctx context.Context, projectID uuid.UUID, req *requests.UpdateContractRequest) error
	Delete(ctx context.Context, projectID uuid.UUID) error
	GetByProjectID(ctx context.Context, projectID uuid.UUID) (*responses.ContractResponse, error)
}

type contractUseCase struct {
	contractRepo repositories.ContractRepository
	periodRepo   repositories.PeriodRepository
	projectRepo  repositories.ProjectRepository
}

func NewContractUsecase(
	contractRepo repositories.ContractRepository,
	periodRepo repositories.PeriodRepository,
	projectRepo repositories.ProjectRepository,
) ContractUseCase {
	return &contractUseCase{
		contractRepo: contractRepo,
		periodRepo:   periodRepo,
		projectRepo:  projectRepo,
	}
}

func (u *contractUseCase) Create(ctx context.Context, req *requests.CreateContractRequest) error {
	// Create the contract first
	contract := &models.Contract{
		ProjectID: req.ProjectID,
		ProjectDescription: sql.NullString{
			String: req.ProjectDescription,
			Valid:  req.ProjectDescription != "",
		},
		AreaSize: sql.NullFloat64{
			Float64: req.AreaSize,
			Valid:   req.AreaSize != 0,
		},
		StartDate: sql.NullTime{
			Time:  req.StartDate,
			Valid: !req.StartDate.IsZero(),
		},
		EndDate: sql.NullTime{
			Time:  req.EndDate,
			Valid: !req.EndDate.IsZero(),
		},
		ForceMajeure: sql.NullString{
			String: req.ForceMajeure,
			Valid:  req.ForceMajeure != "",
		},
		BreachOfContract: sql.NullString{
			String: req.BreachOfContract,
			Valid:  req.BreachOfContract != "",
		},
		EndOfContract: sql.NullString{
			String: req.EndOfContract,
			Valid:  req.EndOfContract != "",
		},
		TerminationContract: sql.NullString{
			String: req.TerminationContract,
			Valid:  req.TerminationContract != "",
		},
		Amendment: sql.NullString{
			String: req.Amendment,
			Valid:  req.Amendment != "",
		},
		GuaranteeWithin: sql.NullInt32{
			Int32: int32(req.GuaranteeWithin),
			Valid: req.GuaranteeWithin != 0,
		},
		RetentionMoney: sql.NullFloat64{
			Float64: req.RetentionMoney,
			Valid:   req.RetentionMoney != 0,
		},
		PayWithin: sql.NullInt32{
			Int32: int32(req.PayWithin),
			Valid: req.PayWithin != 0,
		},
		ValidateWithin: sql.NullInt32{
			Int32: int32(req.ValidateWithin),
			Valid: req.ValidateWithin != 0,
		},
		Format: models.StringArray(req.Format),
	}

	if err := u.contractRepo.Create(ctx, contract); err != nil {
		return fmt.Errorf("failed to create contract: %w", err)
	}

	for _, periodReq := range req.Periods {
		period := &models.Period{
			ContractID:      contract.ContractID,
			PeriodNumber:    periodReq.PeriodNumber,
			AmountPeriod:    periodReq.AmountPeriod,
			DeliveredWithin: periodReq.DeliveredWithin,
			Jobs:            make([]models.JobPeriod, len(periodReq.Jobs)),
		}

		// Convert job requests to job periods
		for i, jobReq := range periodReq.Jobs {
			period.Jobs[i] = models.JobPeriod{
				JobID:     jobReq.JobID,
				JobAmount: jobReq.JobAmount,
			}
		}

		if err := u.periodRepo.CreatePeriod(ctx, contract.ContractID, period); err != nil {
			return fmt.Errorf("failed to create period: %w", err)
		}
	}

	return nil
}
func (u *contractUseCase) Update(ctx context.Context, projectID uuid.UUID, req *requests.UpdateContractRequest) error {
	// Get existing contract
	contract, err := u.contractRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	// Update basic contract fields
	if req.ProjectDescription != "" {
		contract.ProjectDescription = sql.NullString{
			String: req.ProjectDescription,
			Valid:  true,
		}
	}

	if req.AreaSize != 0 {
		contract.AreaSize = sql.NullFloat64{
			Float64: req.AreaSize,
			Valid:   true,
		}
	}

	if !req.StartDate.IsZero() {
		contract.StartDate = sql.NullTime{
			Time:  req.StartDate,
			Valid: true,
		}
	}

	if !req.EndDate.IsZero() {
		contract.EndDate = sql.NullTime{
			Time:  req.EndDate,
			Valid: true,
		}
	}

	if req.ForceMajeure != "" {
		contract.ForceMajeure = sql.NullString{
			String: req.ForceMajeure,
			Valid:  true,
		}
	}

	if req.BreachOfContract != "" {
		contract.BreachOfContract = sql.NullString{
			String: req.BreachOfContract,
			Valid:  true,
		}
	}

	if req.EndOfContract != "" {
		contract.EndOfContract = sql.NullString{
			String: req.EndOfContract,
			Valid:  true,
		}
	}

	if req.TerminationContract != "" {
		contract.TerminationContract = sql.NullString{
			String: req.TerminationContract,
			Valid:  true,
		}
	}

	if req.Amendment != "" {
		contract.Amendment = sql.NullString{
			String: req.Amendment,
			Valid:  true,
		}
	}

	if req.GuaranteeWithin > 0 {
		contract.GuaranteeWithin = sql.NullInt32{
			Int32: int32(req.GuaranteeWithin),
			Valid: true,
		}
	}

	if req.RetentionMoney > 0 {
		contract.RetentionMoney = sql.NullFloat64{
			Float64: req.RetentionMoney,
			Valid:   true,
		}
	}

	if req.PayWithin > 0 {
		contract.PayWithin = sql.NullInt32{
			Int32: int32(req.PayWithin),
			Valid: true,
		}
	}

	if req.ValidateWithin > 0 {
		contract.ValidateWithin = sql.NullInt32{
			Int32: int32(req.ValidateWithin),
			Valid: true,
		}
	}

	if len(req.Format) > 0 {
		contract.Format = models.StringArray(req.Format)
	}

	if err := u.contractRepo.Update(ctx, contract); err != nil {
		return err
	}

	if len(req.Periods) > 0 {
		existingPeriods, err := u.periodRepo.GetPeriodsByContractID(ctx, contract.ContractID)
		if err != nil {
			return fmt.Errorf("failed to get existing periods: %w", err)
		}

		existingPeriodMap := make(map[int]models.Period)
		for _, p := range existingPeriods {
			existingPeriodMap[p.PeriodNumber] = p
		}

		for _, periodReq := range req.Periods {
			if existingPeriod, exists := existingPeriodMap[periodReq.PeriodNumber]; exists {
				// Update existing period
				period := &models.Period{
					PeriodID:        existingPeriod.PeriodID,
					ContractID:      contract.ContractID,
					PeriodNumber:    periodReq.PeriodNumber,
					AmountPeriod:    periodReq.AmountPeriod,
					DeliveredWithin: periodReq.DeliveredWithin,
				}

				// Update period jobs
				for _, jobReq := range periodReq.Jobs {
					// Find if job already exists in period
					var existingJob *models.JobPeriod
					for _, j := range existingPeriod.Jobs {
						if j.JobID == jobReq.JobID {
							existingJob = &j
							break
						}
					}

					if existingJob != nil {
						// Update existing job
						jobPeriod := models.JobPeriod{
							JobID:     jobReq.JobID,
							PeriodID:  period.PeriodID,
							JobAmount: jobReq.JobAmount,
						}
						period.Jobs = append(period.Jobs, jobPeriod)
					} else {
						// Add new job to period
						jobPeriod := models.JobPeriod{
							JobID:     jobReq.JobID,
							PeriodID:  period.PeriodID,
							JobAmount: jobReq.JobAmount,
						}
						period.Jobs = append(period.Jobs, jobPeriod)
					}
				}

				// Update period in database
				if err := u.periodRepo.UpdatePeriod(ctx, period); err != nil {
					return fmt.Errorf("failed to update period %d: %w", periodReq.PeriodNumber, err)
				}
			} else {
				// Create new period
				newPeriod := &models.Period{
					ContractID:      contract.ContractID,
					PeriodNumber:    periodReq.PeriodNumber,
					AmountPeriod:    periodReq.AmountPeriod,
					DeliveredWithin: periodReq.DeliveredWithin,
				}

				// Add jobs to new period
				for _, jobReq := range periodReq.Jobs {
					jobPeriod := models.JobPeriod{
						JobID:     jobReq.JobID,
						JobAmount: jobReq.JobAmount,
					}
					newPeriod.Jobs = append(newPeriod.Jobs, jobPeriod)
				}

				// Create new period in database
				if err := u.periodRepo.CreatePeriod(ctx, contract.ContractID, newPeriod); err != nil {
					return fmt.Errorf("failed to create new period %d: %w", periodReq.PeriodNumber, err)
				}
			}
		}
	}

	return nil
}

func (u *contractUseCase) Delete(ctx context.Context, projectID uuid.UUID) error {
	contract, err := u.contractRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	if err := u.periodRepo.DeletePeriodsByContractID(ctx, contract.ContractID); err != nil {
		return fmt.Errorf("failed to delete periods: %w", err)
	}
	return u.contractRepo.Delete(ctx, projectID)
}
func (u *contractUseCase) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*responses.ContractResponse, error) {
	// Get contract
	contract, err := u.contractRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Get periods for the contract
	periods, err := u.periodRepo.GetPeriodsByContractID(ctx, contract.ContractID)
	if err != nil {
		return nil, fmt.Errorf("failed to get periods: %w", err)
	}
	contract.Periods = periods

	// Convert to response format
	response := &responses.ContractResponse{
		ContractID: contract.ContractID,
		ProjectID:  contract.ProjectID,
		Format:     []string(contract.Format),
		CreatedAt:  contract.CreatedAt,
	}

	// Handle nullable fields
	if contract.ProjectDescription.Valid {
		response.ProjectDescription = contract.ProjectDescription.String
	}
	if contract.AreaSize.Valid {
		response.AreaSize = contract.AreaSize.Float64
	}
	if contract.StartDate.Valid {
		response.StartDate = contract.StartDate.Time
	}
	if contract.EndDate.Valid {
		response.EndDate = contract.EndDate.Time
	}
	if contract.ForceMajeure.Valid {
		response.ForceMajeure = contract.ForceMajeure.String
	}
	if contract.BreachOfContract.Valid {
		response.BreachOfContract = contract.BreachOfContract.String
	}
	if contract.EndOfContract.Valid {
		response.EndOfContract = contract.EndOfContract.String
	}
	if contract.TerminationContract.Valid {
		response.TerminationContract = contract.TerminationContract.String
	}
	if contract.Amendment.Valid {
		response.Amendment = contract.Amendment.String
	}
	if contract.GuaranteeWithin.Valid {
		response.GuaranteeWithin = int(contract.GuaranteeWithin.Int32)
	}
	if contract.RetentionMoney.Valid {
		response.RetentionMoney = contract.RetentionMoney.Float64
	}
	if contract.PayWithin.Valid {
		response.PayWithin = int(contract.PayWithin.Int32)
	}
	if contract.ValidateWithin.Valid {
		response.ValidateWithin = int(contract.ValidateWithin.Int32)
	}
	if contract.UpdatedAt.Valid {
		response.UpdatedAt = contract.UpdatedAt.Time
	}

	// Handle periods and their jobs
	response.Periods = make([]responses.PeriodResponse, len(contract.Periods))
	for i, period := range contract.Periods {
		periodResponse := responses.PeriodResponse{
			PeriodID:        period.PeriodID,
			PeriodNumber:    period.PeriodNumber,
			AmountPeriod:    period.AmountPeriod,
			DeliveredWithin: period.DeliveredWithin,
			Jobs:            make([]responses.JobPeriodResponse, len(period.Jobs)),
		}

		for j, jobPeriod := range period.Jobs {
			periodResponse.Jobs[j] = responses.JobPeriodResponse{
				JobID:     jobPeriod.JobID,
				JobAmount: jobPeriod.JobAmount,
				Job: responses.JobResponse{
					JobID:       jobPeriod.JobDetail.JobID,
					Name:        jobPeriod.JobDetail.Name,
					Description: jobPeriod.JobDetail.Description.String,
					Unit:        jobPeriod.JobDetail.Unit,
				},
			}
		}

		response.Periods[i] = periodResponse
	}

	return response, nil
}
func calculateRetentionMoney(jobs []models.QuotationJob) float64 {
	var total float64
	for _, job := range jobs {
		if job.TotalSellingPrice.Valid {
			total += job.TotalSellingPrice.Float64
		}
	}
	return total * 0.05 // 5% retention
}
