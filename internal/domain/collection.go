package domain

import (
	"context"
	"time"
)

// Collection represents a group of flows or requests within a workspace.
type Collection struct {
	ID          uint      `json:"id"`
	WorkspaceID uint      `json:"workspace_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CollectionRepository defines the contract for Collection data persistence.
type CollectionRepository interface {
	Create(ctx context.Context, collection *Collection) error
	GetByID(ctx context.Context, id uint) (*Collection, error)
	ListByWorkspace(ctx context.Context, workspaceID uint) ([]Collection, error)
	Delete(ctx context.Context, id uint) error
}
