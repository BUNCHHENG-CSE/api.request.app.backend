package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type WorkspaceActivity struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	WorkspaceID uuid.UUID `json:"workspace_id" gorm:"type:uuid;not null;index"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Username    string    `json:"username" gorm:"not null"`
	Action      string    `json:"action" gorm:"not null"` // e.g. "request.create", "flow.run", "member.add"
	Details     string    `json:"details"`
	CreatedAt   time.Time `json:"created_at" gorm:"index"`
}

type WorkspaceActivityRepository interface {
	Create(ctx context.Context, activity *WorkspaceActivity) error
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit int) ([]WorkspaceActivity, error)
}
