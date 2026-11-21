package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Flow represents a collection of requests or logical steps.
type Flow struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	WorkspaceID  uuid.UUID `json:"workspace_id" gorm:"type:uuid;not null;index"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Nodes        string    `json:"nodes"`
	Edges        string    `json:"edges"`
	WebhookToken string    `json:"webhook_token,omitempty" gorm:"type:varchar(64);index"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// FlowRepository defines the contract for Flow data persistence.
type FlowRepository interface {
	Create(ctx context.Context, flow *Flow) error
	GetByID(ctx context.Context, id uuid.UUID) (*Flow, error)
	GetByWebhookToken(ctx context.Context, token string) (*Flow, error)
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]Flow, error)
	Update(ctx context.Context, flow *Flow) error
	Delete(ctx context.Context, id uuid.UUID) error
}
