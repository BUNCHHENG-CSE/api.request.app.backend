package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"api.request.app.backend/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type mockWorkspaceRepo struct{}

func (m *mockWorkspaceRepo) Create(context.Context, *domain.Workspace) error { return nil }
func (m *mockWorkspaceRepo) GetByID(context.Context, uuid.UUID, uuid.UUID) (*domain.Workspace, error) {
	return &domain.Workspace{}, nil
}
func (m *mockWorkspaceRepo) List(context.Context, uuid.UUID) ([]domain.Workspace, error) {
	return nil, nil
}
func (m *mockWorkspaceRepo) Update(context.Context, *domain.Workspace) error      { return nil }
func (m *mockWorkspaceRepo) Delete(context.Context, uuid.UUID, uuid.UUID) error { return nil }

type mockFlowRepo struct {
	flow domain.Flow
}

func (m *mockFlowRepo) Create(context.Context, *domain.Flow) error { return nil }
func (m *mockFlowRepo) GetByID(context.Context, uuid.UUID) (*domain.Flow, error) {
	return &m.flow, nil
}
func (m *mockFlowRepo) GetByWebhookToken(context.Context, string) (*domain.Flow, error) {
	return &m.flow, nil
}
func (m *mockFlowRepo) ListByWorkspace(context.Context, uuid.UUID) ([]domain.Flow, error) {
	return []domain.Flow{m.flow}, nil
}
func (m *mockFlowRepo) Update(context.Context, *domain.Flow) error { return nil }
func (m *mockFlowRepo) Delete(context.Context, uuid.UUID) error    { return nil }

type mockRequestRepo struct {
	request domain.Request
}

func (m *mockRequestRepo) Create(context.Context, *domain.Request) error { return nil }
func (m *mockRequestRepo) GetByID(context.Context, uuid.UUID) (*domain.Request, error) {
	return &m.request, nil
}
func (m *mockRequestRepo) ListByFlow(context.Context, uuid.UUID) ([]domain.Request, error) {
	return []domain.Request{m.request}, nil
}
func (m *mockRequestRepo) Update(context.Context, *domain.Request) error { return nil }
func (m *mockRequestRepo) Delete(context.Context, uuid.UUID) error    { return nil }

func TestMockHandlerServeMock(t *testing.T) {
	gin.SetMode(gin.TestMode)

	workspaceID := uuid.New()
	flowID := uuid.New()
	requestID := uuid.New()

	flowRepo := &mockFlowRepo{
		flow: domain.Flow{ID: flowID, WorkspaceID: workspaceID, Name: "Test Flow"},
	}
	reqRepo := &mockRequestRepo{
		request: domain.Request{
			ID:       requestID,
			FlowID:   flowID,
			Name:     "Get User",
			Method:   "GET",
			URL:      "https://api.example.com/api/users",
			Settings: `{"mock_status":200,"mock_body":"{\"name\":\"Mock Alice\"}"}`,
		},
	}

	handler := NewMockHandler(&mockWorkspaceRepo{}, flowRepo, reqRepo)

	router := gin.New()
	router.Any("/api/v1/mock/:workspaceId/*path", handler.ServeMock)

	// 1. Matching mock route
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mock/"+workspaceID.String()+"/api/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != `{"name":"Mock Alice"}` {
		t.Fatalf("unexpected mock body: %s", w.Body.String())
	}
	if w.Header().Get("X-Powered-By") != "FlowAPI-Mock-Server" {
		t.Fatalf("missing mock header")
	}

	// 2. Non-matching route returns 404
	req404 := httptest.NewRequest(http.MethodPost, "/api/v1/mock/"+workspaceID.String()+"/nonexistent", nil)
	w404 := httptest.NewRecorder()
	router.ServeHTTP(w404, req404)

	if w404.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w404.Code)
	}
}
