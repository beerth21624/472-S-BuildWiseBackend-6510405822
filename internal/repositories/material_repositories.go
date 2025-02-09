package repositories

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/requests"
	"context"

	"github.com/google/uuid"
)

type MaterialRepository interface {
	Create(ctx context.Context, req requests.CreateMaterialRequest) (*models.Material, error)
	Update(ctx context.Context, materialID string, req requests.UpdateMaterialRequest) error
	Delete(ctx context.Context, materialID string) error
	GetByID(ctx context.Context, materialID string) (*models.Material, error)
	List(ctx context.Context) ([]models.Material, error)

	GetMaterialPricesByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.MaterialPriceInfo, error)
	UpdateEstimatedPrices(ctx context.Context, boqID uuid.UUID, materialID string, estimatedPrice float64) error
	GetBOQStatus(ctx context.Context, boqID uuid.UUID) (string, error)
	UpdateActualPrice(ctx context.Context, boqID uuid.UUID, req requests.UpdateMaterialActualPriceRequest) error
	GetProjectStatus(ctx context.Context, projectID uuid.UUID) (string, error)
	GetQuotationStatus(ctx context.Context, projectID uuid.UUID) (string, error)
}
