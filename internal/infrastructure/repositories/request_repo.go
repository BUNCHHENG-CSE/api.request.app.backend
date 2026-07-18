package repositories

import (
	"context"

	"backend/internal/domain"

	"gorm.io/gorm"
)

type requestRepo struct {
	db *gorm.DB
}

func NewRequestRepository(db *gorm.DB) domain.RequestRepository {
	return &requestRepo{db: db}
}

func (r *requestRepo) Create(ctx context.Context, req *domain.Request) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *requestRepo) GetByID(ctx context.Context, id uint) (*domain.Request, error) {
	var req domain.Request
	err := r.db.WithContext(ctx).First(&req, id).Error
	return &req, err
}

func (r *requestRepo) ListByFlow(ctx context.Context, flowID uint) ([]domain.Request, error) {
	var requests []domain.Request
	err := r.db.WithContext(ctx).Where("flow_id = ?", flowID).Find(&requests).Error
	return requests, err
}

func (r *requestRepo) Update(ctx context.Context, req *domain.Request) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *requestRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Request{}, id).Error
}
