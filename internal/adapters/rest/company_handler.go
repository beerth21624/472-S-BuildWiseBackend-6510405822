package rest

import (
	"boonkosang/internal/requests"
	"boonkosang/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CompanyHandler struct {
	companyUseCase usecase.CompanyUseCase
}

func NewCompanyHandler(companyUseCase usecase.CompanyUseCase) *CompanyHandler {
	return &CompanyHandler{
		companyUseCase: companyUseCase,
	}
}

func (h *CompanyHandler) CompanyRoutes(app *fiber.App) {
	company := app.Group("/company")

	company.Get("/:userId", h.GetCompanyByUserID)
	company.Put("/:userId", h.UpdateCompany)
}

// GetCompanyByUserID retrieves company for a specific user
func (h *CompanyHandler) GetCompanyByUserID(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	company, err := h.companyUseCase.GetCompanyByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Company retrieved successfully",
		"data":    company,
	})

}

// UpdateCompany updates company information
func (h *CompanyHandler) UpdateCompany(c *fiber.Ctx) error {
	var req requests.UpdateCompanyRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields",
		})
	}

	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	company, err := h.companyUseCase.UpdateCompany(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Company updated successfully",
		"data":    company,
	})
}
