package domain

import (
	"context"
	"time"
)

// Environment represents configuration variables for a workspace.
type Environment struct {
	ID          uint      `json:"id"`
	WorkspaceID uint      `json:"workspace_id"`
	Name        string    `json:"name"`
	Variables   string    `json:"variables"` // Stored as a JSON string or JSONB in Postgres
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EnvironmentRepository defines the contract for Environment data persistence.
type EnvironmentRepository interface {
	Create(ctx context.Context, env *Environment) error
	GetByID(ctx context.Context, id uint) (*Environment, error)
	ListByWorkspace(ctx context.Context, workspaceID uint) ([]Environment, error)
	Update(ctx context.Context, env *Environment) error
	Delete(ctx context.Context, id uint) error
}
