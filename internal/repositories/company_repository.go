package repositories

import (
	"boonkosang/internal/domain/models"
	"context"

	"github.com/google/uuid"
)

type CompanyRepository interface {
	GetOrCreateCompanyByUserID(ctx context.Context, userID uuid.UUID) (*models.Company, error)
	UpdateCompany(ctx context.Context, company *models.Company) error
}
