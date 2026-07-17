package workspace

import (
	"errors"
	"time"

	"github.com/google/uuid" // Assuming you use UUIDs
)

// Service interface defines the business logic
type Service interface {
	CreateWorkspace(req CreateWorkspaceRequest, ownerID string) (*Workspace, error)
	GetWorkspace(id string) (*Workspace, error)
}

type service struct {
	repo Repository
}

// NewService creates a new workspace service instance
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateWorkspace(req CreateWorkspaceRequest, ownerID string) (*Workspace, error) {
	// Apply business logic / data transformations
	now := time.Now()
	newWorkspace := &Workspace{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Persist data via repository
	if err := s.repo.Create(newWorkspace); err != nil {
		return nil, err
	}

	return newWorkspace, nil
}

func (s *service) GetWorkspace(id string) (*Workspace, error) {
	workspace, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, errors.New("workspace not found")
	}
	return workspace, nil
}
