package services

import (
	"context"
	"fmt"

	"backend/internal/entities"
	"backend/internal/models"
	"backend/internal/repositories"
)

// WorkspaceService handles workspace operations
type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, userID string, dto *models.CreateWorkspaceDTO) (*entities.Workspace, error)
	GetWorkspace(ctx context.Context, id string) (*entities.Workspace, error)
	ListWorkspaces(ctx context.Context, userID string) ([]*entities.Workspace, error)
	UpdateWorkspace(ctx context.Context, id string, dto *models.UpdateWorkspaceDTO) (*entities.Workspace, error)
	DeleteWorkspace(ctx context.Context, id string) error
	AddMember(ctx context.Context, workspaceID, userID, role string) error
	RemoveMember(ctx context.Context, workspaceID, userID string) error
	GetMembers(ctx context.Context, workspaceID string) ([]*entities.User, error)
	UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error
}

type workspaceService struct {
	repo repositories.WorkspaceRepository
	userRepo repositories.UserRepository
}

func NewWorkspaceService(
	repo repositories.WorkspaceRepository,
	userRepo repositories.UserRepository,
) WorkspaceService {
	return &workspaceService{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *workspaceService) CreateWorkspace(ctx context.Context, userID string, dto *models.CreateWorkspaceDTO) (*entities.Workspace, error) {
	workspace := &entities.Workspace{
		OwnerID:     userID,
		Name:        dto.Name,
		Description: dto.Description,
		Icon:        dto.Icon,
	}

	if err := s.repo.Create(ctx, workspace); err != nil {
		return nil, err
	}

	// Add creator as owner
	if err := s.repo.AddMember(ctx, workspace.ID, userID, "owner"); err != nil {
		return nil, err
	}

	return workspace, nil
}

func (s *workspaceService) GetWorkspace(ctx context.Context, id string) (*entities.Workspace, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *workspaceService) ListWorkspaces(ctx context.Context, userID string) ([]*entities.Workspace, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *workspaceService) UpdateWorkspace(ctx context.Context, id string, dto *models.UpdateWorkspaceDTO) (*entities.Workspace, error) {
	workspace, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dto.Name != nil {
		workspace.Name = *dto.Name
	}
	if dto.Description != nil {
		workspace.Description = *dto.Description
	}
	if dto.Icon != nil {
		workspace.Icon = *dto.Icon
	}

	return s.repo.Update(ctx, workspace)
}

func (s *workspaceService) DeleteWorkspace(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *workspaceService) AddMember(ctx context.Context, workspaceID, userID, role string) error {
	if role != "owner" && role != "editor" && role != "viewer" {
		return fmt.Errorf("invalid role: %s", role)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	return s.repo.AddMember(ctx, workspaceID, user.ID, role)
}

func (s *workspaceService) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	return s.repo.RemoveMember(ctx, workspaceID, userID)
}

func (s *workspaceService) GetMembers(ctx context.Context, workspaceID string) ([]*entities.User, error) {
	return s.repo.GetMembers(ctx, workspaceID)
}

func (s *workspaceService) UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error {
	if role != "owner" && role != "editor" && role != "viewer" {
		return fmt.Errorf("invalid role: %s", role)
	}

	return s.repo.UpdateMemberRole(ctx, workspaceID, userID, role)
}
