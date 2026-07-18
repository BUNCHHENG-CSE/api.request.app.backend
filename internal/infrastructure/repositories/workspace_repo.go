package repositories

import (
	"backend/internal/domain"
	"context"

	"gorm.io/gorm"
)

type workspaceRepo struct {
	db *gorm.DB
}

func NewWorkspaceRepository(db *gorm.DB) domain.WorkspaceRepository {
	return &workspaceRepo{db: db}
}

func (r *workspaceRepo) Create(ctx context.Context, workspace *domain.Workspace) error {
	// GORM implementation
	return r.db.WithContext(ctx).Create(workspace).Error
}

func (r *workspaceRepo) GetByID(ctx context.Context, id uint) (*domain.Workspace, error) {
	var workspace domain.Workspace
	err := r.db.WithContext(ctx).First(&workspace, id).Error
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

func (r *workspaceRepo) List(ctx context.Context) ([]domain.Workspace, error) {
	var workspaces []domain.Workspace
	err := r.db.WithContext(ctx).Find(&workspaces).Error
	return workspaces, err
}
