package repositories

import (
	"context"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type activityRepo struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) domain.WorkspaceActivityRepository {
	return &activityRepo{db: db}
}

func (r *activityRepo) Create(ctx context.Context, activity *domain.WorkspaceActivity) error {
	return r.db.WithContext(ctx).Create(activity).Error
}

func (r *activityRepo) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit int) ([]domain.WorkspaceActivity, error) {
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}
	var activities []domain.WorkspaceActivity
	err := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").
		Limit(limit).
		Find(&activities).Error
	return activities, err
}
