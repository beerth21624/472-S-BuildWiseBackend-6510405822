package usecase

import (
	"boonkosang/internal/repositories"
	"boonkosang/internal/requests"
	"boonkosang/internal/responses"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type InvoiceUseCase interface {
	GetProjectInvoices(ctx context.Context, projectID uuid.UUID) ([]responses.InvoiceResponse, error)
	GetInvoiceByID(ctx context.Context, invoiceID uuid.UUID) (*responses.InvoiceResponse, error)
	UpdateInvoiceStatus(ctx context.Context, invoiceID uuid.UUID, req requests.UpdateInvoiceStatusRequest) error
	CreateInvoicesForAllPeriods(ctx context.Context, projectID uuid.UUID) error
	UpdateInvoice(ctx context.Context, invoiceID uuid.UUID, req requests.UpdateInvoiceRequest) error // New method

}

type invoiceUseCase struct {
	invoiceRepo  repositories.InvoiceRepository
	projectRepo  repositories.ProjectRepository
	contractRepo repositories.ContractRepository
}

func NewInvoiceUsecase(
	invoiceRepo repositories.InvoiceRepository,
	projectRepo repositories.ProjectRepository,
	contractRepo repositories.ContractRepository,
) InvoiceUseCase {
	return &invoiceUseCase{
		invoiceRepo:  invoiceRepo,
		projectRepo:  projectRepo,
		contractRepo: contractRepo,
	}
}

func (u *invoiceUseCase) GetProjectInvoices(ctx context.Context, projectID uuid.UUID) ([]responses.InvoiceResponse, error) {
	project, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project not found")
	}

	invoices, err := u.invoiceRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project invoices: %w", err)
	}

	var responseList []responses.InvoiceResponse
	for _, invoice := range invoices {
		response := responses.InvoiceResponse{
			InvoiceID: invoice.InvoiceID,
			ProjectID: invoice.ProjectID,
			PeriodID:  invoice.PeriodID,
			Status:    invoice.Status.String,
			CreatedAt: invoice.CreatedAt,
			Period: responses.PeriodResponse{
				PeriodID:        invoice.Period.PeriodID,
				PeriodNumber:    invoice.Period.PeriodNumber,
				AmountPeriod:    invoice.Period.AmountPeriod,
				DeliveredWithin: invoice.Period.DeliveredWithin,
			},
		}

		if invoice.PaymentTerm.Valid {
			response.PaymentTerm = invoice.PaymentTerm.String
		}

		if invoice.InvoiceDate.Valid {
			response.InvoiceDate = invoice.InvoiceDate.Time
		}

		if invoice.PaymentDueDate.Valid {
			response.PaymentDueDate = invoice.PaymentDueDate.Time
		}

		if invoice.PaidDate.Valid {
			response.PaidDate = invoice.PaidDate.Time
		}

		if invoice.Remarks.Valid {
			response.Remarks = invoice.Remarks.String
		}

		if invoice.UpdatedAt.Valid {
			response.UpdatedAt = invoice.UpdatedAt.Time
		}

		responseList = append(responseList, response)
	}

	return responseList, nil
}

func (u *invoiceUseCase) GetInvoiceByID(ctx context.Context, invoiceID uuid.UUID) (*responses.InvoiceResponse, error) {
	invoice, err := u.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	if invoice == nil {
		return nil, errors.New("invoice not found")
	}

	response := &responses.InvoiceResponse{
		InvoiceID: invoice.InvoiceID,
		ProjectID: invoice.ProjectID,
		PeriodID:  invoice.PeriodID,
		Status:    invoice.Status.String,
		CreatedAt: invoice.CreatedAt,
		Period: responses.PeriodResponse{
			PeriodID:        invoice.Period.PeriodID,
			PeriodNumber:    invoice.Period.PeriodNumber,
			AmountPeriod:    invoice.Period.AmountPeriod,
			DeliveredWithin: invoice.Period.DeliveredWithin,
		},
	}

	if invoice.PaymentTerm.Valid {
		response.PaymentTerm = invoice.PaymentTerm.String
	}

	if invoice.InvoiceDate.Valid {
		response.InvoiceDate = invoice.InvoiceDate.Time
	}

	if invoice.PaymentDueDate.Valid {
		response.PaymentDueDate = invoice.PaymentDueDate.Time
	}

	if invoice.PaidDate.Valid {
		response.PaidDate = invoice.PaidDate.Time
	}

	if invoice.Remarks.Valid {
		response.Remarks = invoice.Remarks.String
	}

	if invoice.Retention.Valid {
		response.Retention = invoice.Retention.Float64
	}

	if invoice.UpdatedAt.Valid {
		response.UpdatedAt = invoice.UpdatedAt.Time
	}

	return response, nil
}

func (u *invoiceUseCase) UpdateInvoiceStatus(ctx context.Context, invoiceID uuid.UUID, req requests.UpdateInvoiceStatusRequest) error {
	invoice, err := u.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	if invoice == nil {
		return errors.New("invoice not found")
	}

	if invoice.Status.Valid && invoice.Status.String == "approved" && req.Status == "draft" {
		return errors.New("cannot change status from approved to draft")
	}

	err = u.invoiceRepo.UpdateStatus(ctx, invoiceID, req.Status)
	if err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	return nil
}

func (u *invoiceUseCase) CreateInvoicesForAllPeriods(ctx context.Context, projectID uuid.UUID) error {
	project, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return errors.New("project not found")
	}
	contract, err := u.contractRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get contract: %w", err)
	}
	if contract == nil {
		return errors.New("contract not found")
	}
	if contract.ProjectID != projectID {
		return errors.New("contract does not belong to the specified project")
	}

	if !contract.PayWithin.Valid {
		return errors.New("contract PayWithin is not valid")
	}
	err = u.invoiceRepo.CreateForAllPeriods(ctx, projectID, contract.ContractID, "")
	if err != nil {
		return fmt.Errorf("failed to create invoices: %w", err)
	}

	return nil
}

func (u *invoiceUseCase) UpdateInvoice(ctx context.Context, invoiceID uuid.UUID, req requests.UpdateInvoiceRequest) error {
	invoice, err := u.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	if invoice == nil {
		return errors.New("invoice not found")
	}

	updates := make(map[string]interface{})

	if req.InvoiceDate != nil {
		date, err := time.Parse("2006-01-02", *req.InvoiceDate)
		if err != nil {
			return fmt.Errorf("invalid invoice date format: %w", err)
		}
		updates["invoice_date"] = date
	}

	if req.PaymentDueDate != nil {
		date, err := time.Parse("2006-01-02", *req.PaymentDueDate)
		if err != nil {
			return fmt.Errorf("invalid payment due date format: %w", err)
		}
		updates["payment_due_date"] = date
	}

	if req.PaidDate != nil {
		date, err := time.Parse("2006-01-02", *req.PaidDate)
		if err != nil {
			return fmt.Errorf("invalid paid date format: %w", err)
		}
		updates["paid_date"] = date
	}

	if req.PaymentTerm != nil {
		updates["payment_term"] = *req.PaymentTerm
	}

	if req.Remarks != nil {
		updates["remarks"] = *req.Remarks
	}

	if req.Retention != nil {
		if *req.Retention < 0 || *req.Retention > 100 {
			return fmt.Errorf("retention must be between 0 and 100")
		}
		updates["retention"] = *req.Retention
	}

	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	// Update the invoice
	err = u.invoiceRepo.Update(ctx, invoiceID, updates)
	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	return nil
}
