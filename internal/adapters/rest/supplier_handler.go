package rest

import (
	"boonkosang/internal/requests"
	"boonkosang/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type SupplierHandler struct {
	supplierUsecase usecase.SupplierUsecase
}

func NewSupplierHandler(supplierUsecase usecase.SupplierUsecase) *SupplierHandler {
	return &SupplierHandler{
		supplierUsecase: supplierUsecase,
	}
}

func (h *SupplierHandler) SupplierRoutes(app *fiber.App) {
	supplier := app.Group("/suppliers")

	supplier.Post("/", h.Create)
	supplier.Get("/", h.List)
	supplier.Get("/:id", h.GetByID)
	supplier.Put("/:id", h.Update)
	supplier.Delete("/:id", h.Delete)
}

func (h *SupplierHandler) Create(c *fiber.Ctx) error {
	var req requests.CreateSupplierRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	supplier, err := h.supplierUsecase.Create(c.Context(), req)
	if err != nil {
		switch err.Error() {
		case "supplier with this email already exists":
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Supplier with this email already exists",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create supplier",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Supplier created successfully",
		"data":    supplier,
	})
}

func (h *SupplierHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

	response, err := h.supplierUsecase.List(c.Context(), page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve suppliers",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Suppliers retrieved successfully",
		"data":    response,
	})
}

func (h *SupplierHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid supplier ID",
		})
	}

	supplier, err := h.supplierUsecase.GetByID(c.Context(), id)
	if err != nil {
		if err.Error() == "supplier not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Supplier not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve supplier",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Supplier retrieved successfully",
		"data":    supplier,
	})
}

func (h *SupplierHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid supplier ID",
		})
	}

	var req requests.UpdateSupplierRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err = h.supplierUsecase.Update(c.Context(), id, req)
	if err != nil {
		switch err.Error() {
		case "supplier not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Supplier not found",
			})
		case "supplier with this email already exists":
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Supplier with this email already exists",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update supplier",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Supplier updated successfully",
	})
}

func (h *SupplierHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid supplier ID",
		})
	}

	err = h.supplierUsecase.Delete(c.Context(), id)
	if err != nil {
		if err.Error() == "supplier not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Supplier not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Supplier deleted successfully",
	})
}
