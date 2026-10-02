package application

import (
	"context"
	"errors"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type CollectionService interface {
	CreateCollection(ctx context.Context, ownerID, workspaceID uuid.UUID, name, description string) (*domain.Collection, error)
	GetCollection(ctx context.Context, ownerID, id uuid.UUID) (*domain.Collection, error)
	GetCollectionsByWorkspace(ctx context.Context, ownerID, workspaceID uuid.UUID) ([]domain.Collection, error)
	UpdateCollection(ctx context.Context, ownerID, id uuid.UUID, name, description *string) (*domain.Collection, error)
	DeleteCollection(ctx context.Context, ownerID, id uuid.UUID) error
}

type collectionService struct {
	repo          domain.CollectionRepository
	workspaceRepo domain.WorkspaceRepository
}

func NewCollectionService(repo domain.CollectionRepository, workspaceRepo domain.WorkspaceRepository) CollectionService {
	return &collectionService{repo: repo, workspaceRepo: workspaceRepo}
}

func (s *collectionService) authorize(ctx context.Context, ownerID uuid.UUID, collection *domain.Collection) error {
	_, err := s.workspaceRepo.GetByID(ctx, collection.WorkspaceID, ownerID)
	return err
}

func (s *collectionService) CreateCollection(ctx context.Context, ownerID, workspaceID uuid.UUID, name, description string) (*domain.Collection, error) {
	if name == "" {
		return nil, errors.New("collection name is required")
	}
	if _, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID); err != nil {
		return nil, err
	}
	collection := &domain.Collection{WorkspaceID: workspaceID, Name: name, Description: description}
	if err := s.repo.Create(ctx, collection); err != nil {
		return nil, err
	}
	return collection, nil
}

func (s *collectionService) GetCollection(ctx context.Context, ownerID, id uuid.UUID) (*domain.Collection, error) {
	collection, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.authorize(ctx, ownerID, collection); err != nil {
		return nil, err
	}
	return collection, nil
}

func (s *collectionService) GetCollectionsByWorkspace(ctx context.Context, ownerID, workspaceID uuid.UUID) ([]domain.Collection, error) {
	if _, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID); err != nil {
		return nil, err
	}
	return s.repo.ListByWorkspace(ctx, workspaceID)
}

func (s *collectionService) UpdateCollection(ctx context.Context, ownerID, id uuid.UUID, name, description *string) (*domain.Collection, error) {
	collection, err := s.GetCollection(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		if *name == "" {
			return nil, errors.New("collection name cannot be empty")
		}
		collection.Name = *name
	}
	if description != nil {
		collection.Description = *description
	}
	if err := s.repo.Update(ctx, collection); err != nil {
		return nil, err
	}
	return collection, nil
}

func (s *collectionService) DeleteCollection(ctx context.Context, ownerID, id uuid.UUID) error {
	if _, err := s.GetCollection(ctx, ownerID, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
