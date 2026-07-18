package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Request represents an individual API request configuration.
type Request struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FlowID    uuid.UUID `json:"flow_id" gorm:"type:uuid"`
	Method    string    `json:"method"`
	URL       string    `json:"url"`
	Headers   string    `json:"headers"` // Can be mapped to a JSON/JSONB struct depending on DB needs
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RequestRepository defines the contract for Request data persistence.
type RequestRepository interface {
	Create(ctx context.Context, req *Request) error
	GetByID(ctx context.Context, id uuid.UUID) (*Request, error)
	ListByFlow(ctx context.Context, flowID uuid.UUID) ([]Request, error)
	Update(ctx context.Context, req *Request) error
	Delete(ctx context.Context, id uuid.UUID) error
}
