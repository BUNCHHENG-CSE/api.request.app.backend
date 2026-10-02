package application

import (
	"context"
	"errors"
	"strings"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type SearchService interface {
	Search(ctx context.Context, ownerID uuid.UUID, workspaceID *uuid.UUID, query string, limit int) ([]domain.SearchResult, error)
}

type searchService struct{ repo domain.SearchRepository }

func NewSearchService(repo domain.SearchRepository) SearchService { return &searchService{repo: repo} }

func (s *searchService) Search(ctx context.Context, ownerID uuid.UUID, workspaceID *uuid.UUID, query string, limit int) ([]domain.SearchResult, error) {
	query = strings.TrimSpace(query)
	if len(query) < 2 {
		return nil, errors.New("search query must contain at least 2 characters")
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.Search(ctx, ownerID, workspaceID, query, limit)
}
