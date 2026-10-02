package application

import (
	"context"
	"errors"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, name string, ownerID uuid.UUID) (*domain.Workspace, error)
	GetWorkspace(ctx context.Context, id, ownerID uuid.UUID) (*domain.Workspace, error)
	ListWorkspaces(ctx context.Context, ownerID uuid.UUID) ([]domain.Workspace, error)
	UpdateWorkspace(ctx context.Context, id, ownerID uuid.UUID, name string) (*domain.Workspace, error)
	DeleteWorkspace(ctx context.Context, id, ownerID uuid.UUID) error
}

type workspaceService struct {
	repo domain.WorkspaceRepository
}

// NewWorkspaceService injects the repository dependency
func NewWorkspaceService(repo domain.WorkspaceRepository) WorkspaceService {
	return &workspaceService{repo: repo}
}

func (s *workspaceService) CreateWorkspace(ctx context.Context, name string, ownerID uuid.UUID) (*domain.Workspace, error) {
	if name == "" {
		return nil, errors.New("workspace name is required")
	}

	ws := &domain.Workspace{
		Name:    name,
		OwnerID: ownerID,
	}

	if err := s.repo.Create(ctx, ws); err != nil {
		return nil, err
	}
	return ws, nil
}

func (s *workspaceService) GetWorkspace(ctx context.Context, id, ownerID uuid.UUID) (*domain.Workspace, error) {
	return s.repo.GetByID(ctx, id, ownerID)
}

func (s *workspaceService) ListWorkspaces(ctx context.Context, ownerID uuid.UUID) ([]domain.Workspace, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *workspaceService) UpdateWorkspace(ctx context.Context, id, ownerID uuid.UUID, name string) (*domain.Workspace, error) {
	if name == "" {
		return nil, errors.New("workspace name is required")
	}
	workspace, err := s.repo.GetByID(ctx, id, ownerID)
	if err != nil {
		return nil, err
	}
	workspace.Name = name
	if err := s.repo.Update(ctx, workspace); err != nil {
		return nil, err
	}
	return workspace, nil
}

func (s *workspaceService) DeleteWorkspace(ctx context.Context, id, ownerID uuid.UUID) error {
	return s.repo.Delete(ctx, id, ownerID)
}
