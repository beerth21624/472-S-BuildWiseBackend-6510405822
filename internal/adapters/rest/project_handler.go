package rest

import (
	"boonkosang/internal/requests"
	"boonkosang/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	projectUsecase usecase.ProjectUsecase
}

func NewProjectHandler(projectUsecase usecase.ProjectUsecase) *ProjectHandler {
	return &ProjectHandler{
		projectUsecase: projectUsecase,
	}
}

func (h *ProjectHandler) ProjectRoutes(app *fiber.App) {
	project := app.Group("/projects")

	project.Post("/", h.Create)
	project.Get("/", h.List)
	project.Get("/:id", h.GetByID)
	project.Put("/:id/cancel", h.Cancel)
	project.Put("/:id", h.Update)

}

func (h *ProjectHandler) Create(c *fiber.Ctx) error {
	var req requests.CreateProjectRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	project, err := h.projectUsecase.Create(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Project created successfully",
		"data":    project,
	})
}

func (h *ProjectHandler) Update(c *fiber.Ctx) error {
	var req requests.UpdateProjectRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	uuid, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID format",
		})
	}

	err = h.projectUsecase.Update(c.Context(), uuid, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Project updated successfully",
	})
}

func (h *ProjectHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	uuid, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID format",
		})
	}

	project, err := h.projectUsecase.GetByID(c.Context(), uuid)
	if err != nil {
		if err.Error() == "project not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Project not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve project",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Project retrieved successfully",
		"data":    project,
	})
}

func (h *ProjectHandler) List(c *fiber.Ctx) error {

	project, err := h.projectUsecase.List(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve projects",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Projects retrieved successfully",
		"data":    project,
	})
}

func (h *ProjectHandler) Cancel(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	uuid, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID format",
		})
	}

	err = h.projectUsecase.Cancel(c.Context(), uuid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Project cancelled successfully",
	})
}
