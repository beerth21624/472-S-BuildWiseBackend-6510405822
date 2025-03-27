package usecase_test

import (
	"boonkosang/internal/domain/models"
	mocks "boonkosang/internal/repositories/mock"
	"boonkosang/internal/requests"
	"boonkosang/internal/usecase"
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Common test utilities
func stringPtr(s string) *string {
	return &s
}

// InvoiceUseCaseTestSuite handles all invoice use case tests
type InvoiceUseCaseTestSuite struct {
	suite.Suite
	mockInvoiceRepo  *mocks.MockInvoiceRepository
	mockProjectRepo  *mocks.MockProjectRepository
	mockContractRepo *mocks.MockContractRepository
	uc               usecase.InvoiceUseCase
	ctx              context.Context
}

func (suite *InvoiceUseCaseTestSuite) SetupTest() {
	suite.mockInvoiceRepo = new(mocks.MockInvoiceRepository)
	suite.mockProjectRepo = new(mocks.MockProjectRepository)
	suite.mockContractRepo = new(mocks.MockContractRepository)
	suite.uc = usecase.NewInvoiceUsecase(
		suite.mockInvoiceRepo,
		suite.mockProjectRepo,
		suite.mockContractRepo,
	)
	suite.ctx = context.Background()
}

func TestInvoiceUseCaseSuite(t *testing.T) {
	suite.Run(t, new(InvoiceUseCaseTestSuite))
}

// Test CreateInvoicesForAllPeriods method
func (suite *InvoiceUseCaseTestSuite) TestCreateInvoicesForAllPeriods() {
	projectID := uuid.New()
	contractID := uuid.New()

	testCases := []struct {
		name           string
		project        *models.Project
		contract       *models.Contract
		repoErr        error
		expectedErr    string
		shouldCallRepo bool
	}{
		{
			name: "Success - Approved contract with pay within",
			project: &models.Project{
				ProjectID: projectID,
				Status:    models.ProjectStatusInProgress,
			},
			contract: &models.Contract{
				ContractID: contractID,
				ProjectID:  projectID,
				Status:     sql.NullString{String: "approved", Valid: true},
				PayWithin:  sql.NullInt32{Int32: 30, Valid: true},
			},
			shouldCallRepo: true,
			expectedErr:    "",
		},
		{
			name: "Success - Approved contract with default pay within",
			project: &models.Project{
				ProjectID: projectID,
				Status:    models.ProjectStatusInProgress,
			},
			contract: &models.Contract{
				ContractID: contractID,
				ProjectID:  projectID,
				Status:     sql.NullString{String: "approved", Valid: true},
				PayWithin:  sql.NullInt32{Valid: false},
			},
			shouldCallRepo: true,
			expectedErr:    "",
		},
		{
			name:        "Failure - Project not found",
			project:     nil,
			expectedErr: "project not found",
		},
		{
			name: "Failure - Contract not found",
			project: &models.Project{
				ProjectID: projectID,
			},
			contract:    nil,
			expectedErr: "contract not found",
		},
		{
			name: "Failure - Contract not approved",
			project: &models.Project{
				ProjectID: projectID,
			},
			contract: &models.Contract{
				ContractID: contractID,
				ProjectID:  projectID,
				Status:     sql.NullString{String: "draft", Valid: true},
			},
			expectedErr: "contract is not approved",
		},
		{
			name: "Failure - Contract project mismatch",
			project: &models.Project{
				ProjectID: projectID,
			},
			contract: &models.Contract{
				ContractID: contractID,
				ProjectID:  uuid.New(), // Different project ID
			},
			expectedErr: "contract does not belong to the specified project",
		},
		{
			name: "Failure - Repository error",
			project: &models.Project{
				ProjectID: projectID,
				Status:    models.ProjectStatusInProgress,
			},
			contract: &models.Contract{
				ContractID: contractID,
				ProjectID:  projectID,
				Status:     sql.NullString{String: "approved", Valid: true},
			},
			repoErr:        errors.New("repository error"),
			expectedErr:    "failed to create invoices",
			shouldCallRepo: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Reset mocks before each test case
			suite.SetupTest()

			// Setup mocks
			suite.mockProjectRepo.On("GetByID", suite.ctx, projectID).Return(tc.project, nil).Once()

			if tc.project != nil {
				suite.mockContractRepo.On("GetByProjectID", suite.ctx, projectID).Return(tc.contract, nil).Once()
			}

			if tc.shouldCallRepo {
				suite.mockInvoiceRepo.On("CreateForAllPeriods", suite.ctx, projectID, contractID, "").Return(tc.repoErr).Once()
			}

			// Execute
			err := suite.uc.CreateInvoicesForAllPeriods(suite.ctx, projectID)

			// Assert
			if tc.expectedErr != "" {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.expectedErr)
			} else {
				assert.NoError(suite.T(), err)
			}

			// Verify mock expectations
			suite.mockProjectRepo.AssertExpectations(suite.T())

			if tc.project != nil {
				suite.mockContractRepo.AssertExpectations(suite.T())
			}

			if tc.shouldCallRepo {
				suite.mockInvoiceRepo.AssertExpectations(suite.T())
			} else {
				suite.mockInvoiceRepo.AssertNotCalled(suite.T(), "CreateForAllPeriods")
			}
		})
	}
}

