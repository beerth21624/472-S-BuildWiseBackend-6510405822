package usecase

import (
	"boonkosang/internal/repositories"
	"boonkosang/internal/requests"
	"boonkosang/internal/responses"
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
)

type InvoiceUseCase interface {
	CreateInvoice(ctx context.Context, projectID uuid.UUID, req requests.CreateInvoiceRequest) error
	DeleteInvoice(ctx context.Context, projectID uuid.UUID, req requests.DeleteInvoiceRequest) error
	GetProjectInvoices(ctx context.Context, projectID uuid.UUID) ([]responses.InvoiceResponse, error)
}

type invoiceUseCase struct {
	invoiceRepo repositories.InvoiceRepository
	projectRepo repositories.ProjectRepository
}

func NewInvoiceUsecase(
	invoiceRepo repositories.InvoiceRepository,
	projectRepo repositories.ProjectRepository,
) InvoiceUseCase {
	return &invoiceUseCase{
		invoiceRepo: invoiceRepo,
		projectRepo: projectRepo,
	}
}

func (u *invoiceUseCase) CreateInvoice(ctx context.Context, projectID uuid.UUID, req requests.CreateInvoiceRequest) error {
	// Validate project exists
	project, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return errors.New("project not found")
	}

	// Validate file URL
	if _, err := url.Parse(req.FileURL); err != nil {
		return errors.New("invalid file URL")
	}

	// Create invoice
	err = u.invoiceRepo.Create(ctx, projectID, req.FileURL)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	return nil
}

func (u *invoiceUseCase) DeleteInvoice(ctx context.Context, projectID uuid.UUID, req requests.DeleteInvoiceRequest) error {
	// Get invoice to verify it exists and belongs to the project
	invoice, err := u.invoiceRepo.GetByID(ctx, req.InvoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return errors.New("invoice not found")
	}
	if invoice.ProjectID != projectID {
		return errors.New("invoice does not belong to the specified project")
	}

	// Validate project status
	err = u.invoiceRepo.ValidateProjectStatus(ctx, projectID)
	if err != nil {
		return err
	}

	// Delete invoice
	err = u.invoiceRepo.Delete(ctx, req.InvoiceID)
	if err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	return nil
}

func (u *invoiceUseCase) GetProjectInvoices(ctx context.Context, projectID uuid.UUID) ([]responses.InvoiceResponse, error) {
	invoices, err := u.invoiceRepo.GetByProjectID(ctx, projectID)

	if err != nil {
		return nil, fmt.Errorf("failed to get project invoices: %w", err)
	}

	if invoices == nil {
		return nil, errors.New("invoices not found")
	}

	var response []responses.InvoiceResponse
	for _, invoice := range invoices {
		response = append(response, responses.InvoiceResponse{
			InvoiceID: invoice.InvoiceID,
			ProjectID: invoice.ProjectID,
			FileURL:   invoice.FileURL.String,
			CreatedAt: invoice.CreatedAt,
			UpdatedAt: invoice.UpdatedAt.Time,
		})
	}

	return response, nil
}
