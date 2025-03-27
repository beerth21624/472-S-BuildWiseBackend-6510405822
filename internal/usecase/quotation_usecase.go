package usecase

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"boonkosang/internal/requests"
	"boonkosang/internal/responses"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type QuotationUsecase interface {
	CreateOrGetQuotation(ctx context.Context, projectID uuid.UUID) (*responses.QuotationResponse, error)
	ApproveQuotation(ctx context.Context, projectID uuid.UUID) error
	ExportQuotation(ctx context.Context, projectID uuid.UUID) (*responses.QuotationExportData, error)

	UpdateProjectSellingPrice(ctx context.Context, req requests.UpdateProjectSellingPriceRequest) error
}

type quotationUsecase struct {
	quotationRepo repositories.QuotationRepository
}

func NewQuotationUsecase(
	quotationRepo repositories.QuotationRepository,
) QuotationUsecase {
	return &quotationUsecase{
		quotationRepo: quotationRepo,
	}
}

func (u *quotationUsecase) buildQuotationResponse(
	quotation *models.Quotation,
	jobs []models.QuotationJob,
	costs []models.QuotationGeneralCost,
) *responses.QuotationResponse {
	response := &responses.QuotationResponse{
		QuotationID: quotation.QuotationID,
		Status:      string(quotation.Status),
		ValidDate:   getValidTime(quotation.ValidDate),
		Jobs:        make([]responses.QuotationJobDetail, 0),
		Costs:       make([]responses.GeneralCostDetail, 0),
	}

	// Process jobs
	var totalLaborCost float64
	var totalMaterialCost float64

	for _, job := range jobs {

		jobDetail := responses.QuotationJobDetail{
			ID:                 job.JobID,
			Name:               job.JobName,
			Unit:               job.Unit,
			Quantity:           job.Quantity,
			LaborCost:          job.LaborCost,
			SellingPrice:       job.SellingPrice.Float64,
			TotalMaterialPrice: job.TotalMaterialPrice.Float64,
			Total:              job.Total.Float64,
			OverallCost:        job.OverallCost.Float64,
			TotalSellingPrice:  job.TotalSellingPrice.Float64,
		}

		if job.SellingPrice.Valid {
			jobDetail.SellingPrice = job.SellingPrice.Float64
		}

		// Accumulate totals
		totalLaborCost += job.LaborCost * job.Quantity
		if job.TotalMaterialPrice.Valid {
			totalMaterialCost += job.TotalMaterialPrice.Float64 * job.Quantity
		}

		response.Jobs = append(response.Jobs, jobDetail)
	}

	// Process general costs
	var totalGeneralCost float64
	for _, cost := range costs {
		if cost.EstimatedCost != nil {
			costDetail := responses.GeneralCostDetail{
				TypeName:      *cost.TypeName,
				EstimatedCost: *cost.EstimatedCost,
			}
			totalGeneralCost += *cost.EstimatedCost
			response.Costs = append(response.Costs, costDetail)
		}
	}

	return response
}

func getValidTime(nullTime sql.NullTime) time.Time {
	if nullTime.Valid {
		return nullTime.Time
	}
	return time.Time{}
}

func (u *quotationUsecase) CreateOrGetQuotation(ctx context.Context, projectID uuid.UUID) (*responses.QuotationResponse, error) {
	// Check BOQ status
	boqStatus, err := u.quotationRepo.CheckBOQStatus(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if boqStatus != "approved" {
		return nil, errors.New("BOQ must be approved before creating quotation")
	}

	// Check existing quotation
	quotation, err := u.quotationRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Create new quotation if none exists
	if quotation == nil {
		quotation, err = u.quotationRepo.Create(ctx, projectID)
		if err != nil {
			return nil, err
		}
	}

	// Get jobs and costs
	jobs, err := u.quotationRepo.GetQuotationJobs(ctx, projectID)
	if err != nil {
		return nil, err
	}

	costs, err := u.quotationRepo.GetQuotationGeneralCosts(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Build response
	response := u.buildQuotationResponse(quotation, jobs, costs)

	if len(jobs) > 0 {
		response.SellingGeneralCost = jobs[0].SellingGeneralCost.Float64
		response.TaxPercentage = jobs[0].TaxPercentage.Float64
	}
	return response, nil
}

func (u *quotationUsecase) ApproveQuotation(ctx context.Context, projectID uuid.UUID) error {
	// Validate approval conditions
	err := u.quotationRepo.ValidateApproval(ctx, projectID)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Start transaction if needed
	// If using transaction, wrap the following operations in tx

	// Approve the quotation
	err = u.quotationRepo.ApproveQuotation(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to approve quotation: %w", err)
	}

	// Get updated quotation details for response
	quotation, err := u.quotationRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get updated quotation: %w", err)
	}

	jobs, err := u.quotationRepo.GetQuotationJobs(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get quotation jobs: %w", err)
	}

	costs, err := u.quotationRepo.GetQuotationGeneralCosts(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get quotation costs: %w", err)
	}

	// Build response
	_ = u.buildQuotationResponse(quotation, jobs, costs)
	return nil
}
func (u *quotationUsecase) ExportQuotation(ctx context.Context, projectID uuid.UUID) (*responses.QuotationExportData, error) {

	boqStatus, err := u.quotationRepo.CheckBOQStatus(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if boqStatus != "approved" {
		return nil, errors.New("BOQ must be approved before exporting quotation")
	}

	quotationStatus, err := u.quotationRepo.GetQuotationStatus(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if quotationStatus != "approved" {
		return nil, errors.New("only approved quotations can be exported")
	}

	exportData, err := u.quotationRepo.GetExportData(ctx, projectID)
	if err != nil {
		return nil, err
	}

	exportData.FormatFinalAmount()

	//format job details
	for job := range exportData.JobDetails {
		jobDetail := &exportData.JobDetails[job]
		jobDetail.FormatSellingPrice()
		jobDetail.FormatAmount()
	}

	return exportData, nil
}

func (u *quotationUsecase) UpdateProjectSellingPrice(ctx context.Context, req requests.UpdateProjectSellingPriceRequest) error {

	boqStatus, err := u.quotationRepo.CheckBOQStatus(ctx, req.ProjectID)
	if err != nil {
		return err
	}

	if boqStatus != "approved" {
		return errors.New("BOQ must be approved before updating selling price")
	}

	quotationStatus, err := u.quotationRepo.GetQuotationStatus(ctx, req.ProjectID)
	if err != nil {
		return err
	}

	if quotationStatus != "draft" {
		return errors.New("can only update selling price for quotation in draft status")
	}

	if err := validateSellingPriceRequest(req); err != nil {
		return err
	}

	return u.quotationRepo.UpdateProjectSellingPrice(ctx, req)
}

func validateSellingPriceRequest(req requests.UpdateProjectSellingPriceRequest) error {
	if req.TaxPercentage <= 0 {
		return errors.New("tax percentage must be greater than 0")
	}

	if req.SellingGeneralCost <= 0 {
		return errors.New("selling general cost must be greater than 0")
	}

	if len(req.JobSellingPrices) == 0 {
		return errors.New("at least one job selling price is required")
	}

	for _, job := range req.JobSellingPrices {
		if job.SellingPrice <= 0 {
			return fmt.Errorf("selling price for job %s must be greater than 0", job.JobID)
		}
	}

	return nil
}
