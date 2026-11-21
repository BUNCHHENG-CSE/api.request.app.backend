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

type requestRepoMemory struct {
	items map[uuid.UUID]*domain.Request
}

func newRequestRepoMemory() *requestRepoMemory {
	return &requestRepoMemory{items: make(map[uuid.UUID]*domain.Request)}
}

func (r *requestRepoMemory) Create(_ context.Context, req *domain.Request) error {
	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}
	copy := *req
	r.items[req.ID] = &copy
	return nil
}
func (r *requestRepoMemory) GetByID(_ context.Context, id uuid.UUID) (*domain.Request, error) {
	req, ok := r.items[id]
	if !ok || req == nil {
		return nil, errors.New("request not found")
	}
	return req, nil
}
func (r *requestRepoMemory) ListByFlow(_ context.Context, flowID uuid.UUID) ([]domain.Request, error) {
	var list []domain.Request
	for _, req := range r.items {
		if req.FlowID == flowID {
			list = append(list, *req)
		}
	}
	return list, nil
}
func (r *requestRepoMemory) Update(_ context.Context, req *domain.Request) error {
	copy := *req
	r.items[req.ID] = &copy
	return nil
}
func (r *requestRepoMemory) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.items, id)
	return nil
}

type flowRepoMemory struct{ flow *domain.Flow }

func (r *flowRepoMemory) Create(_ context.Context, flow *domain.Flow) error {
	if flow.ID == uuid.Nil {
		flow.ID = uuid.New()
	}
	r.flow = flow
	return nil
}
func (r *flowRepoMemory) GetByID(_ context.Context, id uuid.UUID) (*domain.Flow, error) {
	if r.flow == nil || r.flow.ID != id {
		return nil, errors.New("flow not found")
	}
	return r.flow, nil
}
func (r *flowRepoMemory) GetByWebhookToken(_ context.Context, token string) (*domain.Flow, error) {
	if r.flow == nil || r.flow.WebhookToken != token {
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

func (r *collectionRepoMemory) Create(_ context.Context, col *domain.Collection) error {
	if col.ID == uuid.Nil {
		col.ID = uuid.New()
	}
	r.collection = col
	return nil
}
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
	requests := newRequestRepoMemory()
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

func TestRequestServiceAuthAndVariableWithSpaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer secret-token-123" {
			t.Errorf("auth header not applied properly: %s", req.Header.Get("Authorization"))
		}
		if req.Header.Get("Content-Type") != "application/json" {
			t.Errorf("auto Content-Type not applied: %s", req.Header.Get("Content-Type"))
		}
		if req.URL.Query().Get("page") != "1" {
			t.Errorf("array query parameter not applied: %s", req.URL.Query().Get("page"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"authorized"}`))
	}))
	defer server.Close()

	ownerID, workspaceID, flowID, envID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	requests := newRequestRepoMemory()
	flows := &flowRepoMemory{flow: &domain.Flow{ID: flowID, WorkspaceID: workspaceID}}
	workspaces := &workspaceRepoMemory{workspace: &domain.Workspace{ID: workspaceID, OwnerID: ownerID}}
	histories := &historyRepoMemory{}
	environments := &environmentRepoMemory{environment: &domain.Environment{
		ID:          envID,
		WorkspaceID: workspaceID,
		Variables:   `{"my_token":"secret-token-123"}`,
	}}
	service := NewRequestService(requests, flows, workspaces, &collectionRepoMemory{}, environments, histories, true)

	// Note spaces in {{ my_token }} and array query params
	stored, err := service.CreateRequest(context.Background(), ownerID, RequestInput{
		FlowID:      flowID,
		Method:      http.MethodPost,
		URL:         server.URL,
		QueryParams: `[{"key": "page", "value": 1, "enabled": true}, {"key": "disabled", "value": "x", "enabled": false}]`,
		Body:        `{"hello":"world"}`,
		BodyType:    "json",
		Auth:        `{"type":"bearer","token":"{{ my_token }}"}`,
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := service.ExecuteRequest(context.Background(), ownerID, stored.ID, &envID)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Status)
	}
}

func TestRequestServiceBulkDelete(t *testing.T) {
	ownerID, workspaceID, flowID := uuid.New(), uuid.New(), uuid.New()
	requests := newRequestRepoMemory()
	flows := &flowRepoMemory{flow: &domain.Flow{ID: flowID, WorkspaceID: workspaceID}}
	workspaces := &workspaceRepoMemory{workspace: &domain.Workspace{ID: workspaceID, OwnerID: ownerID}}
	service := NewRequestService(requests, flows, workspaces, &collectionRepoMemory{}, &environmentRepoMemory{}, &historyRepoMemory{}, true)

	req1, _ := service.CreateRequest(context.Background(), ownerID, RequestInput{FlowID: flowID, Method: "GET", URL: "http://example.com/1"})
	req2, _ := service.CreateRequest(context.Background(), ownerID, RequestInput{FlowID: flowID, Method: "GET", URL: "http://example.com/2"})

	err := service.BulkDeleteRequests(context.Background(), ownerID, []uuid.UUID{req1.ID, req2.ID})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.GetRequest(context.Background(), ownerID, req1.ID); err == nil {
		t.Fatal("req1 should have been deleted")
	}
	if _, err := service.GetRequest(context.Background(), ownerID, req2.ID); err == nil {
		t.Fatal("req2 should have been deleted")
	}
}
