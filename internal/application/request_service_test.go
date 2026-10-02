package application

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type requestRepoMemory struct{ request *domain.Request }

func (r *requestRepoMemory) Create(_ context.Context, req *domain.Request) error {
	req.ID = uuid.New()
	r.request = req
	return nil
}
func (r *requestRepoMemory) GetByID(_ context.Context, id uuid.UUID) (*domain.Request, error) {
	if r.request == nil || r.request.ID != id {
		return nil, errors.New("request not found")
	}
	return r.request, nil
}
func (r *requestRepoMemory) ListByFlow(_ context.Context, flowID uuid.UUID) ([]domain.Request, error) {
	if r.request != nil && r.request.FlowID == flowID {
		return []domain.Request{*r.request}, nil
	}
	return []domain.Request{}, nil
}
func (r *requestRepoMemory) Update(_ context.Context, req *domain.Request) error {
	r.request = req
	return nil
}
func (r *requestRepoMemory) Delete(_ context.Context, id uuid.UUID) error {
	r.request = nil
	return nil
}

type flowRepoMemory struct{ flow *domain.Flow }

func (r *flowRepoMemory) Create(_ context.Context, flow *domain.Flow) error {
	r.flow = flow
	return nil
}
func (r *flowRepoMemory) GetByID(_ context.Context, id uuid.UUID) (*domain.Flow, error) {
	if r.flow == nil || r.flow.ID != id {
		return nil, errors.New("flow not found")
	}
	return r.flow, nil
}
func (r *flowRepoMemory) ListByWorkspace(context.Context, uuid.UUID) ([]domain.Flow, error) {
	if r.flow == nil {
		return []domain.Flow{}, nil
	}
	return []domain.Flow{*r.flow}, nil
}
func (r *flowRepoMemory) Update(context.Context, *domain.Flow) error { return nil }
func (r *flowRepoMemory) Delete(context.Context, uuid.UUID) error    { return nil }

type workspaceRepoMemory struct{ workspace *domain.Workspace }

func (r *workspaceRepoMemory) Create(context.Context, *domain.Workspace) error { return nil }
func (r *workspaceRepoMemory) GetByID(_ context.Context, id, ownerID uuid.UUID) (*domain.Workspace, error) {
	if r.workspace == nil || r.workspace.ID != id || r.workspace.OwnerID != ownerID {
		return nil, errors.New("workspace not found")
	}
	return r.workspace, nil
}
func (r *workspaceRepoMemory) List(context.Context, uuid.UUID) ([]domain.Workspace, error) {
	if r.workspace == nil {
		return []domain.Workspace{}, nil
	}
	return []domain.Workspace{*r.workspace}, nil
}
func (r *workspaceRepoMemory) Update(context.Context, *domain.Workspace) error    { return nil }
func (r *workspaceRepoMemory) Delete(context.Context, uuid.UUID, uuid.UUID) error { return nil }

type collectionRepoMemory struct{ collection *domain.Collection }

func (r *collectionRepoMemory) Create(context.Context, *domain.Collection) error { return nil }
func (r *collectionRepoMemory) GetByID(_ context.Context, id uuid.UUID) (*domain.Collection, error) {
	if r.collection == nil || r.collection.ID != id {
		return nil, errors.New("collection not found")
	}
	return r.collection, nil
}
func (r *collectionRepoMemory) ListByWorkspace(context.Context, uuid.UUID) ([]domain.Collection, error) {
	if r.collection == nil {
		return []domain.Collection{}, nil
	}
	return []domain.Collection{*r.collection}, nil
}
func (r *collectionRepoMemory) Update(context.Context, *domain.Collection) error { return nil }
func (r *collectionRepoMemory) Delete(context.Context, uuid.UUID) error          { return nil }

type environmentRepoMemory struct{ environment *domain.Environment }

func (r *environmentRepoMemory) Create(context.Context, *domain.Environment) error { return nil }
func (r *environmentRepoMemory) GetByID(_ context.Context, id uuid.UUID) (*domain.Environment, error) {
	if r.environment == nil || r.environment.ID != id {
		return nil, errors.New("environment not found")
	}
	return r.environment, nil
}
func (r *environmentRepoMemory) ListByWorkspace(context.Context, uuid.UUID) ([]domain.Environment, error) {
	if r.environment == nil {
		return []domain.Environment{}, nil
	}
	return []domain.Environment{*r.environment}, nil
}
func (r *environmentRepoMemory) Update(context.Context, *domain.Environment) error { return nil }
func (r *environmentRepoMemory) Delete(context.Context, uuid.UUID) error           { return nil }

type historyRepoMemory struct{ history *domain.RequestHistory }

func (r *historyRepoMemory) Create(_ context.Context, history *domain.RequestHistory) error {
	history.ID = uuid.New()
	r.history = history
	return nil
}
func (r *historyRepoMemory) GetByID(_ context.Context, id, userID uuid.UUID) (*domain.RequestHistory, error) {
	if r.history == nil || r.history.ID != id || r.history.UserID != userID {
		return nil, errors.New("history not found")
	}
	return r.history, nil
}
func (r *historyRepoMemory) List(context.Context, uuid.UUID, *uuid.UUID, int, int) ([]domain.RequestHistory, int64, error) {
	return nil, 0, nil
}
func (r *historyRepoMemory) DeleteByUser(context.Context, uuid.UUID) error { return nil }

func TestRequestServiceExecuteAndOwnership(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/users" {
			t.Errorf("environment variable was not applied to URL: %s", req.URL.Path)
		}
		if req.URL.Query().Get("limit") != "20" {
			t.Errorf("query parameter was not applied")
		}
		if req.Header.Get("X-Test") != "yes" {
			t.Errorf("header was not applied")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	ownerID, workspaceID, flowID, environmentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	requests := &requestRepoMemory{}
	flows := &flowRepoMemory{flow: &domain.Flow{ID: flowID, WorkspaceID: workspaceID}}
	workspaces := &workspaceRepoMemory{workspace: &domain.Workspace{ID: workspaceID, OwnerID: ownerID}}
	histories := &historyRepoMemory{}
	environments := &environmentRepoMemory{environment: &domain.Environment{ID: environmentID, WorkspaceID: workspaceID, Variables: `{"path":"users","token":"yes"}`}}
	service := NewRequestService(requests, flows, workspaces, &collectionRepoMemory{}, environments, histories, true)

	stored, err := service.CreateRequest(context.Background(), ownerID, RequestInput{
		FlowID: flowID, Method: http.MethodGet, URL: server.URL + "/{{path}}",
		Headers: `{"X-Test":"{{token}}"}`, QueryParams: `{"limit":"20"}`,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := service.ExecuteRequest(context.Background(), ownerID, stored.ID, &environmentID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != http.StatusOK || result.Body != `{"ok":true}` {
		t.Fatalf("unexpected result: %+v", result)
	}
	if histories.history == nil || result.HistoryID == uuid.Nil {
		t.Fatal("execution history was not persisted")
	}

	if _, err := service.GetRequest(context.Background(), uuid.New(), stored.ID); err == nil {
		t.Fatal("a different owner was allowed to access the request")
	}
}
