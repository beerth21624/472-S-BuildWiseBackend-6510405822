package rest

import (
	"boonkosang/internal/requests"
	"boonkosang/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type BOQHandler struct {
	boqUsecase usecase.BOQUsecase
}

func NewBOQHandler(boqUsecase usecase.BOQUsecase) *BOQHandler {
	return &BOQHandler{
		boqUsecase: boqUsecase,
	}
}

func (h *BOQHandler) BOQRoutes(app *fiber.App) {
	boq := app.Group("/boqs")

	boq.Get("/project/:projectId/export", h.ExportBOQ)

	boq.Post("/:id/approve", h.Approve)
	boq.Get("/project/:project_id", h.GetBoqWithProject)
	boq.Post("/:id/jobs", h.AddBOQJob)
	boq.Put("/:id/jobs", h.UpdateBOQJob)
	boq.Delete("/:id/jobs/:jobId", h.DeleteBOQJob)
}

func (h *BOQHandler) Approve(c *fiber.Ctx) error {
	boqID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid BOQ ID",
		})
	}

	err = h.boqUsecase.Approve(c.Context(), boqID)
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "BOQ approved successfully",
	})
}

func (h *BOQHandler) GetBoqWithProject(c *fiber.Ctx) error {
	project_id := c.Params("project_id")
	if project_id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	uuid, err := uuid.Parse(project_id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID format",
		})
	}

	boq, err := h.boqUsecase.GetBoqWithProject(c.Context(), uuid)
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "BOQ retrieved successfully",
		"data":    boq,
	})
}

func (h *BOQHandler) AddBOQJob(c *fiber.Ctx) error {
	boqID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid BOQ ID",
		})
	}

	var req requests.BOQJobRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err = h.boqUsecase.AddBOQJob(c.Context(), boqID, req)
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "BOQ job added successfully",
	})
}

func (h *BOQHandler) UpdateBOQJob(c *fiber.Ctx) error {
	boqID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid BOQ ID",
		})
	}

	var req requests.BOQJobRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err = h.boqUsecase.UpdateBOQJob(c.Context(), boqID, req)
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "BOQ job updated successfully",
	})
}

func (h *BOQHandler) DeleteBOQJob(c *fiber.Ctx) error {
	boqID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid BOQ ID",
		})
	}

	jobID, err := uuid.Parse(c.Params("jobId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid job ID",
		})
	}

	err = h.boqUsecase.DeleteBOQJob(c.Context(), boqID, jobID)
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "BOQ job deleted successfully",
	})
}

func (h *BOQHandler) ExportBOQ(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	// Get BOQ summary data
	summary, err := h.boqUsecase.GetBOQSummary(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "BOQ summary retrieved successfully",
		"data":    summary,
	})

}
