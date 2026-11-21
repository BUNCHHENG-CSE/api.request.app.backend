package repositories

import (
	"context"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"

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

func (r *collectionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Collection, error) {
	var collection domain.Collection
	err := r.db.WithContext(ctx).First(&collection, id).Error
	return &collection, err
}

func (r *collectionRepo) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]domain.Collection, error) {
	var collections []domain.Collection
	err := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Order("created_at ASC").Find(&collections).Error
	return collections, err
}

func (r *collectionRepo) Update(ctx context.Context, collection *domain.Collection) error {
	return r.db.WithContext(ctx).Save(collection).Error
}

func (r *collectionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Request{}).Where("collection_id = ?", id).Update("collection_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.Collection{}, id).Error
	})
}
