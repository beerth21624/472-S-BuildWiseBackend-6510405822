package repositories

import (
	"boonkosang/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

type InvoiceRepository interface {
	Create(ctx context.Context, projectID uuid.UUID, fileURL string) error
	Delete(ctx context.Context, invoiceID uuid.UUID) error
	GetByID(ctx context.Context, invoiceID uuid.UUID) (*models.Invoice, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.Invoice, error)
	ValidateProjectStatus(ctx context.Context, projectID uuid.UUID) error
}
