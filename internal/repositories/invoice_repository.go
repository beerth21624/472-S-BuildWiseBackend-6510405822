package repositories

import (
	"boonkosang/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

type InvoiceRepository interface {
	CreateForAllPeriods(ctx context.Context, projectID uuid.UUID, contractID uuid.UUID, paymentTerm string) error
	GetByID(ctx context.Context, invoiceID uuid.UUID) (*models.Invoice, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.Invoice, error)
	UpdateStatus(ctx context.Context, invoiceID uuid.UUID, status string) error
	Update(ctx context.Context, invoiceID uuid.UUID, updates map[string]interface{}) error // New method
}
