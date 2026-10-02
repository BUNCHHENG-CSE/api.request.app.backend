package application

import (
	"context"
	"errors"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type EnvironmentService interface {
	CreateEnvironment(ctx context.Context, ownerID, workspaceID uuid.UUID, name, variables string) (*domain.Environment, error)
	GetEnvironment(ctx context.Context, ownerID, id uuid.UUID) (*domain.Environment, error)
	GetEnvironmentsByWorkspace(ctx context.Context, ownerID, workspaceID uuid.UUID) ([]domain.Environment, error)
	UpdateEnvironment(ctx context.Context, ownerID, id uuid.UUID, name, variables *string) (*domain.Environment, error)
	DeleteEnvironment(ctx context.Context, ownerID, id uuid.UUID) error
}

type environmentService struct {
	repo          domain.EnvironmentRepository
	workspaceRepo domain.WorkspaceRepository
}

func NewEnvironmentService(repo domain.EnvironmentRepository, workspaceRepo domain.WorkspaceRepository) EnvironmentService {
	return &environmentService{repo: repo, workspaceRepo: workspaceRepo}
}

func (s *environmentService) authorize(ctx context.Context, ownerID uuid.UUID, env *domain.Environment) error {
	_, err := s.workspaceRepo.GetByID(ctx, env.WorkspaceID, ownerID)
	return err
}

func (s *environmentService) CreateEnvironment(ctx context.Context, ownerID, workspaceID uuid.UUID, name, variables string) (*domain.Environment, error) {
	if name == "" {
		return nil, errors.New("environment name is required")
	}
	if _, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID); err != nil {
		return nil, err
	}
	env := &domain.Environment{WorkspaceID: workspaceID, Name: name, Variables: variables}
	if err := s.repo.Create(ctx, env); err != nil {
		return nil, err
	}
	return env, nil
}

func (s *environmentService) GetEnvironment(ctx context.Context, ownerID, id uuid.UUID) (*domain.Environment, error) {
	env, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.authorize(ctx, ownerID, env); err != nil {
		return nil, err
	}
	return env, nil
}

func (s *environmentService) GetEnvironmentsByWorkspace(ctx context.Context, ownerID, workspaceID uuid.UUID) ([]domain.Environment, error) {
	if _, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID); err != nil {
		return nil, err
	}
	return s.repo.ListByWorkspace(ctx, workspaceID)
}

func (s *environmentService) UpdateEnvironment(ctx context.Context, ownerID, id uuid.UUID, name, variables *string) (*domain.Environment, error) {
	env, err := s.GetEnvironment(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		if *name == "" {
			return nil, errors.New("environment name cannot be empty")
		}
		env.Name = *name
	}
	if variables != nil {
		env.Variables = *variables
	}
	if err := s.repo.Update(ctx, env); err != nil {
		return nil, err
	}
	return env, nil
}

func (s *environmentService) DeleteEnvironment(ctx context.Context, ownerID, id uuid.UUID) error {
	if _, err := s.GetEnvironment(ctx, ownerID, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
