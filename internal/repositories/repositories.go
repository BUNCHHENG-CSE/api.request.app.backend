package repositories

import (
	"context"

	"backend/internal/entities"
)

// RequestRepository defines data access methods for requests
type RequestRepository interface {
	Create(ctx context.Context, req *entities.RequestTab) (*entities.RequestTab, error)
	GetByID(ctx context.Context, id string) (*entities.RequestTab, error)
	ListByCollection(ctx context.Context, collectionID string) ([]*entities.RequestTab, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]*entities.RequestTab, error)
	Update(ctx context.Context, req *entities.RequestTab) (*entities.RequestTab, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, workspaceID string, query string) ([]*entities.RequestTab, error)
}

// CollectionRepository defines data access methods for collections
type CollectionRepository interface {
	Create(ctx context.Context, col *entities.Collection) (*entities.Collection, error)
	GetByID(ctx context.Context, id string) (*entities.Collection, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]*entities.Collection, error)
	Update(ctx context.Context, col *entities.Collection) (*entities.Collection, error)
	Delete(ctx context.Context, id string) error
	UpdateOrder(ctx context.Context, collectionIDs []string) error
}

// EnvironmentRepository defines data access methods for environments
type EnvironmentRepository interface {
	Create(ctx context.Context, env *entities.Environment) (*entities.Environment, error)
	GetByID(ctx context.Context, id string) (*entities.Environment, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]*entities.Environment, error)
	Update(ctx context.Context, env *entities.Environment) (*entities.Environment, error)
	Delete(ctx context.Context, id string) error
}

// WorkspaceRepository defines data access methods for workspaces
type WorkspaceRepository interface {
	Create(ctx context.Context, ws *entities.Workspace) error
	GetByID(ctx context.Context, id string) (*entities.Workspace, error)
	ListByUser(ctx context.Context, userID string) ([]*entities.Workspace, error)
	Update(ctx context.Context, ws *entities.Workspace) (*entities.Workspace, error)
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, workspaceID, userID, role string) error
	RemoveMember(ctx context.Context, workspaceID, userID string) error
	GetMembers(ctx context.Context, workspaceID string) ([]*entities.User, error)
	UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error
}

// UserRepository defines data access methods for users
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) (*entities.User, error)
	GetByID(ctx context.Context, id string) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) (*entities.User, error)
	Delete(ctx context.Context, id string) error
}

// ResponseRepository defines data access methods for responses
type ResponseRepository interface {
	Create(ctx context.Context, resp *entities.Response) (*entities.Response, error)
	GetByID(ctx context.Context, id string) (*entities.Response, error)
	ListByRequest(ctx context.Context, requestID string, limit int) ([]*entities.Response, error)
	Delete(ctx context.Context, id string) error
}

// HistoryRepository defines data access methods for request history
type HistoryRepository interface {
	Create(ctx context.Context, history *entities.RequestHistory) (*entities.RequestHistory, error)
	ListByRequest(ctx context.Context, requestID string, limit int) ([]*entities.RequestHistory, error)
	ListByWorkspace(ctx context.Context, workspaceID string, limit int) ([]*entities.RequestHistory, error)
	Delete(ctx context.Context, id string) error
}

// FlowRepository defines data access methods for flows
type FlowRepository interface {
	Create(ctx context.Context, flow *entities.Flow) (*entities.Flow, error)
	GetByID(ctx context.Context, id string) (*entities.Flow, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]*entities.Flow, error)
	Update(ctx context.Context, flow *entities.Flow) (*entities.Flow, error)
	Delete(ctx context.Context, id string) error
}

// SpecRepository defines data access methods for specs
type SpecRepository interface {
	Create(ctx context.Context, spec *entities.Spec) (*entities.Spec, error)
	GetByID(ctx context.Context, id string) (*entities.Spec, error)
	ListByWorkspace(ctx context.Context, workspaceID string) ([]*entities.Spec, error)
	Update(ctx context.Context, spec *entities.Spec) (*entities.Spec, error)
	Delete(ctx context.Context, id string) error
}
