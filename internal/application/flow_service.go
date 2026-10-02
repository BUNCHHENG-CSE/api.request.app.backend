package application

import (
	"context"
	"errors"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type FlowService interface {
	CreateFlow(ctx context.Context, ownerID, workspaceID uuid.UUID, name, description, nodes, edges string) (*domain.Flow, error)
	GetFlow(ctx context.Context, ownerID, id uuid.UUID) (*domain.Flow, error)
	GetFlowsByWorkspace(ctx context.Context, ownerID, workspaceID uuid.UUID) ([]domain.Flow, error)
	UpdateFlow(ctx context.Context, ownerID, id uuid.UUID, name, description, nodes, edges *string) (*domain.Flow, error)
	DeleteFlow(ctx context.Context, ownerID, id uuid.UUID) error
}

type flowService struct {
	repo          domain.FlowRepository
	workspaceRepo domain.WorkspaceRepository
}

func NewFlowService(repo domain.FlowRepository, workspaceRepo domain.WorkspaceRepository) FlowService {
	return &flowService{repo: repo, workspaceRepo: workspaceRepo}
}

func (s *flowService) authorizeWorkspace(ctx context.Context, ownerID, workspaceID uuid.UUID) error {
	_, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID)
	return err
}

func (s *flowService) authorizeFlow(ctx context.Context, ownerID, id uuid.UUID) (*domain.Flow, error) {
	flow, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeWorkspace(ctx, ownerID, flow.WorkspaceID); err != nil {
		return nil, err
	}
	return flow, nil
}

func (s *flowService) CreateFlow(ctx context.Context, ownerID, workspaceID uuid.UUID, name, description, nodes, edges string) (*domain.Flow, error) {
	if name == "" {
		return nil, errors.New("flow name is required")
	}
	if err := s.authorizeWorkspace(ctx, ownerID, workspaceID); err != nil {
		return nil, err
	}
	flow := &domain.Flow{WorkspaceID: workspaceID, Name: name, Description: description, Nodes: nodes, Edges: edges}
	if err := s.repo.Create(ctx, flow); err != nil {
		return nil, err
	}
	return flow, nil
}

func (s *flowService) GetFlow(ctx context.Context, ownerID, id uuid.UUID) (*domain.Flow, error) {
	return s.authorizeFlow(ctx, ownerID, id)
}

func (s *flowService) GetFlowsByWorkspace(ctx context.Context, ownerID, workspaceID uuid.UUID) ([]domain.Flow, error) {
	if err := s.authorizeWorkspace(ctx, ownerID, workspaceID); err != nil {
		return nil, err
	}
	return s.repo.ListByWorkspace(ctx, workspaceID)
}

func (s *flowService) UpdateFlow(ctx context.Context, ownerID, id uuid.UUID, name, description, nodes, edges *string) (*domain.Flow, error) {
	flow, err := s.authorizeFlow(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		if *name == "" {
			return nil, errors.New("flow name cannot be empty")
		}
		flow.Name = *name
	}
	if description != nil {
		flow.Description = *description
	}
	if nodes != nil {
		flow.Nodes = *nodes
	}
	if edges != nil {
		flow.Edges = *edges
	}
	if err := s.repo.Update(ctx, flow); err != nil {
		return nil, err
	}
	return flow, nil
}

func (s *flowService) DeleteFlow(ctx context.Context, ownerID, id uuid.UUID) error {
	if _, err := s.authorizeFlow(ctx, ownerID, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
