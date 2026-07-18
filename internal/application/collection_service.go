package application

import (
	"context"
	"errors"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type CollectionService interface {
	CreateCollection(ctx context.Context, workspaceID uuid.UUID, name, description string) (*domain.Collection, error)
}

type collectionService struct {
	repo domain.CollectionRepository
}

func NewCollectionService(repo domain.CollectionRepository) CollectionService {
	return &collectionService{repo: repo}
}

func (s *collectionService) CreateCollection(ctx context.Context, workspaceID uuid.UUID, name, description string) (*domain.Collection, error) {
	if name == "" {
		return nil, errors.New("collection name is required")
	}

	collection := &domain.Collection{
		WorkspaceID: workspaceID,
		Name:        name,
		Description: description,
	}

	if err := s.repo.Create(ctx, collection); err != nil {
		return nil, err
	}
	return collection, nil
}
