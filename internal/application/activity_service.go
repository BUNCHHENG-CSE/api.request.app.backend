package application

import (
	"context"
	"time"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type ActivityService interface {
	Log(ctx context.Context, workspaceID, userID uuid.UUID, username, action, details string)
	List(ctx context.Context, ownerID, workspaceID uuid.UUID, limit int) ([]domain.WorkspaceActivity, error)
}

type activityService struct {
	activityRepo  domain.WorkspaceActivityRepository
	workspaceRepo domain.WorkspaceRepository
}

func NewActivityService(activityRepo domain.WorkspaceActivityRepository, workspaceRepo domain.WorkspaceRepository) ActivityService {
	return &activityService{
		activityRepo:  activityRepo,
		workspaceRepo: workspaceRepo,
	}
}

func (s *activityService) Log(ctx context.Context, workspaceID, userID uuid.UUID, username, action, details string) {
	activity := &domain.WorkspaceActivity{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Username:    username,
		Action:      action,
		Details:     details,
		CreatedAt:   time.Now(),
	}
	_ = s.activityRepo.Create(ctx, activity)
}

func (s *activityService) List(ctx context.Context, ownerID, workspaceID uuid.UUID, limit int) ([]domain.WorkspaceActivity, error) {
	if _, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID); err != nil {
		return nil, err
	}
	return s.activityRepo.ListByWorkspace(ctx, workspaceID, limit)
}
