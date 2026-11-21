package repositories

import (
	"context"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type monitorRepo struct {
	db *gorm.DB
}

func NewMonitorRepository(db *gorm.DB) domain.MonitorRepository {
	return &monitorRepo{db: db}
}

func (r *monitorRepo) Create(ctx context.Context, monitor *domain.Monitor) error {
	return r.db.WithContext(ctx).Create(monitor).Error
}

func (r *monitorRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Monitor, error) {
	var monitor domain.Monitor
	if err := r.db.WithContext(ctx).First(&monitor, id).Error; err != nil {
		return nil, err
	}
	return &monitor, nil
}

func (r *monitorRepo) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]domain.Monitor, error) {
	var monitors []domain.Monitor
	err := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").
		Find(&monitors).Error
	return monitors, err
}

func (r *monitorRepo) Update(ctx context.Context, monitor *domain.Monitor) error {
	return r.db.WithContext(ctx).Save(monitor).Error
}

func (r *monitorRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("monitor_id = ?", id).Delete(&domain.MonitorRunLog{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.Monitor{}, id).Error
	})
}

func (r *monitorRepo) CreateRunLog(ctx context.Context, log *domain.MonitorRunLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *monitorRepo) ListRunLogs(ctx context.Context, monitorID uuid.UUID, limit int) ([]domain.MonitorRunLog, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var logs []domain.MonitorRunLog
	err := r.db.WithContext(ctx).
		Where("monitor_id = ?", monitorID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
