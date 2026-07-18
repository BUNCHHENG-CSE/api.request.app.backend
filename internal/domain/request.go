package domain

import (
	"context"
	"time"
)

// Request represents an individual API request configuration.
type Request struct {
	ID        uint      `json:"id"`
	FlowID    uint      `json:"flow_id"`
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
	GetByID(ctx context.Context, id uint) (*Request, error)
	ListByFlow(ctx context.Context, flowID uint) ([]Request, error)
	Update(ctx context.Context, req *Request) error
	Delete(ctx context.Context, id uint) error
}
