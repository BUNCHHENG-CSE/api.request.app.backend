package flow

import (
	"errors"
	// Import the workspace module ONLY for its exported interfaces/models
	"backend/internal/modules/workspace"
)

type Service interface {
	CreateFlow(workspaceID string, flowName string) error
}

type service struct {
	repo Repository
	// Inject the workspace service here
	workspaceService workspace.Service
}

// NewService now requires the workspace service to be passed in
func NewService(repo Repository, ws workspace.Service) Service {
	return &service{
		repo:             repo,
		workspaceService: ws,
	}
}

func (s *service) CreateFlow(workspaceID string, flowName string) error {
	// 1. Ask the Workspace module if the ID is valid via the injected service
	_, err := s.workspaceService.GetWorkspace(workspaceID)
	if err != nil {
		return errors.New("cannot create flow: invalid workspace")
	}

	// 2. Proceed with flow creation logic...
	return nil
}
