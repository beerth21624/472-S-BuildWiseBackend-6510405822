package mocks

import (
	"boonkosang/internal/domain/models"
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockContractRepository struct {
	mock.Mock
}

func (m *MockContractRepository) Create(ctx context.Context, contract *models.Contract) error {
	args := m.Called(ctx, contract)
	return args.Error(0)
}

func (m *MockContractRepository) Update(ctx context.Context, contract *models.Contract) error {
	args := m.Called(ctx, contract)
	return args.Error(0)
}

func (m *MockContractRepository) Delete(ctx context.Context, projectID uuid.UUID) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

func (m *MockContractRepository) GetByID(ctx context.Context, projectID uuid.UUID) (*models.Contract, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Contract), args.Error(1)
}

func (m *MockContractRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) (*models.Contract, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Contract), args.Error(1)
}

func (m *MockContractRepository) ValidateProjectStatus(ctx context.Context, projectID uuid.UUID) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

func (m *MockContractRepository) ChangeStatus(ctx context.Context, projectID uuid.UUID, status string) error {
	args := m.Called(ctx, projectID, status)
	return args.Error(0)
}
