package repositories

import (
	"context"
	"strings"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type cookieRepo struct {
	db *gorm.DB
}

func NewCookieRepository(db *gorm.DB) domain.WorkspaceCookieRepository {
	return &cookieRepo{db: db}
}

func (r *cookieRepo) Save(ctx context.Context, cookie *domain.WorkspaceCookie) error {
	var existing domain.WorkspaceCookie
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND domain = ? AND path = ? AND name = ?",
			cookie.WorkspaceID, cookie.Domain, cookie.Path, cookie.Name).
		First(&existing).Error

	if err == nil {
		existing.Value = cookie.Value
		existing.ExpiresAt = cookie.ExpiresAt
		existing.Secure = cookie.Secure
		existing.HttpOnly = cookie.HttpOnly
		return r.db.WithContext(ctx).Save(&existing).Error
	}

	return r.db.WithContext(ctx).Create(cookie).Error
}

func (r *cookieRepo) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]domain.WorkspaceCookie, error) {
	var cookies []domain.WorkspaceCookie
	err := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("domain ASC, name ASC").
		Find(&cookies).Error
	return cookies, err
}

func (r *cookieRepo) GetMatching(ctx context.Context, workspaceID uuid.UUID, host, path string) ([]domain.WorkspaceCookie, error) {
	var cookies []domain.WorkspaceCookie
	host = strings.ToLower(host)

	// Fetch all cookies for workspace and match in Go for bulletproof domain/path semantics
	err := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Find(&cookies).Error
	if err != nil {
		return nil, err
	}

	matched := make([]domain.WorkspaceCookie, 0)
	for _, c := range cookies {
		cookieDomain := strings.TrimPrefix(strings.ToLower(c.Domain), ".")
		domainMatch := host == cookieDomain || strings.HasSuffix(host, "."+cookieDomain)
		if !domainMatch {
			continue
		}
		cookiePath := c.Path
		if cookiePath == "" {
			cookiePath = "/"
		}
		pathMatch := strings.HasPrefix(path, cookiePath) || (path == "" && cookiePath == "/")
		if pathMatch {
			matched = append(matched, c)
		}
	}

	return matched, nil
}

func (r *cookieRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.WorkspaceCookie{}, id).Error
}

func (r *cookieRepo) ClearByWorkspace(ctx context.Context, workspaceID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Delete(&domain.WorkspaceCookie{}).Error
}

func (r *cookieRepo) DeleteByDomain(ctx context.Context, workspaceID uuid.UUID, dom string) error {
	return r.db.WithContext(ctx).Where("workspace_id = ? AND domain = ?", workspaceID, dom).Delete(&domain.WorkspaceCookie{}).Error
}
