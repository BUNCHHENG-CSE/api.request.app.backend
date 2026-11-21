package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dummyMemberService struct {
	members []domain.WorkspaceMember
}

func (s *dummyMemberService) AddMember(_ context.Context, _, wsID uuid.UUID, id string, role domain.WorkspaceRole) (*domain.WorkspaceMember, error) {
	m := domain.WorkspaceMember{
		WorkspaceID: wsID,
		UserID:      uuid.New(),
		Role:        role,
		Username:    id,
	}
	s.members = append(s.members, m)
	return &m, nil
}

func (s *dummyMemberService) ListMembers(context.Context, uuid.UUID, uuid.UUID) ([]domain.WorkspaceMember, error) {
	return s.members, nil
}

func (s *dummyMemberService) UpdateRole(_ context.Context, _, wsID, uID uuid.UUID, role domain.WorkspaceRole) (*domain.WorkspaceMember, error) {
	return &domain.WorkspaceMember{WorkspaceID: wsID, UserID: uID, Role: role}, nil
}

func (s *dummyMemberService) RemoveMember(_ context.Context, _, _, _ uuid.UUID) error {
	return nil
}

func TestWorkspaceMemberHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &dummyMemberService{}
	h := NewWorkspaceMemberHandler(svc)
	wsID := uuid.New()
	userID := uuid.New()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.POST("/workspaces/:id/members", h.AddMember)
	router.GET("/workspaces/:id/members", h.ListMembers)
	router.PATCH("/workspaces/:id/members/:userId", h.UpdateRole)
	router.DELETE("/workspaces/:id/members/:userId", h.RemoveMember)

	t.Run("AddMember", func(t *testing.T) {
		body := `{"identifier": "alice@example.com", "role": "editor"}`
		req := httptest.NewRequest(http.MethodPost, "/workspaces/"+wsID.String()+"/members", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "alice@example.com")
	})

	t.Run("ListMembers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/workspaces/"+wsID.String()+"/members", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "alice@example.com")
	})

	t.Run("UpdateRole", func(t *testing.T) {
		targetUID := uuid.New()
		body := `{"role": "admin"}`
		req := httptest.NewRequest(http.MethodPatch, "/workspaces/"+wsID.String()+"/members/"+targetUID.String(), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "admin")
	})

	t.Run("RemoveMember", func(t *testing.T) {
		targetUID := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/workspaces/"+wsID.String()+"/members/"+targetUID.String(), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

type dummyOpenAPIService struct{}

func (s *dummyOpenAPIService) ImportOpenAPI(context.Context, uuid.UUID, uuid.UUID, []byte) (*application.OpenAPIImportResult, error) {
	return &application.OpenAPIImportResult{
		Collection: domain.Collection{Name: "Imported Petstore"},
	}, nil
}

func (s *dummyOpenAPIService) ExportWorkspaceOpenAPI(context.Context, uuid.UUID, uuid.UUID) ([]byte, error) {
	return []byte(`{"openapi": "3.0.3", "info": {"title": "Workspace API"}}`), nil
}

func (s *dummyOpenAPIService) ExportCollectionOpenAPI(context.Context, uuid.UUID, uuid.UUID) ([]byte, error) {
	return []byte(`{"openapi": "3.0.3", "info": {"title": "Collection API"}}`), nil
}

func TestOpenAPIHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOpenAPIHandler(&dummyOpenAPIService{})
	userID := uuid.New()
	wsID := uuid.New()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.POST("/collections/import-openapi", h.ImportOpenAPI)
	router.GET("/workspaces/:id/export-openapi", h.ExportWorkspace)

	t.Run("ImportOpenAPI JSON payload", func(t *testing.T) {
		payload := map[string]any{
			"workspace_id": wsID,
			"spec":         "openapi: 3.0.0\ninfo:\n  title: Petstore",
		}
		b, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/collections/import-openapi", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "Imported Petstore")
	})

	t.Run("ExportWorkspace", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/workspaces/"+wsID.String()+"/export-openapi", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Workspace API")
	})
}

type dummyMonitorService struct {
	monitors []domain.Monitor
}

