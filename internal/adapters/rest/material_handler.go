// rest/material_handler.go
package rest

import (
	"boonkosang/internal/requests"
	"boonkosang/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MaterialHandler struct {
	materialUsecase usecase.MaterialUsecase
}

func NewMaterialHandler(materialUsecase usecase.MaterialUsecase) *MaterialHandler {
	return &MaterialHandler{
		materialUsecase: materialUsecase,
	}
}

func (h *MaterialHandler) MaterialRoutes(app *fiber.App) {
	material := app.Group("/materials")

	material.Post("/", h.Create)
	material.Get("/", h.List)

	material.Get("/:projectId/prices", h.GetMaterialPrices)
	material.Put("/:boqId/estimated-price", h.UpdateEstimatedPrice)
	material.Put("/:boqId/actual-price", h.UpdateActualPrice)

	material.Get("/:id", h.GetByID)
	material.Put("/:id", h.Update)
	material.Delete("/:id", h.Delete)

}

func (h *MaterialHandler) Create(c *fiber.Ctx) error {
	var req requests.CreateMaterialRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	material, err := h.materialUsecase.Create(c.Context(), req)
	if err != nil {
		switch err.Error() {
		case "material ID already exists":
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create material",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Material created successfully",
		"data":    material,
	})
}

func (h *MaterialHandler) List(c *fiber.Ctx) error {

	response, err := h.materialUsecase.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve materials",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Materials retrieved successfully",
		"data":    response,
	})
}

func (h *MaterialHandler) GetByID(c *fiber.Ctx) error {
	materialID := c.Params("id")
	if materialID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Material ID is required",
		})
	}

	material, err := h.materialUsecase.GetByID(c.Context(), materialID)
	if err != nil {
		switch err.Error() {
		case "material not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Material not found",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve material",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Material retrieved successfully",
		"data":    material,
	})
}

func (h *MaterialHandler) Update(c *fiber.Ctx) error {
	materialID := c.Params("id")
	if materialID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Material ID is required",
		})
	}

	var req requests.UpdateMaterialRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" || req.Unit == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields",
		})
	}

	err := h.materialUsecase.Update(c.Context(), materialID, req)
	if err != nil {
		switch err.Error() {
		case "material not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Material not found",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update material",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Material updated successfully",
	})
}

func (h *MaterialHandler) Delete(c *fiber.Ctx) error {
	materialID := c.Params("id")
	if materialID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Material ID is required",
		})
	}

	err := h.materialUsecase.Delete(c.Context(), materialID)
	if err != nil {
		switch err.Error() {
		case "material not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Material not found",
			})
		case "material is in use and cannot be deleted":
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Material deleted successfully",
	})
}

func (h *MaterialHandler) GetMaterialPrices(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	response, err := h.materialUsecase.GetMaterialPrices(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Material prices retrieved successfully",
		"data":    response,
	})
}

func (h *MaterialHandler) UpdateEstimatedPrice(c *fiber.Ctx) error {
	boqID, err := uuid.Parse(c.Params("boqId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid BOQ ID",
		})
	}

	var req requests.UpdateMaterialEstimatedPriceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.materialUsecase.UpdateEstimatedPrice(c.Context(), boqID, req); err != nil {
		switch err.Error() {
		case "can only update estimated prices for BOQ in draft status":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		case "estimated price must be greater than 0":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update estimated price",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Estimated price updated successfully",
	})
}

func (h *MaterialHandler) UpdateActualPrice(c *fiber.Ctx) error {
	boqID, err := uuid.Parse(c.Params("boqId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid BOQ ID",
		})
	}

	var req requests.UpdateMaterialActualPriceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.materialUsecase.UpdateActualPrice(c.Context(), boqID, req); err != nil {
		switch err.Error() {
		case "can only update actual prices for approved BOQ":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		case "cannot update actual prices for completed projects":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		case "can only update actual prices when quotation is approved":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		case "actual price must be greater than 0":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Actual price updated successfully",
	})
}
