package repositories

import (
	"backend/internal/domain"
	"context"

	"gorm.io/gorm"
)

type environmentRepo struct {
	db *gorm.DB
}

func NewEnvironmentRepository(db *gorm.DB) domain.EnvironmentRepository {
	return &environmentRepo{db: db}
}

func (r *environmentRepo) Create(ctx context.Context, req *domain.Environment) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *environmentRepo) GetByID(ctx context.Context, id uint) (*domain.Environment, error) {
	var environment domain.Environment
	err := r.db.WithContext(ctx).First(&environment, id).Error
	return &environment, err
}

func (r *environmentRepo) ListByWorkspace(ctx context.Context, flowID uint) ([]domain.Environment, error) {
	var environments []domain.Environment
	err := r.db.WithContext(ctx).Where("flow_id = ?", flowID).Find(&environments).Error
	return environments, err
}

func (r *environmentRepo) Update(ctx context.Context, req *domain.Environment) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *environmentRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Environment{}, id).Error
}