func (s *dummyMonitorService) CreateMonitor(_ context.Context, _ uuid.UUID, in application.MonitorInput) (*domain.Monitor, error) {
	m := &domain.Monitor{
		ID:          uuid.New(),
		WorkspaceID: in.WorkspaceID,
		FlowID:      in.FlowID,
		Name:        in.Name,
		IntervalMin: in.IntervalMin,
	}
	s.monitors = append(s.monitors, *m)
	return m, nil
}
func (s *dummyMonitorService) GetByID(_ context.Context, _, id uuid.UUID) (*domain.Monitor, error) {
	return &domain.Monitor{ID: id, Name: "Found"}, nil
}
func (s *dummyMonitorService) ListByWorkspace(_ context.Context, _, _ uuid.UUID) ([]domain.Monitor, error) {
	return s.monitors, nil
}
func (s *dummyMonitorService) UpdateMonitor(_ context.Context, _, id uuid.UUID, _ application.MonitorPatch) (*domain.Monitor, error) {
	return &domain.Monitor{ID: id, Name: "Updated"}, nil
}
func (s *dummyMonitorService) DeleteMonitor(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
func (s *dummyMonitorService) RunMonitor(_ context.Context, _, id uuid.UUID) (*domain.MonitorRunLog, error) {
	return &domain.MonitorRunLog{MonitorID: id, Status: "passed", TotalSteps: 2, PassedSteps: 2}, nil
}
func (s *dummyMonitorService) ListRunLogs(_ context.Context, _, id uuid.UUID, _ int) ([]domain.MonitorRunLog, error) {
	return []domain.MonitorRunLog{{MonitorID: id, Status: "passed"}}, nil
}

func TestMonitorHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewMonitorHandler(&dummyMonitorService{})
	userID := uuid.New()
	wsID := uuid.New()
	flowID := uuid.New()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.POST("/monitors", h.Create)
	router.GET("/monitors", h.List)
	router.POST("/monitors/:id/run", h.Run)
	router.GET("/monitors/:id/logs", h.ListLogs)

	t.Run("CreateMonitor", func(t *testing.T) {
		payload := CreateMonitorRequest{
			WorkspaceID: wsID,
			FlowID:      flowID,
			Name:        "Daily Check",
			IntervalMin: 60,
		}
		b, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/monitors", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "Daily Check")
	})

	t.Run("ListMonitors", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/monitors?workspace_id="+wsID.String(), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("RunMonitor", func(t *testing.T) {
		mID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/monitors/"+mID.String()+"/run", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "passed")
	})
}

type dummyRequestDiffService struct {
	application.RequestService
}

func (s *dummyRequestDiffService) CompareHistory(_ context.Context, _, id1, id2 uuid.UUID) (*application.HistoryDiff, error) {
	return &application.HistoryDiff{
		StatusMatch: true,
		BodyMatch:   false,
		TimeDiffMS:  45,
	}, nil
}

func TestRequestCompareHistoryHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewRequestHandler(&dummyRequestDiffService{})
	userID := uuid.New()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.GET("/history/diff", h.CompareHistory)

	id1 := uuid.New()
	id2 := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/history/diff?id1="+id1.String()+"&id2="+id2.String(), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "status_match")
}

type dummyCookieSvc struct{}

