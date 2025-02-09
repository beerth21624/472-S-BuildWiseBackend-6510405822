package repositories

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/requests"
	"boonkosang/internal/responses"
	"context"

	"github.com/google/uuid"
)

type GeneralCostRepository interface {
	GetByProjectID(ctx context.Context, projectID uuid.UUID) (*responses.GeneralCostListResponse, error)
	GetByID(ctx context.Context, gID uuid.UUID) (*models.GeneralCost, error)
	Update(ctx context.Context, gID uuid.UUID, req requests.UpdateGeneralCostRequest) error
	GetType(ctx context.Context) ([]models.Type, error)

	UpdateActualCost(ctx context.Context, gID uuid.UUID, req requests.UpdateActualGeneralCostRequest) error
	ValidateProjectStatus(ctx context.Context, projectID uuid.UUID) error
}
