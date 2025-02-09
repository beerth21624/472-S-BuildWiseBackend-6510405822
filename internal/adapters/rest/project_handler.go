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
	project.Get("/:projectId/summary", h.GetProjectSummary)
	project.Get("/:projectId/overview", h.GetProjectOverview)
	project.Get("/:id", h.GetByID)
	project.Put("/:projectId/status", h.UpdateStatus)

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

func (h *ProjectHandler) UpdateStatus(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var req requests.UpdateProjectStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	req.ProjectID = projectID

	if err := h.projectUsecase.UpdateProjectStatus(c.Context(), req); err != nil {
		switch err.Error() {
		case "BOQ must be approved", "quotation must be approved":
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		case "project must be in planning status to move to in_progress",
			"project must be in in_progress status to move to completed":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update project status",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Project status updated successfully",
	})
}

func (h *ProjectHandler) GetProjectOverview(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	overview, err := h.projectUsecase.GetProjectOverview(c.Context(), projectID)
	if err != nil {
		if err.Error() == "some materials are missing price information" {
			return c.Status(fiber.StatusPreconditionFailed).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get project overview",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Project overview retrieved successfully",
		"data":    overview,
	})
}

func (h *ProjectHandler) GetProjectSummary(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("projectId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	summary, err := h.projectUsecase.GetProjectSummary(c.Context(), projectID)
	if err != nil {
		switch err.Error() {
		case "project must be completed to view summary":
			return c.Status(fiber.StatusPreconditionFailed).JSON(fiber.Map{
				"error": err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to get project summary",
				"details": err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Project summary retrieved successfully",
		"data":    summary,
	})
}
