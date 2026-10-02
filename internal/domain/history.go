package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RequestHistory struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID          uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	WorkspaceID     uuid.UUID `json:"workspace_id" gorm:"type:uuid;not null;index"`
	RequestID       uuid.UUID `json:"request_id" gorm:"type:uuid;not null;index"`
	Method          string    `json:"method"`
	URL             string    `json:"url"`
	Status          int       `json:"status"`
	DurationMS      int64     `json:"time"`
	SizeBytes       int       `json:"size_bytes"`
	ResponseHeaders string    `json:"response_headers"`
	ResponseBody    string    `json:"response_body"`
	Error           string    `json:"error,omitempty"`
	CreatedAt       time.Time `json:"created_at" gorm:"index"`
}

type RequestHistoryRepository interface {
	Create(ctx context.Context, history *RequestHistory) error
	GetByID(ctx context.Context, id, userID uuid.UUID) (*RequestHistory, error)
	List(ctx context.Context, userID uuid.UUID, workspaceID *uuid.UUID, limit, offset int) ([]RequestHistory, int64, error)
	DeleteByUser(ctx context.Context, userID uuid.UUID) error
}
