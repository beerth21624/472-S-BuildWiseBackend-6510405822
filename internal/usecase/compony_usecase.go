// usecase/company_usecase.go
package usecase

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"boonkosang/internal/requests"
	"boonkosang/internal/responses"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type CompanyUseCase interface {
	GetCompanyByUserID(ctx context.Context, userID uuid.UUID) (*responses.CompanyResponse, error)
	UpdateCompany(ctx context.Context, userID uuid.UUID, req requests.UpdateCompanyRequest) (*responses.CompanyResponse, error)
}

type companyUseCase struct {
	companyRepo repositories.CompanyRepository
}

func NewCompanyUsecase(companyRepo repositories.CompanyRepository) CompanyUseCase {
	return &companyUseCase{
		companyRepo: companyRepo,
	}
}

// GetCompanyByUserID retrieves or creates a company for a user
func (u *companyUseCase) GetCompanyByUserID(ctx context.Context, userID uuid.UUID) (*responses.CompanyResponse, error) {
	company, err := u.companyRepo.GetOrCreateCompanyByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create company: %w", err)
	}

	response, err := u.createCompanyResponse(company)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// UpdateCompany updates company information
func (u *companyUseCase) UpdateCompany(ctx context.Context, userID uuid.UUID, req requests.UpdateCompanyRequest) (*responses.CompanyResponse, error) {
	// Get existing company
	existingCompany, err := u.companyRepo.GetOrCreateCompanyByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	// Convert address to JSON
	addressJSON, err := json.Marshal(req.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal address: %w", err)
	}

	// Update company
	updatedCompany := &models.Company{
		CompanyID: existingCompany.CompanyID,
		Name:      req.Name,
		Email:     req.Email,
		Tel:       req.Tel,
		Address:   addressJSON,
		TaxID:     req.TaxID,
	}

	// Update in repository
	if err := u.companyRepo.UpdateCompany(ctx, updatedCompany); err != nil {
		return nil, fmt.Errorf("failed to update company: %w", err)
	}

	// Get updated company
	company, err := u.companyRepo.GetOrCreateCompanyByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated company: %w", err)
	}

	response, err := u.createCompanyResponse(company)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// createCompanyResponse creates a standardized company response
func (u *companyUseCase) createCompanyResponse(company *models.Company) (*responses.CompanyResponse, error) {

	return &responses.CompanyResponse{
		CompanyID: company.CompanyID,
		Name:      company.Name,
		Email:     company.Email,
		Tel:       company.Tel,
		Address:   company.Address,
		TaxID:     company.TaxID,
		IsNew:     company.TaxID == "",
	}, nil
}
