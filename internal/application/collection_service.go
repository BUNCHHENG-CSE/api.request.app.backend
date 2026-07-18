package application

import (
	"context"
	"errors"

	"backend/internal/domain"
)

type CollectionService interface {
	CreateCollection(ctx context.Context, workspaceID uint, name, description string) (*domain.Collection, error)
}

type collectionService struct {
	repo domain.CollectionRepository
}

func NewCollectionService(repo domain.CollectionRepository) CollectionService {
	return &collectionService{repo: repo}
}

func (s *collectionService) CreateCollection(ctx context.Context, workspaceID uint, name, description string) (*domain.Collection, error) {
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
