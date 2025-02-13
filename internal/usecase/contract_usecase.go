package usecase

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"boonkosang/internal/requests"
	"boonkosang/internal/responses"
	"context"
	"database/sql"
	"fmt"
	"time"

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
	projectRepo  repositories.ProjectRepository
}

func NewContractUsecase(
	contractRepo repositories.ContractRepository,
	projectRepo repositories.ProjectRepository,
) ContractUseCase {
	return &contractUseCase{
		contractRepo: contractRepo,
		projectRepo:  projectRepo,
	}
}
func (u *contractUseCase) Create(ctx context.Context, req *requests.CreateContractRequest) error {
	contract := &models.Contract{
		ProjectID: req.ProjectID,
	}

	if err := u.contractRepo.Create(ctx, contract); err != nil {
		return fmt.Errorf("failed to create contract: %w", err)
	}

	return nil
}

func (u *contractUseCase) Update(ctx context.Context, projectID uuid.UUID, req *requests.UpdateContractRequest) error {
	// Get existing contract
	contract, err := u.contractRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	// Update fields only if they are provided in the request
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

	now := time.Now()
	contract.UpdatedAt = sql.NullTime{
		Time:  now,
		Valid: true,
	}

	// Update contract
	if err := u.contractRepo.Update(ctx, contract); err != nil {
		return err
	}

	return nil
}

func (u *contractUseCase) Delete(ctx context.Context, projectID uuid.UUID) error {
	return u.contractRepo.Delete(ctx, projectID)
}
func (u *contractUseCase) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*responses.ContractResponse, error) {
	contract, err := u.contractRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

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

	// Handle periods if they exist
	response.Periods = make([]responses.PeriodResponse, len(contract.Periods))
	for i, period := range contract.Periods {
		periodResponse := responses.PeriodResponse{
			PeriodID:        period.PeriodID,
			PeriodNumber:    period.PeriodNumber,
			AmountPeriod:    period.AmountPeriod,
			DeliveredWithin: period.DeliveredWithin,
			Jobs:            make([]responses.JobPeriodResponse, len(period.Jobs)),
		}

		for j, job := range period.Jobs {
			periodResponse.Jobs[j] = responses.JobPeriodResponse{
				JobID:     job.JobID,
				JobAmount: job.JobAmount,
				Job:       responses.JobResponse{}, // Map job details here if needed
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
