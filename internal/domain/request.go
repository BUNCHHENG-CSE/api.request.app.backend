package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Request represents an individual API request configuration.
type Request struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FlowID       uuid.UUID  `json:"flow_id" gorm:"type:uuid;not null;index"`
	CollectionID *uuid.UUID `json:"collection_id,omitempty" gorm:"type:uuid;index"`
	Name         string     `json:"name"`
	Position     int        `json:"position" gorm:"not null;default:0;index"`
	Method       string     `json:"method"`
	URL          string     `json:"url"`
	Headers      string     `json:"headers"`
	QueryParams  string     `json:"query_params"`
	Body         string     `json:"body"`
	BodyType     string     `json:"body_type"`
	Auth         string     `json:"auth"`
	Scripts      string     `json:"scripts"`
	Settings     string     `json:"settings"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// RequestRepository defines the contract for Request data persistence.
type RequestRepository interface {
	Create(ctx context.Context, req *Request) error
	GetByID(ctx context.Context, id uuid.UUID) (*Request, error)
	ListByFlow(ctx context.Context, flowID uuid.UUID) ([]Request, error)
	Update(ctx context.Context, req *Request) error
	Delete(ctx context.Context, id uuid.UUID) error
}