// Test UpdateInvoiceStatus method
func (suite *InvoiceUseCaseTestSuite) TestUpdateInvoiceStatus() {
	testCases := []struct {
		name          string
		setupInvoice  func() *models.Invoice
		requestStatus string
		expectedErr   string
		shouldUpdate  bool
	}{
		{
			name: "Reject approval when required fields are missing",
			setupInvoice: func() *models.Invoice {
				return &models.Invoice{
					InvoiceID: uuid.New(),
					Status:    sql.NullString{String: "draft", Valid: true},
					// Missing required fields
				}
			},
			requestStatus: "approved",
			expectedErr:   "required fields are missing",
			shouldUpdate:  false,
		},
		{
			name: "Allow approval when all required fields are filled",
			setupInvoice: func() *models.Invoice {
				return &models.Invoice{
					InvoiceID:      uuid.New(),
					Status:         sql.NullString{String: "draft", Valid: true},
					InvoiceDate:    sql.NullTime{Time: time.Now(), Valid: true},
					PaymentDueDate: sql.NullTime{Time: time.Now().AddDate(0, 0, 30), Valid: true},
					PaymentTerm:    sql.NullString{String: "NET30", Valid: true},
				}
			},
			requestStatus: "approved",
			expectedErr:   "",
			shouldUpdate:  true,
		},
		{
			name: "Prevent reverting from approved to draft",
			setupInvoice: func() *models.Invoice {
				return &models.Invoice{
					InvoiceID: uuid.New(),
					Status:    sql.NullString{String: "approved", Valid: true},
				}
			},
			requestStatus: "draft",
			expectedErr:   "cannot change status from approved to draft",
			shouldUpdate:  false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Reset mocks before each test case
			suite.SetupTest()

			// Setup
			invoice := tc.setupInvoice()
			invoiceID := invoice.InvoiceID

			suite.mockInvoiceRepo.On("GetByID", suite.ctx, invoiceID).Return(invoice, nil).Once()

			if tc.shouldUpdate {
				suite.mockInvoiceRepo.On("UpdateStatus", suite.ctx, invoiceID, tc.requestStatus).Return(nil).Once()
			}

			req := requests.UpdateInvoiceStatusRequest{Status: tc.requestStatus}

			// Execute
			err := suite.uc.UpdateInvoiceStatus(suite.ctx, invoiceID, req)

			// Assert
			if tc.expectedErr != "" {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), tc.expectedErr)
				if !tc.shouldUpdate {
					suite.mockInvoiceRepo.AssertNotCalled(suite.T(), "UpdateStatus")
				}
			} else {
				assert.NoError(suite.T(), err)
				suite.mockInvoiceRepo.AssertExpectations(suite.T())
			}
		})
	}
}

// Test UpdateInvoice method
func (suite *InvoiceUseCaseTestSuite) TestUpdateInvoice() {
	testCases := []struct {
		name         string
		setupInvoice func() *models.Invoice
		request      requests.UpdateInvoiceRequest
		expectedErr  string
		shouldUpdate bool
	}{
		{
			name: "Allow editing draft invoice",
			setupInvoice: func() *models.Invoice {
				return &models.Invoice{
					InvoiceID: uuid.New(),
					Status:    sql.NullString{String: "draft", Valid: true},
				}
			},
			request: requests.UpdateInvoiceRequest{
				InvoiceDate:    stringPtr("2023-01-01"),
				PaymentDueDate: stringPtr("2023-01-31"),
				PaymentTerm:    stringPtr("NET30"),
				Remarks:        stringPtr("Updated remarks"),
			},
			expectedErr:  "",
			shouldUpdate: true,
		},
		{
			name: "Prevent editing approved invoice",
			setupInvoice: func() *models.Invoice {
				return &models.Invoice{
					InvoiceID: uuid.New(),
					Status:    sql.NullString{String: "approved", Valid: true},
				}
			},
			request: requests.UpdateInvoiceRequest{
				InvoiceDate: stringPtr("2023-01-01"),
			},
			expectedErr:  "cannot edit approved invoice",
			shouldUpdate: false,
		},
		{
			name: "Invoice not found",
			setupInvoice: func() *models.Invoice {
				return nil
			},
			request:      requests.UpdateInvoiceRequest{},
			expectedErr:  "invoice not found",
			shouldUpdate: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Reset mocks before each test case
			suite.SetupTest()

			// Setup
			invoice := tc.setupInvoice()
			var invoiceID uuid.UUID
			if invoice != nil {
				invoiceID = invoice.InvoiceID
			} else {
				invoiceID = uuid.New()
			}

			suite.mockInvoiceRepo.On("GetByID", suite.ctx, invoiceID).Return(invoice, nil).Once()

			if tc.shouldUpdate {
				suite.mockInvoiceRepo.On("Update", suite.ctx, invoiceID, mock.Anything).Return(nil).Once()
			}

			// Execute
			err := suite.uc.UpdateInvoice(suite.ctx, invoiceID, tc.request)

			// Assert
			if tc.expectedErr != "" {
				assert.Error(suite.T(), err)
				assert.Equal(suite.T(), tc.expectedErr, err.Error())
				if !tc.shouldUpdate {
					suite.mockInvoiceRepo.AssertNotCalled(suite.T(), "Update")
				}
			} else {
				assert.NoError(suite.T(), err)
				suite.mockInvoiceRepo.AssertExpectations(suite.T())
			}
		})
	}
}
