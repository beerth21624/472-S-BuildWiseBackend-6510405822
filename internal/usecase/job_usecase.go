package usecase

import (
	"boonkosang/internal/repositories"
	"boonkosang/internal/requests"
	"boonkosang/internal/responses"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type JobUseCase interface {
	Create(ctx context.Context, req requests.CreateJobRequest) (*responses.JobResponse, error)
	Update(ctx context.Context, id uuid.UUID, req requests.UpdateJobRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (responses.JobMaterialResponse, error)
	GetJobList(ctx context.Context) (responses.JobListResponse, error)
	Delete(ctx context.Context, jobID uuid.UUID) error
	AddMaterial(ctx context.Context, jobID uuid.UUID, req requests.AddJobMaterialRequest) error
	DeleteMaterial(ctx context.Context, jobID uuid.UUID, materialID string) error
	UpdateMaterialQuantity(ctx context.Context, jobID uuid.UUID, req requests.UpdateJobMaterialQuantityRequest) error
}

type jobUseCase struct {
	jobRepo repositories.JobRepository
}

func NewJobUseCase(jobRepo repositories.JobRepository) JobUseCase {
	return &jobUseCase{
		jobRepo: jobRepo,
	}
}

func (u *jobUseCase) Create(ctx context.Context, req requests.CreateJobRequest) (*responses.JobResponse, error) {
	job, err := u.jobRepo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}
	return job, nil
}

func (u *jobUseCase) Update(ctx context.Context, id uuid.UUID, req requests.UpdateJobRequest) error {
	existing, err := u.jobRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("job not found")
	}

	return u.jobRepo.Update(ctx, id, req)
}

func (u *jobUseCase) GetByID(ctx context.Context, id uuid.UUID) (responses.JobMaterialResponse, error) {
	return u.jobRepo.GetJobMaterialByID(ctx, id)
}

func (u *jobUseCase) GetJobList(ctx context.Context) (responses.JobListResponse, error) {
	jobList, err := u.jobRepo.List(ctx)
	if err != nil {
		return responses.JobListResponse{}, err
	}
	return *jobList, nil
}

func (u *jobUseCase) Delete(ctx context.Context, jobID uuid.UUID) error {
	existing, err := u.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("job not found")
	}

	return u.jobRepo.Delete(ctx, jobID)

}

func (u *jobUseCase) AddMaterial(ctx context.Context, jobID uuid.UUID, req requests.AddJobMaterialRequest) error {
	existing, err := u.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("job not found")
	}

	return u.jobRepo.AddJobMaterial(ctx, jobID, req)
}

func (u *jobUseCase) DeleteMaterial(ctx context.Context, jobID uuid.UUID, materialID string) error {
	existing, err := u.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("job not found")
	}

	return u.jobRepo.DeleteJobMaterial(ctx, jobID, materialID)
}

func (u *jobUseCase) UpdateMaterialQuantity(ctx context.Context, jobID uuid.UUID, req requests.UpdateJobMaterialQuantityRequest) error {
	existing, err := u.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("job not found")
	}

	return u.jobRepo.UpdateJobMaterialQuantity(ctx, jobID, req)
}
