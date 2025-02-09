package usecase

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"boonkosang/internal/requests"
	"boonkosang/internal/responses"
	"context"
	"errors"

	"github.com/google/uuid"
)

type GeneralCostUseCase interface {
	GetByProjectID(ctx context.Context, projectID uuid.UUID) (*responses.GeneralCostListResponse, error)
	GetByID(ctx context.Context, gID uuid.UUID) (*responses.GeneralCostResponse, error)
	Update(ctx context.Context, gID uuid.UUID, req requests.UpdateGeneralCostRequest) error
	GetType(ctx context.Context) ([]models.Type, error)
	UpdateActualCost(ctx context.Context, gID uuid.UUID, req requests.UpdateActualGeneralCostRequest) error
}

type generalCostUseCase struct {
	generalCostRepo repositories.GeneralCostRepository
	boqRepo         repositories.BOQRepository
}

func NewGeneralCostUsecase(generalCostRepo repositories.GeneralCostRepository, boqRepo repositories.BOQRepository) GeneralCostUseCase {
	return &generalCostUseCase{
		generalCostRepo: generalCostRepo,
		boqRepo:         boqRepo,
	}
}

func (u *generalCostUseCase) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*responses.GeneralCostListResponse, error) {
	return u.generalCostRepo.GetByProjectID(ctx, projectID)
}

func (u *generalCostUseCase) GetByID(ctx context.Context, gID uuid.UUID) (*responses.GeneralCostResponse, error) {
	generalCost, err := u.generalCostRepo.GetByID(ctx, gID)
	if err != nil {
		return nil, err
	}

	return &responses.GeneralCostResponse{
		GID:           generalCost.GID,
		BOQID:         generalCost.BOQID,
		TypeName:      generalCost.TypeName,
		ActualCost:    generalCost.ActualCost.Float64,
		EstimatedCost: generalCost.EstimatedCost.Float64,
	}, nil
}

func (u *generalCostUseCase) Update(ctx context.Context, gID uuid.UUID, req requests.UpdateGeneralCostRequest) error {
	// Get general cost to check if exists
	generalCost, err := u.generalCostRepo.GetByID(ctx, gID)
	if err != nil {
		return err
	}

	// Check BOQ status
	boq, err := u.boqRepo.GetByID(ctx, generalCost.BOQID)
	if err != nil {
		return err
	}
	if boq.Status != "draft" {
		return errors.New("can only update general cost for BOQ in draft status")
	}

	// Validate estimated cost
	if req.EstimatedCost < 0 {
		return errors.New("estimated cost must be positive")
	}

	return u.generalCostRepo.Update(ctx, gID, req)
}

func (u *generalCostUseCase) GetType(ctx context.Context) ([]models.Type, error) {
	return u.generalCostRepo.GetType(ctx)
}

func (u *generalCostUseCase) UpdateActualCost(ctx context.Context, gID uuid.UUID, req requests.UpdateActualGeneralCostRequest) error {
	// Basic validation
	if req.ActualCost < 0 {
		return errors.New("actual cost must be positive")
	}

	// Get general cost to check if exists
	generalCost, err := u.generalCostRepo.GetByID(ctx, gID)
	if err != nil {
		return err
	}

	if generalCost == nil {
		return errors.New("general cost not found")
	}

	// Update the actual cost
	return u.generalCostRepo.UpdateActualCost(ctx, gID, req)
}
