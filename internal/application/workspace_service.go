package application

import (
	"backend/internal/domain"
	"context"
	"errors"
)

type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, name string, ownerID uint) (*domain.Workspace, error)
	GetWorkspace(ctx context.Context, id uint) (*domain.Workspace, error)
}

type workspaceService struct {
	repo domain.WorkspaceRepository
}

// NewWorkspaceService injects the repository dependency
func NewWorkspaceService(repo domain.WorkspaceRepository) WorkspaceService {
	return &workspaceService{repo: repo}
}

func (s *workspaceService) CreateWorkspace(ctx context.Context, name string, ownerID uint) (*domain.Workspace, error) {
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

func (s *workspaceService) GetWorkspace(ctx context.Context, id uint) (*domain.Workspace, error) {
	return s.repo.GetByID(ctx, id)
}
