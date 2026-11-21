package application

import (
	"context"
	"errors"
	"strings"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type WorkspaceMemberService interface {
	AddMember(ctx context.Context, requesterID, workspaceID uuid.UUID, identifier string, role domain.WorkspaceRole) (*domain.WorkspaceMember, error)
	ListMembers(ctx context.Context, requesterID, workspaceID uuid.UUID) ([]domain.WorkspaceMember, error)
	UpdateRole(ctx context.Context, requesterID, workspaceID, targetUserID uuid.UUID, role domain.WorkspaceRole) (*domain.WorkspaceMember, error)
	RemoveMember(ctx context.Context, requesterID, workspaceID, targetUserID uuid.UUID) error
}

type workspaceMemberService struct {
	workspaceRepo domain.WorkspaceRepository
	memberRepo    domain.WorkspaceMemberRepository
	userRepo      domain.UserRepository
}

func NewWorkspaceMemberService(
	workspaceRepo domain.WorkspaceRepository,
	memberRepo domain.WorkspaceMemberRepository,
	userRepo domain.UserRepository,
) WorkspaceMemberService {
	return &workspaceMemberService{
		workspaceRepo: workspaceRepo,
		memberRepo:    memberRepo,
		userRepo:      userRepo,
	}
}

func normalizeRole(role domain.WorkspaceRole) domain.WorkspaceRole {
	r := domain.WorkspaceRole(strings.ToLower(strings.TrimSpace(string(role))))
	switch r {
	case domain.RoleAdmin:
		return domain.RoleAdmin
	case domain.RoleEditor:
		return domain.RoleEditor
	default:
		return domain.RoleViewer
	}
}

func (s *workspaceMemberService) checkAdminAccess(ctx context.Context, workspaceID, requesterID uuid.UUID) (*domain.Workspace, error) {
	ws, err := s.workspaceRepo.GetByID(ctx, workspaceID, requesterID)
	if err != nil {
		return nil, err
	}
	if ws.OwnerID == requesterID {
		return ws, nil
	}
	member, err := s.memberRepo.Get(ctx, workspaceID, requesterID)
	if err != nil || member.Role != domain.RoleAdmin {
		return nil, errors.New("unauthorized: admin privileges required")
	}
	return ws, nil
}

func (s *workspaceMemberService) AddMember(ctx context.Context, requesterID, workspaceID uuid.UUID, identifier string, role domain.WorkspaceRole) (*domain.WorkspaceMember, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return nil, errors.New("username or email is required")
	}
	ws, err := s.checkAdminAccess(ctx, workspaceID, requesterID)
	if err != nil {
		return nil, err
	}

	targetUser, err := s.userRepo.GetByUsernameOrEmail(ctx, identifier)
	if err != nil {
		return nil, errors.New("target user not found")
	}

	if targetUser.ID == ws.OwnerID {
		return nil, errors.New("user is already the workspace owner")
	}

	existing, _ := s.memberRepo.Get(ctx, workspaceID, targetUser.ID)
	if existing != nil {
		return nil, errors.New("user is already a member of this workspace")
	}

	validRole := normalizeRole(role)
	member := &domain.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      targetUser.ID,
		Role:        validRole,
		Username:    targetUser.Username,
		Email:       targetUser.Email,
	}

	if err := s.memberRepo.Add(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

func (s *workspaceMemberService) ListMembers(ctx context.Context, requesterID, workspaceID uuid.UUID) ([]domain.WorkspaceMember, error) {
	if _, err := s.workspaceRepo.GetByID(ctx, workspaceID, requesterID); err != nil {
		return nil, err
	}
	return s.memberRepo.List(ctx, workspaceID)
}

func (s *workspaceMemberService) UpdateRole(ctx context.Context, requesterID, workspaceID, targetUserID uuid.UUID, role domain.WorkspaceRole) (*domain.WorkspaceMember, error) {
	ws, err := s.checkAdminAccess(ctx, workspaceID, requesterID)
	if err != nil {
		return nil, err
	}

	if targetUserID == ws.OwnerID {
		return nil, errors.New("cannot change role of workspace owner")
	}

	member, err := s.memberRepo.Get(ctx, workspaceID, targetUserID)
	if err != nil {
		return nil, errors.New("workspace member not found")
	}

	member.Role = normalizeRole(role)
	if err := s.memberRepo.Update(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

func (s *workspaceMemberService) RemoveMember(ctx context.Context, requesterID, workspaceID, targetUserID uuid.UUID) error {
	ws, err := s.workspaceRepo.GetByID(ctx, workspaceID, requesterID)
	if err != nil {
		return err
	}

	if targetUserID == ws.OwnerID {
		return errors.New("cannot remove workspace owner")
	}

	// User can remove themselves (leave workspace)
	if requesterID != targetUserID {
		// Otherwise, requester must be admin
		if _, err := s.checkAdminAccess(ctx, workspaceID, requesterID); err != nil {
			return err
		}
	}

	return s.memberRepo.Delete(ctx, workspaceID, targetUserID)
}
