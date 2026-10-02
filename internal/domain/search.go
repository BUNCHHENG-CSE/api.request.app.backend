package domain

import (
	"context"

	"github.com/google/uuid"
)

type SearchResult struct {
	Type        string    `json:"type"`
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Name        string    `json:"name"`
	Method      string    `json:"method,omitempty"`
	URL         string    `json:"url,omitempty"`
}

type SearchRepository interface {
	Search(ctx context.Context, ownerID uuid.UUID, workspaceID *uuid.UUID, query string, limit int) ([]SearchResult, error)
}
