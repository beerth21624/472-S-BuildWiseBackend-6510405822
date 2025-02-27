package rest

import (
	"boonkosang/internal/requests"
	"boonkosang/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type InvoiceHandler struct {
	invoiceUseCase usecase.InvoiceUseCase
}

func NewInvoiceHandler(invoiceUseCase usecase.InvoiceUseCase) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceUseCase: invoiceUseCase,
	}
}

func (h *InvoiceHandler) InvoiceRoutes(app *fiber.App) {
	// Project-specific invoice routes
	invoice := app.Group("/invoices/:projectId")
	invoice.Post("/all-periods", h.CreateInvoicesForAllPeriods)
	invoice.Get("/", h.GetProjectInvoices)

	invoiceDetail := app.Group("/invoice")
	invoiceDetail.Get("/:invoiceId", h.GetInvoiceByID)
	invoiceDetail.Put("/:invoiceId/status", h.UpdateInvoiceStatus)
	invoiceDetail.Put("/:invoiceId", h.UpdateInvoice)

}

func (h *InvoiceHandler) GetProjectInvoices(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	invoices, err := h.invoiceUseCase.GetProjectInvoices(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Invoices retrieved successfully",
		"data":    invoices,
	})
}

func (h *InvoiceHandler) GetInvoiceByID(c *fiber.Ctx) error {
	invoiceID, err := uuid.Parse(c.Params("invoiceId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid invoice ID",
		})
	}

	invoice, err := h.invoiceUseCase.GetInvoiceByID(c.Context(), invoiceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Invoice retrieved successfully",
		"data":    invoice,
	})
}

func (h *InvoiceHandler) UpdateInvoiceStatus(c *fiber.Ctx) error {
	invoiceID, err := uuid.Parse(c.Params("invoiceId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid invoice ID",
		})
	}

	var req requests.UpdateInvoiceStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.invoiceUseCase.UpdateInvoiceStatus(c.Context(), invoiceID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Invoice status updated successfully",
	})
}

func (h *InvoiceHandler) CreateInvoicesForAllPeriods(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	if err := h.invoiceUseCase.CreateInvoicesForAllPeriods(c.Context(), projectID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Invoices created successfully for all periods",
	})
}

func (h *InvoiceHandler) UpdateInvoice(c *fiber.Ctx) error {
	invoiceID, err := uuid.Parse(c.Params("invoiceId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid invoice ID",
		})
	}

	var req requests.UpdateInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.invoiceUseCase.UpdateInvoice(c.Context(), invoiceID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Invoice updated successfully",
	})
}
