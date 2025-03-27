package mocks

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/requests"
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockProjectRepository is a mock implementation of the ProjectRepository interface
type MockProjectRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockProjectRepository) Create(ctx context.Context, req requests.CreateProjectRequest) (*models.Project, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

// Update mocks the Update method
func (m *MockProjectRepository) Update(ctx context.Context, id uuid.UUID, req requests.UpdateProjectRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// GetByID mocks the GetByID method
func (m *MockProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

// GetByIDWithClient mocks the GetByIDWithClient method
func (m *MockProjectRepository) GetByIDWithClient(ctx context.Context, id uuid.UUID) (*models.Project, *models.Client, error) {
	args := m.Called(ctx, id)
	var project *models.Project
	var client *models.Client

	if args.Get(0) != nil {
		project = args.Get(0).(*models.Project)
	}

	if args.Get(1) != nil {
		client = args.Get(1).(*models.Client)
	}

	return project, client, args.Error(2)
}

// List mocks the List method
func (m *MockProjectRepository) List(ctx context.Context) ([]models.Project, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Project), args.Error(1)
}

// Cancel mocks the Cancel method
func (m *MockProjectRepository) Cancel(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// GetProjectStatus mocks the GetProjectStatus method
func (m *MockProjectRepository) GetProjectStatus(ctx context.Context, projectID uuid.UUID) (*models.ProjectStatusCheck, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProjectStatusCheck), args.Error(1)
}

// ValidateStatusTransition mocks the ValidateStatusTransition method
func (m *MockProjectRepository) ValidateStatusTransition(ctx context.Context, projectID uuid.UUID, newStatus models.ProjectStatus) error {
	args := m.Called(ctx, projectID, newStatus)
	return args.Error(0)
}

// UpdateStatus mocks the UpdateStatus method
func (m *MockProjectRepository) UpdateStatus(ctx context.Context, projectID uuid.UUID, status models.ProjectStatus) error {
	args := m.Called(ctx, projectID, status)
	return args.Error(0)
}

// ValidateProjectData mocks the ValidateProjectData method
func (m *MockProjectRepository) ValidateProjectData(ctx context.Context, projectID uuid.UUID) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

// GetProjectOverview mocks the GetProjectOverview method
func (m *MockProjectRepository) GetProjectOverview(ctx context.Context, projectID uuid.UUID) (*models.ProjectOverview, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProjectOverview), args.Error(1)
}

// ValidateProjectStatus mocks the ValidateProjectStatus method
func (m *MockProjectRepository) ValidateProjectStatus(ctx context.Context, projectID uuid.UUID) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

// GetProjectSummary mocks the GetProjectSummary method
func (m *MockProjectRepository) GetProjectSummary(ctx context.Context, projectID uuid.UUID) (*models.ProjectSummary, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProjectSummary), args.Error(1)
}
