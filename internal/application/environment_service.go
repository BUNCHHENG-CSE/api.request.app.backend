package application

import (
	"context"
	"errors"

	"backend/internal/domain"
)

type EnvironmentService interface {
	CreateEnvironment(ctx context.Context, workspaceID uint, name, variables string) (*domain.Environment, error)
	GetEnvironmentsByWorkspace(ctx context.Context, workspaceID uint) ([]domain.Environment, error)
}

type environmentService struct {
	repo domain.EnvironmentRepository
}

func NewEnvironmentService(repo domain.EnvironmentRepository) EnvironmentService {
	return &environmentService{repo: repo}
}

func (s *environmentService) CreateEnvironment(ctx context.Context, workspaceID uint, name, variables string) (*domain.Environment, error) {
	if name == "" {
		return nil, errors.New("environment name is required")
	}

	env := &domain.Environment{
		WorkspaceID: workspaceID,
		Name:        name,
		Variables:   variables,
	}

	if err := s.repo.Create(ctx, env); err != nil {
		return nil, err
	}

	return env, nil
}

func (s *environmentService) GetEnvironmentsByWorkspace(ctx context.Context, workspaceID uint) ([]domain.Environment, error) {
	return s.repo.ListByWorkspace(ctx, workspaceID)
}
