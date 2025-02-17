package repositories

import (
	"boonkosang/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

type PeriodRepository interface {
	CreatePeriod(ctx context.Context, contractID uuid.UUID, period *models.Period) error
	UpdatePeriod(ctx context.Context, period *models.Period) error
	DeletePeriodsByContractID(ctx context.Context, contractID uuid.UUID) error
	GetPeriodsByContractID(ctx context.Context, contractID uuid.UUID) ([]models.Period, error)
}
