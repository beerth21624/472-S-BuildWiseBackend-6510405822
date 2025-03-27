package postgres_test

import (
	"boonkosang/internal/adapters/postgres"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestInvoiceRepository(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := postgres.NewInvoiceRepository(sqlxDB)

	t.Run("ValidateProjectStatus", func(t *testing.T) {
		projectID := uuid.New()

		t.Run("Success - Valid project status", func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"project_status", "boq_status", "quotation_status", "contract_id"}).
				AddRow("active", "approved", "approved", uuid.New())

			mock.ExpectQuery(`SELECT`).
				WithArgs(projectID).
				WillReturnRows(rows)

			err := repo.ValidateProjectStatus(context.Background(), projectID)
			assert.NoError(t, err)
		})

		t.Run("Failure - Project completed", func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"project_status", "boq_status", "quotation_status", "contract_id"}).
				AddRow("completed", "approved", "approved", uuid.New())

			mock.ExpectQuery(`SELECT`).
				WithArgs(projectID).
				WillReturnRows(rows)

			err := repo.ValidateProjectStatus(context.Background(), projectID)
			assert.EqualError(t, err, "project is already completed")
		})

		t.Run("Failure - BOQ not approved", func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"project_status", "boq_status", "quotation_status", "contract_id"}).
				AddRow("active", "pending", "approved", uuid.New())

			mock.ExpectQuery(`SELECT`).
				WithArgs(projectID).
				WillReturnRows(rows)

			err := repo.ValidateProjectStatus(context.Background(), projectID)
			assert.EqualError(t, err, "BOQ must be approved")
		})

		t.Run("Failure - Quotation not approved", func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"project_status", "boq_status", "quotation_status", "contract_id"}).
				AddRow("active", "approved", "pending", uuid.New())

			mock.ExpectQuery(`SELECT`).
				WithArgs(projectID).
				WillReturnRows(rows)

			err := repo.ValidateProjectStatus(context.Background(), projectID)
			assert.EqualError(t, err, "quotation must be approved")
		})

		t.Run("Failure - No contract", func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"project_status", "boq_status", "quotation_status", "contract_id"}).
				AddRow("active", "approved", "approved", nil)

			mock.ExpectQuery(`SELECT`).
				WithArgs(projectID).
				WillReturnRows(rows)

			err := repo.ValidateProjectStatus(context.Background(), projectID)
			assert.EqualError(t, err, "contract must exist for the project")
		})
	})

	t.Run("CreateForAllPeriods", func(t *testing.T) {
		projectID := uuid.New()
		contractID := uuid.New()

		t.Run("Success - Create invoices for all periods", func(t *testing.T) {
			// Setup validation mocks
			validationRows := sqlmock.NewRows([]string{"project_status", "boq_status", "quotation_status", "contract_id"}).
				AddRow("active", "approved", "approved", contractID)
			mock.ExpectQuery(`SELECT`).WithArgs(projectID).WillReturnRows(validationRows)

			// Contract validation
			mock.ExpectQuery(`SELECT c.contract_id`).
				WithArgs(projectID, contractID).
				WillReturnRows(sqlmock.NewRows([]string{"contract_id"}).AddRow(contractID))

			// Periods query
			periodRows := sqlmock.NewRows([]string{"period_id", "period_number", "pay_within"}).
				AddRow(uuid.New(), 1, 30).
				AddRow(uuid.New(), 2, 30)
			mock.ExpectQuery(`SELECT p.period_id`).
				WithArgs(contractID).
				WillReturnRows(periodRows)

			// Expect transaction begin
			mock.ExpectBegin()

			// Expect two invoice inserts
			mock.ExpectExec(`INSERT INTO invoice`).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec(`INSERT INTO invoice`).WillReturnResult(sqlmock.NewResult(1, 1))

			// Expect transaction commit
			mock.ExpectCommit()

			err := repo.CreateForAllPeriods(context.Background(), projectID, contractID, "NET30")
			assert.NoError(t, err)
		})

		t.Run("Failure - No available periods", func(t *testing.T) {
			// Setup validation mocks
			validationRows := sqlmock.NewRows([]string{"project_status", "boq_status", "quotation_status", "contract_id"}).
				AddRow("active", "approved", "approved", contractID)
			mock.ExpectQuery(`SELECT`).WithArgs(projectID).WillReturnRows(validationRows)

			// Contract validation
			mock.ExpectQuery(`SELECT c.contract_id`).
				WithArgs(projectID, contractID).
				WillReturnRows(sqlmock.NewRows([]string{"contract_id"}).AddRow(contractID))

			// Empty periods
			mock.ExpectQuery(`SELECT p.period_id`).
				WithArgs(contractID).
				WillReturnRows(sqlmock.NewRows([]string{"period_id", "period_number", "pay_within"}))

			err := repo.CreateForAllPeriods(context.Background(), projectID, contractID, "NET30")
			assert.EqualError(t, err, "no available periods found for invoicing in this contract")
		})
	})

	t.Run("GetByID", func(t *testing.T) {
		invoiceID := uuid.New()

		t.Run("Success - Found invoice", func(t *testing.T) {
			invoiceRow := sqlmock.NewRows([]string{"invoice_id", "project_id", "period_id", "status"}).
				AddRow(invoiceID, uuid.New(), uuid.New(), "draft")
			mock.ExpectQuery(`SELECT \* FROM invoice`).
				WithArgs(invoiceID).
				WillReturnRows(invoiceRow)

			// Expect period query
			periodRow := sqlmock.NewRows([]string{"period_id", "period_number"}).
				AddRow(uuid.New(), 1)
			mock.ExpectQuery(`SELECT \* FROM period`).
				WillReturnRows(periodRow)

			invoice, err := repo.GetByID(context.Background(), invoiceID)
			assert.NoError(t, err)
			assert.NotNil(t, invoice)
			assert.Equal(t, "draft", invoice.Status.String)
		})

		t.Run("Success - Not found", func(t *testing.T) {
			mock.ExpectQuery(`SELECT \* FROM invoice`).
				WithArgs(invoiceID).
				WillReturnError(sql.ErrNoRows)

			invoice, err := repo.GetByID(context.Background(), invoiceID)
			assert.NoError(t, err)
			assert.Nil(t, invoice)
		})
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		invoiceID := uuid.New()

		t.Run("Success - Update status", func(t *testing.T) {
			mock.ExpectExec(`UPDATE invoice`).
				WithArgs("approved", invoiceID).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := repo.UpdateStatus(context.Background(), invoiceID, "approved")
			assert.NoError(t, err)
		})

		t.Run("Failure - Invoice not found", func(t *testing.T) {
			mock.ExpectExec(`UPDATE invoice`).
				WithArgs("approved", invoiceID).
				WillReturnResult(sqlmock.NewResult(0, 0))

			err := repo.UpdateStatus(context.Background(), invoiceID, "approved")
			assert.EqualError(t, err, "invoice not found")
		})
	})

	t.Run("Update", func(t *testing.T) {
		invoiceID := uuid.New()

		t.Run("Success - Update fields", func(t *testing.T) {
			// Create a fixed time for testing
			testTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

			// We need to match the exact SQL query with expected parameters
			mock.ExpectExec(`UPDATE invoice SET invoice_date = \$1, payment_term = \$2, updated_at = \$3 WHERE invoice_id = \$4`).
				WithArgs(testTime, "NET30", sqlmock.AnyArg(), invoiceID).
				WillReturnResult(sqlmock.NewResult(1, 1))

			updates := map[string]interface{}{
				"invoice_date": testTime,
				"payment_term": "NET30",
			}

			err := repo.Update(context.Background(), invoiceID, updates)
			assert.NoError(t, err)
		})

		t.Run("Failure - No fields to update", func(t *testing.T) {
			err := repo.Update(context.Background(), invoiceID, map[string]interface{}{})
			assert.EqualError(t, err, "no fields to update")
		})
	})
	// Make sure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
