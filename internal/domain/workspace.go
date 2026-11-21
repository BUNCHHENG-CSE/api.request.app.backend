package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type WorkspaceRole string

const (
	RoleAdmin  WorkspaceRole = "admin"
	RoleEditor WorkspaceRole = "editor"
	RoleViewer WorkspaceRole = "viewer"
)

// Workspace represents the core business entity
type Workspace struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string    `json:"name"`
	OwnerID   uuid.UUID `json:"owner_id" gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WorkspaceMember struct {
	WorkspaceID uuid.UUID     `json:"workspace_id" gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID     `json:"user_id" gorm:"type:uuid;primaryKey;index"`
	Role        WorkspaceRole `json:"role" gorm:"type:varchar(20);not null;default:'viewer'"`
	Username    string        `json:"username,omitempty" gorm:"-"`
	Email       string        `json:"email,omitempty" gorm:"-"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// WorkspaceRepository defines the contract for data persistence
type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *Workspace) error
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Workspace, error)
	List(ctx context.Context, userID uuid.UUID) ([]Workspace, error)
	Update(ctx context.Context, workspace *Workspace) error
	Delete(ctx context.Context, id, ownerID uuid.UUID) error
}

type WorkspaceMemberRepository interface {
	Add(ctx context.Context, member *WorkspaceMember) error
	Get(ctx context.Context, workspaceID, userID uuid.UUID) (*WorkspaceMember, error)
	List(ctx context.Context, workspaceID uuid.UUID) ([]WorkspaceMember, error)
	Update(ctx context.Context, member *WorkspaceMember) error
	Delete(ctx context.Context, workspaceID, userID uuid.UUID) error
}
