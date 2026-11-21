package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type WorkspaceCookie struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	WorkspaceID uuid.UUID  `json:"workspace_id" gorm:"type:uuid;not null;index"`
	Domain      string     `json:"domain" gorm:"not null;index"`
	Path        string     `json:"path" gorm:"not null;default:'/'"`
	Name        string     `json:"name" gorm:"not null"`
	Value       string     `json:"value" gorm:"not null"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Secure      bool       `json:"secure" gorm:"default:false"`
	HttpOnly    bool       `json:"http_only" gorm:"default:false"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type WorkspaceCookieRepository interface {
	Save(ctx context.Context, cookie *WorkspaceCookie) error
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]WorkspaceCookie, error)
	GetMatching(ctx context.Context, workspaceID uuid.UUID, domain, path string) ([]WorkspaceCookie, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ClearByWorkspace(ctx context.Context, workspaceID uuid.UUID) error
	DeleteByDomain(ctx context.Context, workspaceID uuid.UUID, domain string) error
}
