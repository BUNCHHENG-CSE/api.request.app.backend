package services

import (
	"context"

	"backend/internal/entities"
	"backend/internal/models"
	"backend/internal/repositories"
	"encoding/json"
)

// FlowService handles flow operations
type FlowService interface {
	CreateFlow(ctx context.Context, workspaceID string, dto *models.CreateFlowDTO) (*entities.Flow, error)
	GetFlow(ctx context.Context, id string) (*entities.Flow, error)
	ListFlows(ctx context.Context, workspaceID string) ([]*entities.Flow, error)
	UpdateFlow(ctx context.Context, id string, dto *models.UpdateFlowDTO) (*entities.Flow, error)
	DeleteFlow(ctx context.Context, id string) error
	ExecuteFlow(ctx context.Context, flowID string, executionCtx map[string]interface{}) (map[string]interface{}, error)
}

type flowService struct {
	repo           repositories.FlowRepository
	requestService RequestService
}

func NewFlowService(
	repo repositories.FlowRepository,
	requestService RequestService,
) FlowService {
	return &flowService{
		repo:           repo,
		requestService: requestService,
	}
}

func (s *flowService) CreateFlow(ctx context.Context, workspaceID string, dto *models.CreateFlowDTO) (*entities.Flow, error) {
	nodesJSON, err := json.Marshal(dto.Nodes)
	if err != nil {
		return nil, err
	}

	edgesJSON, err := json.Marshal(dto.Edges)
	if err != nil {
		return nil, err
	}

	flow := &entities.Flow{
		WorkspaceID: workspaceID,
		Name:        dto.Name,
		Description: dto.Description,
		Nodes:       nodesJSON,
		Edges:       edgesJSON,
	}

	return s.repo.Create(ctx, flow)
}

func (s *flowService) GetFlow(ctx context.Context, id string) (*entities.Flow, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *flowService) ListFlows(ctx context.Context, workspaceID string) ([]*entities.Flow, error) {
	return s.repo.ListByWorkspace(ctx, workspaceID)
}

func (s *flowService) UpdateFlow(ctx context.Context, id string, dto *models.UpdateFlowDTO) (*entities.Flow, error) {
	flow, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dto.Name != nil {
		flow.Name = *dto.Name
	}
	if dto.Description != nil {
		flow.Description = *dto.Description
	}

	if dto.Nodes != nil {
		nodesJSON, err := json.Marshal(dto.Nodes)
		if err != nil {
			return nil, err
		}

		flow.Nodes = nodesJSON
	}

	if dto.Edges != nil {
		edgesJSON, err := json.Marshal(dto.Edges)
		if err != nil {
			return nil, err
		}
		flow.Edges = edgesJSON
	}

	return s.repo.Update(ctx, flow)
}

func (s *flowService) DeleteFlow(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *flowService) ExecuteFlow(ctx context.Context, flowID string, executionCtx map[string]interface{}) (map[string]interface{}, error) {
	flow, err := s.repo.GetByID(ctx, flowID)
	if err != nil {
		return nil, err
	}

	results := make(map[string]interface{})
	var nodes []entities.FlowNode
	if err := json.Unmarshal(flow.Nodes, &nodes); err != nil {
		return nil, err // Or handle the parsing error as needed
	}

	// Parse nodes
	for _, node := range nodes {
		// Execute request nodes
		if node.Type == "request" && node.RequestID != nil {
			dto := &models.ExecuteRequestDTO{
				RequestID: *node.RequestID,
			}

			resp, err := s.requestService.ExecuteRequest(ctx, dto)
			if err != nil {
				results[node.ID] = map[string]interface{}{
					"error": err.Error(),
				}
				continue
			}

			results[node.ID] = resp
			executionCtx[node.ID] = resp
		}
	}

	return results, nil
}
