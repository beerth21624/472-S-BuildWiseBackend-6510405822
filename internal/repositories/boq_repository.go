package repositories

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/requests"
	"boonkosang/internal/responses"
	"context"

	"github.com/google/uuid"
)

type BOQRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.BOQ, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) (*models.BOQ, error)
	Approve(ctx context.Context, boqID uuid.UUID) error
	GetBoqWithProject(ctx context.Context, projectID uuid.UUID) (*responses.BOQResponse, error)
	AddBOQJob(ctx context.Context, boqID uuid.UUID, req requests.BOQJobRequest) error
	UpdateBOQJob(ctx context.Context, boqID uuid.UUID, req requests.BOQJobRequest) error
	DeleteBOQJob(ctx context.Context, boqID uuid.UUID, jobID uuid.UUID) error

	GetBOQGeneralCosts(ctx context.Context, boqID uuid.UUID) ([]models.BOQGeneralCost, error)
	GetBOQDetails(ctx context.Context, projectID uuid.UUID) ([]models.BOQDetails, error)
	GetBOQMaterialDetails(ctx context.Context, projectID uuid.UUID) ([]models.BOQMaterialDetails, error)
}
