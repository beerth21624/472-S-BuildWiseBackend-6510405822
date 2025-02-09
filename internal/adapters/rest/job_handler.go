package rest

import (
	"boonkosang/internal/requests"
	"boonkosang/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type JobHandler struct {
	jobUsecase usecase.JobUseCase
}

func NewJobHandler(jobUsecase usecase.JobUseCase) *JobHandler {
	return &JobHandler{
		jobUsecase: jobUsecase,
	}
}

func (h *JobHandler) JobRoutes(app *fiber.App) {
	job := app.Group("/jobs")

	job.Get("/", h.List)
	job.Post("/", h.Create)
	job.Get("/:id", h.GetByID)
	job.Put("/:id", h.Update)
	job.Delete("/:id", h.Delete)

	// Material management routes
	job.Post("/:id/materials", h.AddMaterial)
	job.Delete("/:id/materials/:materialId", h.DeleteMaterial)
	job.Put("/:id/materials/:materialId/quantity", h.UpdateMaterialQuantity)

}

func (h *JobHandler) Create(c *fiber.Ctx) error {
	var req requests.CreateJobRequest

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

	job, err := h.jobUsecase.Create(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create job",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Job created successfully",
		"data":    job,
	})
}

func (h *JobHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid job ID",
		})
	}

	job, err := h.jobUsecase.GetByID(c.Context(), id)
	if err != nil {
		if err.Error() == "job not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Job not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Job retrieved successfully",
		"data":    job,
	})
}

func (h *JobHandler) List(c *fiber.Ctx) error {
	jobs, err := h.jobUsecase.GetJobList(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve jobs",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Jobs retrieved successfully",
		"data":    jobs,
	})
}

func (h *JobHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid job ID",
		})
	}

	err = h.jobUsecase.Delete(c.Context(), id)
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Job deleted successfully",
	})
}

func (h *JobHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid job ID",
		})
	}

	var req requests.UpdateJobRequest
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

	err = h.jobUsecase.Update(c.Context(), id, req)
	if err != nil {
		if err.Error() == "job not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Job not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update job",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Job updated successfully",
	})
}

// Material management handlers
func (h *JobHandler) AddMaterial(c *fiber.Ctx) error {
	jobID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid job ID",
		})
	}

	var req requests.AddJobMaterialRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err = h.jobUsecase.AddMaterial(c.Context(), jobID, req)
	if err != nil {
		switch err.Error() {
		case "job not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Job not found",
			})
		case "material already exists":
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Material already exists for this job",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to add material",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Material added successfully",
	})
}

func (h *JobHandler) DeleteMaterial(c *fiber.Ctx) error {
	jobID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid job ID",
		})
	}

	materialID := c.Params("materialId")
	if materialID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid material ID",
		})
	}

	err = h.jobUsecase.DeleteMaterial(c.Context(), jobID, materialID)
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Material deleted successfully",
	})
}

func (h *JobHandler) UpdateMaterialQuantity(c *fiber.Ctx) error {
	jobID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid job ID",
		})
	}

	materialID := c.Params("materialId")
	if materialID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid material ID",
		})
	}

	var req requests.UpdateJobMaterialQuantityRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	req.JobID = jobID
	req.MaterialID = materialID
	err = h.jobUsecase.UpdateMaterialQuantity(c.Context(), jobID, req)
	if err != nil {
		if err.Error() == "job material not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Job material not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update material quantity",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Material quantity updated successfully",
	})
}
