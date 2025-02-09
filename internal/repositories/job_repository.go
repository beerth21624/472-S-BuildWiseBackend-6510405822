package repositories

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/requests"
	"boonkosang/internal/responses"
	"context"

	"github.com/google/uuid"
)

type JobRepository interface {
	Create(ctx context.Context, req requests.CreateJobRequest) (*responses.JobResponse, error)
	Update(ctx context.Context, id uuid.UUID, req requests.UpdateJobRequest) error
	List(ctx context.Context) (*responses.JobListResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Job, error)
	GetJobMaterialByID(ctx context.Context, id uuid.UUID) (responses.JobMaterialResponse, error)
	Delete(ctx context.Context, jobID uuid.UUID) error

	AddJobMaterial(ctx context.Context, jobID uuid.UUID, req requests.AddJobMaterialRequest) error
	DeleteJobMaterial(ctx context.Context, jobID uuid.UUID, materialID string) error
	UpdateJobMaterialQuantity(ctx context.Context, jobID uuid.UUID, req requests.UpdateJobMaterialQuantityRequest) error
}
