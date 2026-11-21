package repositories

import (
	"context"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"

	"gorm.io/gorm"
)

type workspaceRepo struct {
	db *gorm.DB
}

func NewWorkspaceRepository(db *gorm.DB) domain.WorkspaceRepository {
	return &workspaceRepo{db: db}
}

func (r *workspaceRepo) Create(ctx context.Context, workspace *domain.Workspace) error {
	return r.db.WithContext(ctx).Create(workspace).Error
}

func (r *workspaceRepo) GetByID(ctx context.Context, id, ownerID uuid.UUID) (*domain.Workspace, error) {
	var workspace domain.Workspace
	err := r.db.WithContext(ctx).Where("id = ? AND owner_id = ?", id, ownerID).First(&workspace).Error
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

func (r *workspaceRepo) List(ctx context.Context, ownerID uuid.UUID) ([]domain.Workspace, error) {
	var workspaces []domain.Workspace
	err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Order("created_at ASC").Find(&workspaces).Error
	return workspaces, err
}

func (r *workspaceRepo) Update(ctx context.Context, workspace *domain.Workspace) error {
	return r.db.WithContext(ctx).Save(workspace).Error
}

func (r *workspaceRepo) Delete(ctx context.Context, id, ownerID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&domain.Workspace{}).Where("id = ? AND owner_id = ?", id, ownerID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Where("flow_id IN (?)", tx.Model(&domain.Flow{}).Select("id").Where("workspace_id = ?", id)).Delete(&domain.Request{}).Error; err != nil {
			return err
		}
		if err := tx.Where("workspace_id = ?", id).Delete(&domain.Flow{}).Error; err != nil {
			return err
		}
		if err := tx.Where("workspace_id = ?", id).Delete(&domain.Environment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("workspace_id = ?", id).Delete(&domain.Collection{}).Error; err != nil {
			return err
		}
		if err := tx.Where("workspace_id = ?", id).Delete(&domain.RequestHistory{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND owner_id = ?", id, ownerID).Delete(&domain.Workspace{}).Error
	})
}
