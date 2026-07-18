package application

import (
	"context"
	"errors"

	"backend/internal/domain"
)

type RequestService interface {
	CreateRequest(ctx context.Context, flowID uint, method, url, headers, body string) (*domain.Request, error)
	GetRequestsByFlow(ctx context.Context, flowID uint) ([]domain.Request, error)
}

type requestService struct {
	repo domain.RequestRepository
}

func NewRequestService(repo domain.RequestRepository) RequestService {
	return &requestService{repo: repo}
}

func (s *requestService) CreateRequest(ctx context.Context, flowID uint, method, url, headers, body string) (*domain.Request, error) {
	if method == "" || url == "" {
		return nil, errors.New("method and url are required")
	}

	req := &domain.Request{
		FlowID:  flowID,
		Method:  method,
		URL:     url,
		Headers: headers,
		Body:    body,
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return req, nil
}

func (s *requestService) GetRequestsByFlow(ctx context.Context, flowID uint) ([]domain.Request, error) {
	return s.repo.ListByFlow(ctx, flowID)
}
