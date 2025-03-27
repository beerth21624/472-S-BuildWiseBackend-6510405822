package mocks

import (
	"boonkosang/internal/domain/models"
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockInvoiceRepository is a mock implementation of the InvoiceRepository interface
type MockInvoiceRepository struct {
	mock.Mock
}

// ValidateProjectStatus mocks the ValidateProjectStatus method
func (m *MockInvoiceRepository) ValidateProjectStatus(ctx context.Context, projectID uuid.UUID) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

// CreateForAllPeriods mocks the CreateForAllPeriods method
func (m *MockInvoiceRepository) CreateForAllPeriods(ctx context.Context, projectID uuid.UUID, contractID uuid.UUID, paymentTerm string) error {
	args := m.Called(ctx, projectID, contractID, paymentTerm)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *MockInvoiceRepository) GetByID(ctx context.Context, invoiceID uuid.UUID) (*models.Invoice, error) {
	args := m.Called(ctx, invoiceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Invoice), args.Error(1)
}

// GetByProjectID mocks the GetByProjectID method
func (m *MockInvoiceRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]models.Invoice, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Invoice), args.Error(1)
}

// UpdateStatus mocks the UpdateStatus method
func (m *MockInvoiceRepository) UpdateStatus(ctx context.Context, invoiceID uuid.UUID, status string) error {
	args := m.Called(ctx, invoiceID, status)
	return args.Error(0)
}

// Update mocks the Update method
func (m *MockInvoiceRepository) Update(ctx context.Context, invoiceID uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(ctx, invoiceID, updates)
	return args.Error(0)
}
