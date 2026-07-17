package workspace

import "time"

// Workspace represents the core database entity
type Workspace struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateWorkspaceRequest is the DTO (Data Transfer Object) for incoming HTTP requests
type CreateWorkspaceRequest struct {
	Name        string `json:"name" binding:"required,min=3"`
	Description string `json:"description"`
}

// UpdateWorkspaceRequest is the DTO for updates
type UpdateWorkspaceRequest struct {
	Name        string `json:"name" binding:"omitempty,min=3"`
	Description string `json:"description"`
}
