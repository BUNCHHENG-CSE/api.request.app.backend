package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Collection represents a group of flows or requests within a workspace.
type Collection struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	WorkspaceID uuid.UUID `json:"workspace_id" gorm:"type:uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CollectionRepository defines the contract for Collection data persistence.
type CollectionRepository interface {
	Create(ctx context.Context, collection *Collection) error
	GetByID(ctx context.Context, id uuid.UUID) (*Collection, error)
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]Collection, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
