package application

import (
	"context"
	"errors"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type FlowService interface {
	CreateFlow(ctx context.Context, workspaceID uuid.UUID, name, description string) (*domain.Flow, error)
	GetFlowsByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]domain.Flow, error)
}

type flowService struct {
	repo domain.FlowRepository
}

func NewFlowService(repo domain.FlowRepository) FlowService {
	return &flowService{repo: repo}
}

func (s *flowService) CreateFlow(ctx context.Context, workspaceID uuid.UUID, name, description string) (*domain.Flow, error) {
	if name == "" {
		return nil, errors.New("flow name is required")
	}

	flow := &domain.Flow{
		WorkspaceID: workspaceID,
		Name:        name,
		Description: description,
	}

	if err := s.repo.Create(ctx, flow); err != nil {
		return nil, err
	}

	return flow, nil
}

func (s *flowService) GetFlowsByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]domain.Flow, error) {
	return s.repo.ListByWorkspace(ctx, workspaceID)
}
