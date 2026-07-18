package domain

import (
	"context"
	"time"
)

// Workspace represents the core business entity
type Workspace struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	OwnerID   uint      `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WorkspaceRepository defines the contract for data persistence
// This Dependency Inversion keeps the domain unaware of GORM.
type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *Workspace) error
	GetByID(ctx context.Context, id uint) (*Workspace, error)
	List(ctx context.Context) ([]Workspace, error)
}
