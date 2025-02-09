package requests

import (
	"boonkosang/internal/domain/models"
	"encoding/json"

	"github.com/google/uuid"
)

type CreateProjectRequest struct {
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description" validate:"required"`
	Address     json.RawMessage `json:"address" validate:"required"`
	ClientID    uuid.UUID       `json:"client_id" validate:"required"`
}

type UpdateProjectRequest struct {
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description" validate:"required"`
	Address     json.RawMessage `json:"address" validate:"required"`
	ClientID    uuid.UUID       `json:"client_id" validate:"required"`
}

type UpdateProjectStatusRequest struct {
	ProjectID uuid.UUID            `json:"project_id"`
	Status    models.ProjectStatus `json:"status" validate:"required,oneof=planning in_progress completed cancelled"`
}
