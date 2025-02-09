// rest/client_handler.go
package rest

import (
	"boonkosang/internal/requests"
	"boonkosang/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ClientHandler struct {
	clientUsecase usecase.ClientUsecase
}

func NewClientHandler(clientUsecase usecase.ClientUsecase) *ClientHandler {
	return &ClientHandler{
		clientUsecase: clientUsecase,
	}
}

func (h *ClientHandler) ClientRoutes(app *fiber.App) {
	client := app.Group("/clients")

	client.Post("/", h.Create)
	client.Get("/", h.List)
	client.Get("/:id", h.GetByID)
	client.Put("/:id", h.Update)
	client.Delete("/:id", h.Delete)
}

func (h *ClientHandler) Create(c *fiber.Ctx) error {
	var req requests.CreateClientRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" || req.Email == "" || req.Tel == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields",
		})
	}

	client, err := h.clientUsecase.Create(c.Context(), req)
	if err != nil {
		switch err.Error() {
		case "client with this email already exists":
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Client with this email already exists",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create client",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Client created successfully",
		"data":    client,
	})
}

func (h *ClientHandler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	response, err := h.clientUsecase.List(c.Context(), page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve clients",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Clients retrieved successfully",
		"data":    response,
	})
}

func (h *ClientHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid client ID",
		})
	}

	client, err := h.clientUsecase.GetByID(c.Context(), id)
	if err != nil {
		if err.Error() == "client not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Client not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve client",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Client retrieved successfully",
		"data":    client,
	})
}

func (h *ClientHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid client ID",
		})
	}

	var req requests.UpdateClientRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" || req.Email == "" || req.Tel == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields",
		})
	}

	err = h.clientUsecase.Update(c.Context(), id, req)
	if err != nil {
		switch err.Error() {
		case "client not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Client not found",
			})
		case "client with this email already exists":
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Client with this email already exists",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update client",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Client updated successfully",
	})
}

func (h *ClientHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid client ID",
		})
	}

	err = h.clientUsecase.Delete(c.Context(), id)
	if err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Client deleted successfully",
	})
}
