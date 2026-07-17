package services

import (
	"context"
	"encoding/json"

	"backend/internal/entities"
	"backend/internal/models"
	"backend/internal/repositories"
)

// EnvironmentService handles environment operations
type EnvironmentService interface {
	CreateEnvironment(ctx context.Context, workspaceID string, dto *models.CreateEnvironmentDTO) (*entities.Environment, error)
	GetEnvironment(ctx context.Context, id string) (*entities.Environment, error)
	ListEnvironments(ctx context.Context, workspaceID string) ([]*entities.Environment, error)
	UpdateEnvironment(ctx context.Context, id string, dto *models.UpdateEnvironmentDTO) (*entities.Environment, error)
	DeleteEnvironment(ctx context.Context, id string) error
	SetActiveEnvironment(ctx context.Context, workspaceID, environmentID string) error
	GetActiveEnvironment(ctx context.Context, workspaceID string) (*entities.Environment, error)
}

type environmentService struct {
	repo repositories.EnvironmentRepository
}

func NewEnvironmentService(repo repositories.EnvironmentRepository) EnvironmentService {
	return &environmentService{repo: repo}
}

func (s *environmentService) CreateEnvironment(ctx context.Context, workspaceID string, dto *models.CreateEnvironmentDTO) (*entities.Environment, error) {
	// Convert DTOs to JSON
	varsJSON, err := json.Marshal(dto.Variables)
	if err != nil {
		return nil, err
	}

	env := &entities.Environment{
		WorkspaceID: workspaceID,
		Name:        dto.Name,
		Active:      dto.Active,
		Variables:   varsJSON, // Update your entities struct to accept []byte or datatypes.JSON
	}

	return s.repo.Create(ctx, env)
}

func (s *environmentService) GetEnvironment(ctx context.Context, id string) (*entities.Environment, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *environmentService) ListEnvironments(ctx context.Context, workspaceID string) ([]*entities.Environment, error) {
	return s.repo.ListByWorkspace(ctx, workspaceID)
}

func (s *environmentService) UpdateEnvironment(ctx context.Context, id string, dto *models.UpdateEnvironmentDTO) (*entities.Environment, error) {
	env, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dto.Name != nil {
		env.Name = *dto.Name
	}
	if dto.Active != nil {
		env.Active = *dto.Active
	}
	if dto.Variables != nil {
		varsJSON, err := json.Marshal(dto.Variables)
		if err != nil {
			return nil, err
		}
		env.Variables = varsJSON
	}

	return s.repo.Update(ctx, env)
}

func (s *environmentService) DeleteEnvironment(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *environmentService) SetActiveEnvironment(ctx context.Context, workspaceID, environmentID string) error {
	// Get all environments for workspace
	envs, err := s.repo.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Deactivate all
	for _, env := range envs {
		env.Active = false
		_, err2 := s.repo.Update(ctx, env)
		if err2 != nil {
			return err2
		}
	}

	// Activate the specified one
	env, err := s.repo.GetByID(ctx, environmentID)
	if err != nil {
		return err
	}

	env.Active = true
	_, err = s.repo.Update(ctx, env)
	return err
}

func (s *environmentService) GetActiveEnvironment(ctx context.Context, workspaceID string) (*entities.Environment, error) {
	envs, err := s.repo.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	for _, env := range envs {
		if env.Active {
			return env, nil
		}
	}

	return nil, nil
}
