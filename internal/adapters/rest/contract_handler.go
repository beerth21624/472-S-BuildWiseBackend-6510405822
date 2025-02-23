package rest

import (
	"boonkosang/internal/requests"
	"boonkosang/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ContractHandler struct {
	contractUseCase usecase.ContractUseCase
}

func NewContractHandler(contractUseCase usecase.ContractUseCase) *ContractHandler {
	return &ContractHandler{
		contractUseCase: contractUseCase,
	}
}

func (h *ContractHandler) ContractRoutes(app *fiber.App) {
	contract := app.Group("/contracts")
	contract.Post("/", h.CreateContract)
	contract.Put("/:project_id", h.UpdateContract)
	contract.Delete("/:project_id", h.DeleteContract)
	//change status of contract

	contract.Put("/:project_id/status", h.ChangeStatus)
	contract.Get("/:project_id", h.GetContractByProjectID)
}

func (h *ContractHandler) CreateContract(c *fiber.Ctx) error {
	var req requests.CreateContractRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	if err := h.contractUseCase.Create(c.Context(), &req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Contract created successfully",
	})
}

func (h *ContractHandler) UpdateContract(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("project_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid project ID",
		})
	}

	var req requests.UpdateContractRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	if err := h.contractUseCase.Update(c.Context(), projectID, &req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Contract updated successfully",
	})
}

func (h *ContractHandler) DeleteContract(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("project_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid project ID",
		})
	}

	if err := h.contractUseCase.Delete(c.Context(), projectID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Contract deleted successfully",
	})
}

func (h *ContractHandler) GetContractByProjectID(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("project_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid project ID",
		})
	}

	contract, err := h.contractUseCase.GetByProjectID(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(contract)
}

func (h *ContractHandler) ChangeStatus(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("project_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid project ID",
		})
	}

	if err := h.contractUseCase.ChangeStatus(c.Context(), projectID, "approved"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Contract status updated successfully",
	})
}
