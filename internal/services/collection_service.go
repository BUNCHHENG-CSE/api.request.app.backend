package services

import (
	"context"

	"backend/internal/entities"
	"backend/internal/models"
	"backend/internal/repositories"
)

// CollectionService handles collection operations
type CollectionService interface {
	CreateCollection(ctx context.Context, workspaceID string, dto *models.CreateCollectionDTO) (*entities.Collection, error)
	GetCollection(ctx context.Context, id string) (*entities.Collection, error)
	ListCollections(ctx context.Context, workspaceID string) ([]*entities.Collection, error)
	UpdateCollection(ctx context.Context, id string, dto *models.UpdateCollectionDTO) (*entities.Collection, error)
	DeleteCollection(ctx context.Context, id string) error
	ReorderCollections(ctx context.Context, collectionIDs []string) error
}

type collectionService struct {
	repo repositories.CollectionRepository
}

func NewCollectionService(repo repositories.CollectionRepository) CollectionService {
	return &collectionService{repo: repo}
}

func (s *collectionService) CreateCollection(ctx context.Context, workspaceID string, dto *models.CreateCollectionDTO) (*entities.Collection, error) {
	collection := &entities.Collection{
		WorkspaceID: workspaceID,
		Name:        dto.Name,
		Description: dto.Description,
		Icon:        dto.Icon,
	}

	return s.repo.Create(ctx, collection)
}

func (s *collectionService) GetCollection(ctx context.Context, id string) (*entities.Collection, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *collectionService) ListCollections(ctx context.Context, workspaceID string) ([]*entities.Collection, error) {
	return s.repo.ListByWorkspace(ctx, workspaceID)
}

func (s *collectionService) UpdateCollection(ctx context.Context, id string, dto *models.UpdateCollectionDTO) (*entities.Collection, error) {
	collection, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dto.Name != nil {
		collection.Name = *dto.Name
	}
	if dto.Description != nil {
		collection.Description = *dto.Description
	}
	if dto.Icon != nil {
		collection.Icon = *dto.Icon
	}

	return s.repo.Update(ctx, collection)
}

func (s *collectionService) DeleteCollection(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *collectionService) ReorderCollections(ctx context.Context, collectionIDs []string) error {
	return s.repo.UpdateOrder(ctx, collectionIDs)
}
