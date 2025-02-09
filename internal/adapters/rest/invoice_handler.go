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
	invoice := app.Group("/invoices/:projectId")
	invoice.Post("/", h.CreateInvoice)
	invoice.Delete("/:invoiceId", h.DeleteInvoice)
	invoice.Get("/", h.GetProjectInvoices)
}

func (h *InvoiceHandler) CreateInvoice(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var req requests.CreateInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.invoiceUseCase.CreateInvoice(c.Context(), projectID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Invoice created successfully",
	})
}

func (h *InvoiceHandler) DeleteInvoice(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	invoiceID, err := uuid.Parse(c.Params("invoiceId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid invoice ID",
		})
	}

	req := requests.DeleteInvoiceRequest{
		InvoiceID: invoiceID,
	}

	if err := h.invoiceUseCase.DeleteInvoice(c.Context(), projectID, req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Invoice deleted successfully",
	})
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
