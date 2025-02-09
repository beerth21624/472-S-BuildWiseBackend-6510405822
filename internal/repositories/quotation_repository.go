package repositories

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/requests"
	"boonkosang/internal/responses"
	"context"

	"github.com/google/uuid"
)

type QuotationRepository interface {
	Create(ctx context.Context, projectID uuid.UUID) (*models.Quotation, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) (*models.Quotation, error)
	GetQuotationJobs(ctx context.Context, projectID uuid.UUID) ([]models.QuotationJob, error)
	GetQuotationGeneralCosts(ctx context.Context, projectID uuid.UUID) ([]models.QuotationGeneralCost, error)
	CheckBOQStatus(ctx context.Context, projectID uuid.UUID) (string, error)

	ApproveQuotation(ctx context.Context, projectID uuid.UUID) error
	GetQuotationStatus(ctx context.Context, projectID uuid.UUID) (string, error)
	ValidateApproval(ctx context.Context, projectID uuid.UUID) error

	GetExportData(ctx context.Context, projectID uuid.UUID) (*responses.QuotationExportData, error)

	UpdateProjectSellingPrice(ctx context.Context, req requests.UpdateProjectSellingPriceRequest) error
}
