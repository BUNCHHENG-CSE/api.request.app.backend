package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Monitor struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	WorkspaceID uuid.UUID  `json:"workspace_id" gorm:"type:uuid;not null;index"`
	FlowID      uuid.UUID  `json:"flow_id" gorm:"type:uuid;not null;index"`
	Name        string     `json:"name" gorm:"not null"`
	IntervalMin int        `json:"interval_minutes" gorm:"not null;default:60"`
	WebhookURL  string     `json:"webhook_url"`
	IsActive    bool       `json:"is_active" gorm:"not null;default:true"`
	LastRunAt   *time.Time `json:"last_run_at,omitempty"`
	LastStatus  string     `json:"last_status,omitempty"` // "passed", "failed"
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type MonitorRunLog struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	MonitorID   uuid.UUID `json:"monitor_id" gorm:"type:uuid;not null;index"`
	Status      string    `json:"status"` // "passed", "failed"
	TotalSteps  int       `json:"total_steps"`
	PassedSteps int       `json:"passed_steps"`
	FailedSteps int       `json:"failed_steps"`
	DurationMS  int64     `json:"duration_ms"`
	Error       string    `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at" gorm:"index"`
}

type MonitorRepository interface {
	Create(ctx context.Context, monitor *Monitor) error
	GetByID(ctx context.Context, id uuid.UUID) (*Monitor, error)
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]Monitor, error)
	Update(ctx context.Context, monitor *Monitor) error
	Delete(ctx context.Context, id uuid.UUID) error
	CreateRunLog(ctx context.Context, log *MonitorRunLog) error
	ListRunLogs(ctx context.Context, monitorID uuid.UUID, limit int) ([]MonitorRunLog, error)
}
