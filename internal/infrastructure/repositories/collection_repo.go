package repositories

import (
	"context"

	"backend/internal/domain"

	"gorm.io/gorm"
)

type collectionRepo struct {
	db *gorm.DB
}

func NewCollectionRepository(db *gorm.DB) domain.CollectionRepository {
	return &collectionRepo{db: db}
}

func (r *collectionRepo) Create(ctx context.Context, collection *domain.Collection) error {
	return r.db.WithContext(ctx).Create(collection).Error
}

func (r *collectionRepo) GetByID(ctx context.Context, id uint) (*domain.Collection, error) {
	var collection domain.Collection
	err := r.db.WithContext(ctx).First(&collection, id).Error
	return &collection, err
}

func (r *collectionRepo) ListByWorkspace(ctx context.Context, workspaceID uint) ([]domain.Collection, error) {
	var collections []domain.Collection
	err := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Find(&collections).Error
	return collections, err
}

func (r *collectionRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Collection{}, id).Error
}
