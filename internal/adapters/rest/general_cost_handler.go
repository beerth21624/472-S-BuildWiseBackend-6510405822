package rest

import (
	"boonkosang/internal/requests"
	"boonkosang/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type GeneralCostHandler struct {
	generalCostUseCase usecase.GeneralCostUseCase
}

func NewGeneralCostHandler(generalCostUseCase usecase.GeneralCostUseCase) *GeneralCostHandler {
	return &GeneralCostHandler{
		generalCostUseCase: generalCostUseCase,
	}
}

func (h *GeneralCostHandler) GeneralCostRoutes(app *fiber.App) {
	generalCost := app.Group("/general-costs")

	generalCost.Get("/project/:projectId", h.GetByProjectID)
	generalCost.Get("/types", h.GetTypes)
	generalCost.Get("/:id", h.GetByID)
	generalCost.Put("/:id/actual-cost", h.UpdateActualCost)
	generalCost.Put("/:id", h.Update)
}

func (h *GeneralCostHandler) GetByProjectID(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	generalCosts, err := h.generalCostUseCase.GetByProjectID(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get general costs",
		})
	}

	return c.JSON(fiber.Map{
		"message": "General costs retrieved successfully",
		"data":    generalCosts,
	})
}

func (h *GeneralCostHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid general cost ID",
		})
	}

	generalCost, err := h.generalCostUseCase.GetByID(c.Context(), id)
	if err != nil {
		switch err.Error() {
		case "general cost not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "General cost not found",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get general cost",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "General cost retrieved successfully",
		"data":    generalCost,
	})
}

func (h *GeneralCostHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid general cost ID",
		})
	}

	var req requests.UpdateGeneralCostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate estimated cost
	if req.EstimatedCost < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Estimated cost must be positive",
		})
	}

	err = h.generalCostUseCase.Update(c.Context(), id, req)
	if err != nil {
		switch err.Error() {
		case "general cost not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "General cost not found",
			})
		case "can only update general cost for BOQ in draft status":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "BOQ must be in draft status",
			})
		case "estimated cost must be positive":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Estimated cost must be positive",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "General cost updated successfully",
	})
}

func (h *GeneralCostHandler) GetTypes(c *fiber.Ctx) error {
	types, err := h.generalCostUseCase.GetType(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get general cost types",
		})
	}

	return c.JSON(fiber.Map{
		"message": "General cost types retrieved successfully",
		"data":    types,
	})
}

func (h *GeneralCostHandler) UpdateActualCost(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid general cost ID",
		})
	}

	var req requests.UpdateActualGeneralCostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err = h.generalCostUseCase.UpdateActualCost(c.Context(), id, req)
	if err != nil {
		switch err.Error() {
		case "general cost not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "General cost not found",
			})
		case "cannot update actual cost for completed project":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot update actual cost for completed project",
			})
		case "BOQ must be approved to update actual cost":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "BOQ must be approved to update actual cost",
			})
		case "Quotation must be approved to update actual cost":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Quotation must be approved to update actual cost",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Actual cost updated successfully",
	})
}
