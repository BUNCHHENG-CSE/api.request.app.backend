package application

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type CookieService interface {
	AddCookie(ctx context.Context, ownerID, workspaceID uuid.UUID, cookie domain.WorkspaceCookie) (*domain.WorkspaceCookie, error)
	ListCookies(ctx context.Context, ownerID, workspaceID uuid.UUID) ([]domain.WorkspaceCookie, error)
	DeleteCookie(ctx context.Context, ownerID, workspaceID, cookieID uuid.UUID) error
	ClearCookies(ctx context.Context, ownerID, workspaceID uuid.UUID) error
	AttachCookiesToRequest(ctx context.Context, workspaceID uuid.UUID, targetURL string, req *http.Request) error
	SaveCookiesFromResponse(ctx context.Context, workspaceID uuid.UUID, targetURL string, headers http.Header) error
}

type cookieService struct {
	cookieRepo    domain.WorkspaceCookieRepository
	workspaceRepo domain.WorkspaceRepository
}

func NewCookieService(cookieRepo domain.WorkspaceCookieRepository, workspaceRepo domain.WorkspaceRepository) CookieService {
	return &cookieService{
		cookieRepo:    cookieRepo,
		workspaceRepo: workspaceRepo,
	}
}

func (s *cookieService) AddCookie(ctx context.Context, ownerID, workspaceID uuid.UUID, cookie domain.WorkspaceCookie) (*domain.WorkspaceCookie, error) {
	if _, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID); err != nil {
		return nil, err
	}
	cookie.Name = strings.TrimSpace(cookie.Name)
	if cookie.Name == "" {
		return nil, errors.New("cookie name is required")
	}
	cookie.Domain = strings.TrimSpace(cookie.Domain)
	if cookie.Domain == "" {
		return nil, errors.New("cookie domain is required")
	}
	if cookie.Path == "" {
		cookie.Path = "/"
	}
	cookie.WorkspaceID = workspaceID
	if err := s.cookieRepo.Save(ctx, &cookie); err != nil {
		return nil, err
	}
	return &cookie, nil
}

func (s *cookieService) ListCookies(ctx context.Context, ownerID, workspaceID uuid.UUID) ([]domain.WorkspaceCookie, error) {
	if _, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID); err != nil {
		return nil, err
	}
	return s.cookieRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *cookieService) DeleteCookie(ctx context.Context, ownerID, workspaceID, cookieID uuid.UUID) error {
	if _, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID); err != nil {
		return err
	}
	return s.cookieRepo.Delete(ctx, cookieID)
}

func (s *cookieService) ClearCookies(ctx context.Context, ownerID, workspaceID uuid.UUID) error {
	if _, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID); err != nil {
		return err
	}
	return s.cookieRepo.ClearByWorkspace(ctx, workspaceID)
}

func (s *cookieService) AttachCookiesToRequest(ctx context.Context, workspaceID uuid.UUID, targetURL string, req *http.Request) error {
	parsed, err := url.Parse(targetURL)
	if err != nil || parsed.Host == "" {
		return nil
	}

	host := parsed.Hostname()
	path := parsed.Path
	if path == "" {
		path = "/"
	}

	cookies, err := s.cookieRepo.GetMatching(ctx, workspaceID, host, path)
	if err != nil || len(cookies) == 0 {
		return nil
	}

	for _, c := range cookies {
		if c.ExpiresAt != nil && c.ExpiresAt.Before(time.Now()) {
			continue
		}
		if c.Secure && parsed.Scheme != "https" {
			continue
		}
		req.AddCookie(&http.Cookie{
			Name:  c.Name,
			Value: c.Value,
		})
	}

	return nil
}

func (s *cookieService) SaveCookiesFromResponse(ctx context.Context, workspaceID uuid.UUID, targetURL string, headers http.Header) error {
	parsed, err := url.Parse(targetURL)
	if err != nil || parsed.Host == "" {
		return nil
	}
	defaultDomain := parsed.Hostname()
	defaultPath := parsed.Path
	if defaultPath == "" {
		defaultPath = "/"
	}

	rawResp := http.Response{Header: headers}
	parsedCookies := rawResp.Cookies()

	for _, c := range parsedCookies {
		dom := c.Domain
		if dom == "" {
			dom = defaultDomain
		}
		cPath := c.Path
		if cPath == "" {
			cPath = defaultPath
		}

		var expiresAt *time.Time
		if !c.Expires.IsZero() {
			expiresAt = &c.Expires
		}

		cookieEntity := domain.WorkspaceCookie{
			WorkspaceID: workspaceID,
			Domain:      dom,
			Path:        cPath,
			Name:        c.Name,
			Value:       c.Value,
			ExpiresAt:   expiresAt,
			Secure:      c.Secure,
			HttpOnly:    c.HttpOnly,
		}
		_ = s.cookieRepo.Save(ctx, &cookieEntity)
	}

	return nil
}
