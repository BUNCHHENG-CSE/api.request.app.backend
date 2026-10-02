package repositories

import (
	"context"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type historyRepo struct{ db *gorm.DB }

func NewRequestHistoryRepository(db *gorm.DB) domain.RequestHistoryRepository {
	return &historyRepo{db: db}
}

func (r *historyRepo) Create(ctx context.Context, history *domain.RequestHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *historyRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.RequestHistory, error) {
	var history domain.RequestHistory
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&history).Error
	return &history, err
}

func (r *historyRepo) List(ctx context.Context, userID uuid.UUID, workspaceID *uuid.UUID, limit, offset int) ([]domain.RequestHistory, int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.RequestHistory{}).Where("user_id = ?", userID)
	if workspaceID != nil {
		query = query.Where("workspace_id = ?", *workspaceID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var histories []domain.RequestHistory
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&histories).Error
	return histories, total, err
}

func (r *historyRepo) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&domain.RequestHistory{}).Error
}
