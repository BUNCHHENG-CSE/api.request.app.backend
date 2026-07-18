package repositories

import (
	"backend/internal/domain"
	"context"

	"gorm.io/gorm"
)

type flowRepo struct {
	db *gorm.DB
}

func NewFlowRepository(db *gorm.DB) domain.FlowRepository {
	return &flowRepo{db: db}
}

func (r *flowRepo) Create(ctx context.Context, req *domain.Flow) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *flowRepo) GetByID(ctx context.Context, id uint) (*domain.Flow, error) {
	var flow domain.Flow
	err := r.db.WithContext(ctx).First(&flow, id).Error
	return &flow, err
}

func (r *flowRepo) ListByWorkspace(ctx context.Context, flowID uint) ([]domain.Flow, error) {
	var flows []domain.Flow
	err := r.db.WithContext(ctx).Where("flow_id = ?", flowID).Find(&flows).Error
	return flows, err
}

func (r *flowRepo) Update(ctx context.Context, req *domain.Flow) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *flowRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Flow{}, id).Error
}
