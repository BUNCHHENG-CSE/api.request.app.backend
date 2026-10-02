package repositories

import (
	"context"
	"strings"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type searchRepo struct{ db *gorm.DB }

func NewSearchRepository(db *gorm.DB) domain.SearchRepository { return &searchRepo{db: db} }

func (r *searchRepo) Search(ctx context.Context, ownerID uuid.UUID, workspaceID *uuid.UUID, query string, limit int) ([]domain.SearchResult, error) {
	pattern := "%" + strings.TrimSpace(query) + "%"
	results := make([]domain.SearchResult, 0)
	appendRows := func(db *gorm.DB) error {
		var rows []domain.SearchResult
		if err := db.Limit(limit - len(results)).Scan(&rows).Error; err != nil {
			return err
		}
		results = append(results, rows...)
		return nil
	}
	workspaceFilter := func(db *gorm.DB, column string) *gorm.DB {
		if workspaceID != nil {
			return db.Where(column+" = ?", *workspaceID)
		}
		return db
	}

	flows := r.db.WithContext(ctx).Table("flows f").Select("'flow' AS type, f.id, f.workspace_id, f.name, '' AS method, '' AS url").Joins("JOIN workspaces w ON w.id = f.workspace_id").Where("w.owner_id = ? AND (f.name ILIKE ? OR f.description ILIKE ?)", ownerID, pattern, pattern)
	if err := appendRows(workspaceFilter(flows, "f.workspace_id")); err != nil {
		return nil, err
	}
	if len(results) >= limit {
		return results, nil
	}

	requests := r.db.WithContext(ctx).Table("requests r").Select("'request' AS type, r.id, f.workspace_id, r.name, r.method, r.url").Joins("JOIN flows f ON f.id = r.flow_id").Joins("JOIN workspaces w ON w.id = f.workspace_id").Where("w.owner_id = ? AND (r.name ILIKE ? OR r.url ILIKE ? OR r.method ILIKE ?)", ownerID, pattern, pattern, pattern)
	if err := appendRows(workspaceFilter(requests, "f.workspace_id")); err != nil {
		return nil, err
	}
	if len(results) >= limit {
		return results, nil
	}

	collections := r.db.WithContext(ctx).Table("collections c").Select("'collection' AS type, c.id, c.workspace_id, c.name, '' AS method, '' AS url").Joins("JOIN workspaces w ON w.id = c.workspace_id").Where("w.owner_id = ? AND (c.name ILIKE ? OR c.description ILIKE ?)", ownerID, pattern, pattern)
	if err := appendRows(workspaceFilter(collections, "c.workspace_id")); err != nil {
		return nil, err
	}
	if len(results) >= limit {
		return results, nil
	}

	environments := r.db.WithContext(ctx).Table("environments e").Select("'environment' AS type, e.id, e.workspace_id, e.name, '' AS method, '' AS url").Joins("JOIN workspaces w ON w.id = e.workspace_id").Where("w.owner_id = ? AND e.name ILIKE ?", ownerID, pattern)
	if err := appendRows(workspaceFilter(environments, "e.workspace_id")); err != nil {
		return nil, err
	}
	return results, nil
}