func (s *dummyCookieSvc) AddCookie(_ context.Context, _, wsID uuid.UUID, c domain.WorkspaceCookie) (*domain.WorkspaceCookie, error) {
	c.WorkspaceID = wsID
	return &c, nil
}
func (s *dummyCookieSvc) ListCookies(context.Context, uuid.UUID, uuid.UUID) ([]domain.WorkspaceCookie, error) {
	return []domain.WorkspaceCookie{{Name: "test_cookie", Value: "123"}}, nil
}
func (s *dummyCookieSvc) DeleteCookie(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	return nil
}
func (s *dummyCookieSvc) ClearCookies(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
func (s *dummyCookieSvc) AttachCookiesToRequest(context.Context, uuid.UUID, string, *http.Request) error {
	return nil
}
func (s *dummyCookieSvc) SaveCookiesFromResponse(context.Context, uuid.UUID, string, http.Header) error {
	return nil
}

func TestCookieHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCookieHandler(&dummyCookieSvc{})
	userID := uuid.New()
	wsID := uuid.New()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.GET("/workspaces/:id/cookies", h.ListCookies)
	router.POST("/workspaces/:id/cookies", h.AddCookie)
	router.DELETE("/workspaces/:id/cookies", h.ClearCookies)

	t.Run("ListCookies", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/workspaces/"+wsID.String()+"/cookies", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "test_cookie")
	})

	t.Run("AddCookie", func(t *testing.T) {
		body := `{"domain": "example.com", "name": "auth", "value": "secret"}`
		req := httptest.NewRequest(http.MethodPost, "/workspaces/"+wsID.String()+"/cookies", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "secret")
	})
}

type dummyBenchmarkSvc struct{}

func (s *dummyBenchmarkSvc) RunBenchmark(context.Context, uuid.UUID, application.BenchmarkInput) (*application.BenchmarkResult, error) {
	return &application.BenchmarkResult{TotalRuns: 10, PassedRuns: 10, RPS: 150.5}, nil
}

func TestBenchmarkHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewBenchmarkHandler(&dummyBenchmarkSvc{})
	userID := uuid.New()
	flowID := uuid.New()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.POST("/flows/:id/benchmark", h.RunBenchmark)

	req := httptest.NewRequest(http.MethodPost, "/flows/"+flowID.String()+"/benchmark", bytes.NewBufferString(`{"iterations": 10}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "150.5")
}

type dummyGraphQLSvc struct{}

func (s *dummyGraphQLSvc) Introspect(context.Context, string, map[string]string) (map[string]any, error) {
	return map[string]any{"__schema": map[string]any{"types": []any{}}}, nil
}

func TestGraphQLHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewGraphQLHandler(&dummyGraphQLSvc{})

	router := gin.New()
	router.POST("/graphql/introspect", h.Introspect)

	req := httptest.NewRequest(http.MethodPost, "/graphql/introspect", bytes.NewBufferString(`{"url": "https://example.com/graphql"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "__schema")
}

type dummyDocsSvc struct{}

func (s *dummyDocsSvc) GenerateCollectionDocs(context.Context, uuid.UUID, uuid.UUID, string) (string, string, error) {
	return "# Collection API Docs", "text/markdown; charset=utf-8", nil
}
func (s *dummyDocsSvc) GenerateWorkspaceDocs(context.Context, uuid.UUID, uuid.UUID, string) (string, string, error) {
	return "<html><body><h1>Workspace API</h1></body></html>", "text/html; charset=utf-8", nil
}

func TestDocsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewDocsHandler(&dummyDocsSvc{})
	userID := uuid.New()
	id := uuid.New()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.GET("/collections/:id/docs", h.GenerateCollectionDocs)
	router.GET("/workspaces/:id/docs", h.GenerateWorkspaceDocs)

	req1 := httptest.NewRequest(http.MethodGet, "/collections/"+id.String()+"/docs", nil)
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)
	assert.Contains(t, rec1.Body.String(), "# Collection API Docs")

	req2 := httptest.NewRequest(http.MethodGet, "/workspaces/"+id.String()+"/docs", nil)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Contains(t, rec2.Body.String(), "Workspace API")
}

type dummyActivitySvc struct{}

func (s *dummyActivitySvc) Log(context.Context, uuid.UUID, uuid.UUID, string, string, string) {}
func (s *dummyActivitySvc) List(context.Context, uuid.UUID, uuid.UUID, int) ([]domain.WorkspaceActivity, error) {
	return []domain.WorkspaceActivity{{Action: "flow.created", Username: "alice"}}, nil
}

func TestActivityHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewActivityHandler(&dummyActivitySvc{})
	userID := uuid.New()
	wsID := uuid.New()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.GET("/workspaces/:id/activity", h.List)

	req := httptest.NewRequest(http.MethodGet, "/workspaces/"+wsID.String()+"/activity", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "flow.created")
}

