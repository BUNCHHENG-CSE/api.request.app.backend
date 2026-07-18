package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Workspace represents the core business entity
type Workspace struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string    `json:"name"`
	OwnerID   uuid.UUID `json:"owner_id" gorm:"type:uuid"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WorkspaceRepository defines the contract for data persistence
// This Dependency Inversion keeps the domain unaware of GORM.
type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *Workspace) error
	GetByID(ctx context.Context, id uuid.UUID) (*Workspace, error)
	List(ctx context.Context) ([]Workspace, error)
}
