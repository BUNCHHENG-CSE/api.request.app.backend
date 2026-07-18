package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Environment represents configuration variables for a workspace.
type Environment struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	WorkspaceID uuid.UUID `json:"workspace_id" gorm:"type:uuid"`
	Name        string    `json:"name"`
	Variables   string    `json:"variables"` // Stored as a JSON string or JSONB in Postgres
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EnvironmentRepository defines the contract for Environment data persistence.
type EnvironmentRepository interface {
	Create(ctx context.Context, env *Environment) error
	GetByID(ctx context.Context, id uuid.UUID) (*Environment, error)
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]Environment, error)
	Update(ctx context.Context, env *Environment) error
	Delete(ctx context.Context, id uuid.UUID) error
}
