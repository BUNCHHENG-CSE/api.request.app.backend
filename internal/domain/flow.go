package domain

import (
	"context"
	"time"
)

// Flow represents a collection of requests or logical steps.
type Flow struct {
	ID          uint      `json:"id"`
	WorkspaceID uint      `json:"workspace_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FlowRepository defines the contract for Flow data persistence.
type FlowRepository interface {
	Create(ctx context.Context, flow *Flow) error
	GetByID(ctx context.Context, id uint) (*Flow, error)
	ListByWorkspace(ctx context.Context, workspaceID uint) ([]Flow, error)
	Update(ctx context.Context, flow *Flow) error
	Delete(ctx context.Context, id uint) error
}
