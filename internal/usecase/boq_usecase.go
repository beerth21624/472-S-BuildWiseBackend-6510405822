package usecase

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"boonkosang/internal/requests"
	"boonkosang/internal/responses"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type BOQUsecase interface {
	Approve(ctx context.Context, boqID uuid.UUID) error
	GetBoqWithProject(ctx context.Context, project_id uuid.UUID) (*responses.BOQResponse, error)
	AddBOQJob(ctx context.Context, boqID uuid.UUID, req requests.BOQJobRequest) error
	UpdateBOQJob(ctx context.Context, boqID uuid.UUID, req requests.BOQJobRequest) error
	DeleteBOQJob(ctx context.Context, boqID uuid.UUID, jobID uuid.UUID) error
	GetBOQSummary(ctx context.Context, projectID uuid.UUID) (*responses.BOQSummaryResponse, error)
}

type boqUsecase struct {
	boqRepo     repositories.BOQRepository
	projectRepo repositories.ProjectRepository
}

func NewBOQUsecase(boqRepo repositories.BOQRepository, projectRepo repositories.ProjectRepository) BOQUsecase {
	return &boqUsecase{
		boqRepo:     boqRepo,
		projectRepo: projectRepo,
	}
}

func (u *boqUsecase) Approve(ctx context.Context, boqID uuid.UUID) error {
	return u.boqRepo.Approve(ctx, boqID)
}
func (u *boqUsecase) GetBoqWithProject(ctx context.Context, project_id uuid.UUID) (*responses.BOQResponse, error) {
	return u.boqRepo.GetBoqWithProject(ctx, project_id)
}

func (u *boqUsecase) AddBOQJob(ctx context.Context, boqID uuid.UUID, req requests.BOQJobRequest) error {
	return u.boqRepo.AddBOQJob(ctx, boqID, req)
}

func (u *boqUsecase) UpdateBOQJob(ctx context.Context, boqID uuid.UUID, req requests.BOQJobRequest) error {
	return u.boqRepo.UpdateBOQJob(ctx, boqID, req)
}

func (u *boqUsecase) DeleteBOQJob(ctx context.Context, boqID uuid.UUID, jobID uuid.UUID) error {
	return u.boqRepo.DeleteBOQJob(ctx, boqID, jobID)
}

func (u *boqUsecase) GetBOQSummary(ctx context.Context, projectID uuid.UUID) (*responses.BOQSummaryResponse, error) {
	boq, err := u.boqRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error getting BOQ: %w", err)
	}

	if boq.Status != models.BOQStatusApproved {
		return nil, errors.New("BOQ is not approved")
	}

	// Get all required data
	generalCosts, err := u.boqRepo.GetBOQGeneralCosts(ctx, boq.BOQID)
	if err != nil {
		return nil, fmt.Errorf("error getting general costs: %w", err)
	}

	details, err := u.boqRepo.GetBOQDetails(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error getting BOQ details: %w", err)
	}

	materials, err := u.boqRepo.GetBOQMaterialDetails(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("error getting material details: %w", err)
	}

	// Transform data to DTOs
	return transformToResponse(details[0], generalCosts, details, materials), nil
}

func transformGeneralCosts(costs []models.BOQGeneralCost) []responses.GeneralCostDTO {
	dtos := make([]responses.GeneralCostDTO, len(costs))
	for i, cost := range costs {
		dtos[i] = responses.GeneralCostDTO{
			TypeName:      cost.TypeName,
			EstimatedCost: cost.EstimatedCost,
		}
	}
	return dtos
}

// Update the transform function to handle the grouping
func transformToResponse(firstDetail models.BOQDetails, generalCosts []models.BOQGeneralCost, details []models.BOQDetails, materials []models.BOQMaterialDetails) *responses.BOQSummaryResponse {
	response := &responses.BOQSummaryResponse{
		ProjectInfo: responses.ProjectInfo{
			ProjectName:    firstDetail.ProjectName,
			ProjectAddress: json.RawMessage(firstDetail.ProjectAddress.String),
		},
		GeneralCosts: transformGeneralCosts(generalCosts),
		Details:      transformBOQDetailsWithMaterials(details, materials),
	}

	response.SummaryMetrics = calculateSummaryMetrics(response.GeneralCosts, response.Details)

	return response
}

func transformBOQDetailsWithMaterials(details []models.BOQDetails, materials []models.BOQMaterialDetails) []responses.BOQDetailDTO {
	// Create a map to group materials by JobID
	materialsByJob := make(map[uuid.UUID][]models.BOQMaterialDetails)
	for _, material := range materials {
		materialsByJob[material.JobID] = append(materialsByJob[material.JobID], material)
	}

	dtos := make([]responses.BOQDetailDTO, len(details))
	for i, detail := range details {
		totalEstimatedPrice := detail.EstimatedPrice.Float64 * float64(detail.Quantity)
		totalLaborCost := detail.LaborCost * float64(detail.Quantity)

		// Transform materials for this job
		jobMaterials := transformMaterials(materialsByJob[detail.JobID])

		dtos[i] = responses.BOQDetailDTO{
			JobID:               detail.JobID,
			JobName:             detail.JobName,
			Description:         detail.Description.String,
			Quantity:            detail.Quantity,
			Unit:                detail.Unit,
			LaborCost:           detail.LaborCost,
			EstimatedPrice:      detail.EstimatedPrice.Float64,
			TotalEstimatedPrice: totalEstimatedPrice,
			TotalLaborCost:      totalLaborCost,
			Total:               detail.Total.Float64,
			Materials:           jobMaterials,
		}
	}
	return dtos
}

func transformMaterials(materials []models.BOQMaterialDetails) []responses.MaterialDTO {
	if materials == nil {
		return []responses.MaterialDTO{}
	}

	dtos := make([]responses.MaterialDTO, len(materials))
	for i, material := range materials {
		quantity := material.Quantity.Float64
		estimatedPrice := material.EstimatedPrice.Float64

		dtos[i] = responses.MaterialDTO{
			JobID:          material.JobID,
			JobName:        material.JobName,
			MaterialName:   material.MaterialName,
			Quantity:       quantity,
			Unit:           material.Unit,
			EstimatedPrice: estimatedPrice,
			Total:          material.Total.Float64,
		}
	}
	return dtos
}

func calculateSummaryMetrics(generalCosts []responses.GeneralCostDTO, details []responses.BOQDetailDTO) responses.SummaryMetrics {
	var metrics responses.SummaryMetrics

	for _, cost := range generalCosts {
		metrics.TotalGeneralCost += cost.EstimatedCost
	}

	for _, detail := range details {
		metrics.TotalLaborCost += detail.TotalLaborCost
		metrics.TotalEstimatedPrice += detail.TotalEstimatedPrice
		metrics.TotalAmount += detail.Total

		// Calculate material costs for this job
		for _, material := range detail.Materials {
			metrics.TotalMaterialCost += material.Total
		}
	}

	metrics.GrandTotal = metrics.TotalGeneralCost + metrics.TotalLaborCost + metrics.TotalMaterialCost

	return metrics
}
