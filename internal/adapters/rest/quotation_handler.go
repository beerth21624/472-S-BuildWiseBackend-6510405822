// rest/quotation_handler.go
package rest

import (
	"boonkosang/internal/requests"
	"boonkosang/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type QuotationHandler struct {
	quotationUsecase usecase.QuotationUsecase
}

func NewQuotationHandler(quotationUsecase usecase.QuotationUsecase) *QuotationHandler {
	return &QuotationHandler{
		quotationUsecase: quotationUsecase,
	}
}

func (h *QuotationHandler) QuotationRoutes(app *fiber.App) {
	quotation := app.Group("/quotations")

	// Create or Get Quotation
	quotation.Get("/projects/:projectId/export", h.ExportQuotation)

	quotation.Put("/projects/:projectId/selling-price", h.UpdateProjectSellingPrice)

	quotation.Post("/projects/:projectId", h.CreateOrGetQuotation)
	quotation.Put("/projects/:projectId/approve", h.ApproveQuotation)

}

// CreateOrGetQuotation creates a new quotation or retrieves existing one
func (h *QuotationHandler) CreateOrGetQuotation(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID format",
		})
	}

	response, err := h.quotationUsecase.CreateOrGetQuotation(c.Context(), projectID)
	if err != nil {
		switch err.Error() {
		case "BOQ must be approved before creating quotation":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		case "BOQ not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "BOQ not found for this project",
			})
		case "project not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Project not found",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to process quotation",
				"details": err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Quotation processed successfully",
		"data":    response,
	})
}

func (h *QuotationHandler) ApproveQuotation(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID format",
		})
	}

	err = h.quotationUsecase.ApproveQuotation(c.Context(), projectID)
	if err != nil {
		switch err.Error() {
		case "BOQ not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "BOQ not found",
			})
		case "quotation not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Quotation not found",
			})
		case "BOQ must be approved before approving quotation":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		case "only draft quotations can be approved":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		case "no draft quotation found to approve":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No draft quotation found to approve",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to approve quotation",
				"details": err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Quotation approved successfully",
	})
}

func (h *QuotationHandler) ExportQuotation(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID format",
		})
	}

	exportData, err := h.quotationUsecase.ExportQuotation(c.Context(), projectID)
	if err != nil {
		switch err.Error() {
		case "BOQ must be approved before exporting quotation":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		case "BOQ not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "BOQ not found",
			})
		case "quotation not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Quotation not found",
			})
		case "only approved quotations can be exported":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to export quotation",
				"details": err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Quotation exported successfully",
		"data":    exportData,
	})
}

func (h *QuotationHandler) UpdateProjectSellingPrice(c *fiber.Ctx) error {
	var req requests.UpdateProjectSellingPriceRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Parse project ID from URL
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}
	req.ProjectID = projectID

	err = h.quotationUsecase.UpdateProjectSellingPrice(c.Context(), req)
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Project selling prices updated successfully",
	})
}
