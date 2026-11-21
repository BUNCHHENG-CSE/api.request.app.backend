package application

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type cookieRepoMem struct {
	cookies map[string]*domain.WorkspaceCookie
}

func newCookieRepoMem() *cookieRepoMem {
	return &cookieRepoMem{cookies: make(map[string]*domain.WorkspaceCookie)}
}

func (r *cookieRepoMem) Save(_ context.Context, c *domain.WorkspaceCookie) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	key := c.WorkspaceID.String() + ":" + c.Domain + ":" + c.Path + ":" + c.Name
	copy := *c
	r.cookies[key] = &copy
	return nil
}

func (r *cookieRepoMem) ListByWorkspace(_ context.Context, wsID uuid.UUID) ([]domain.WorkspaceCookie, error) {
	var list []domain.WorkspaceCookie
	for _, c := range r.cookies {
		if c.WorkspaceID == wsID {
			list = append(list, *c)
		}
	}
	return list, nil
}

func (r *cookieRepoMem) GetMatching(_ context.Context, wsID uuid.UUID, dom, path string) ([]domain.WorkspaceCookie, error) {
	var list []domain.WorkspaceCookie
	for _, c := range r.cookies {
		if c.WorkspaceID == wsID && c.Domain == dom {
			list = append(list, *c)
		}
	}
	return list, nil
}

func (r *cookieRepoMem) Delete(_ context.Context, id uuid.UUID) error {
	for k, c := range r.cookies {
		if c.ID == id {
			delete(r.cookies, k)
		}
	}
	return nil
}

func (r *cookieRepoMem) ClearByWorkspace(_ context.Context, wsID uuid.UUID) error {
	for k, c := range r.cookies {
		if c.WorkspaceID == wsID {
			delete(r.cookies, k)
		}
	}
	return nil
}

func (r *cookieRepoMem) DeleteByDomain(_ context.Context, wsID uuid.UUID, dom string) error {
	return nil
}

func TestCookieService(t *testing.T) {
	ownerID := uuid.New()
	wsID := uuid.New()

	wsRepo := &workspaceRepoMemory{workspace: &domain.Workspace{ID: wsID, OwnerID: ownerID, Name: "WS"}}
	cRepo := newCookieRepoMem()
	svc := NewCookieService(cRepo, wsRepo)

	t.Run("Add and List cookies", func(t *testing.T) {
		cookie, err := svc.AddCookie(context.Background(), ownerID, wsID, domain.WorkspaceCookie{
			Domain: "example.com",
			Name:   "session_id",
			Value:  "xyz123",
		})
		require.NoError(t, err)
		assert.Equal(t, "session_id", cookie.Name)

		list, err := svc.ListCookies(context.Background(), ownerID, wsID)
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("AttachCookiesToRequest", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "https://example.com/api", nil)
		err := svc.AttachCookiesToRequest(context.Background(), wsID, "https://example.com/api", req)
		require.NoError(t, err)
		c, err := req.Cookie("session_id")
		require.NoError(t, err)
		assert.Equal(t, "xyz123", c.Value)
	})

	t.Run("SaveCookiesFromResponse", func(t *testing.T) {
		respHeaders := http.Header{}
		respHeaders.Add("Set-Cookie", "token=jwt999; Domain=example.com; Path=/; Secure; HttpOnly")
		err := svc.SaveCookiesFromResponse(context.Background(), wsID, "https://example.com/login", respHeaders)
		require.NoError(t, err)

		list, err := svc.ListCookies(context.Background(), ownerID, wsID)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})
}

func TestPreRequestDirectives(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "https://example.com/api", nil)
	body := []byte(`{"amount": 100}`)
	scripts := `
		@hmac_sha256 X-Signature mysecretkey
		@timestamp_header X-Timestamp
		@uuid_header X-Nonce
		@base64_header X-Custom-Auth testuser
	`
	applyPreRequestDirectives(req, scripts, body)

	assert.NotEmpty(t, req.Header.Get("X-Signature"))
	assert.NotEmpty(t, req.Header.Get("X-Timestamp"))
	assert.NotEmpty(t, req.Header.Get("X-Nonce"))
	assert.Equal(t, "dGVzdHVzZXI=", req.Header.Get("X-Custom-Auth"))
}

func TestDocsService(t *testing.T) {
	ownerID := uuid.New()
	wsID := uuid.New()
	flowID := uuid.New()
	colID := uuid.New()

	wsRepo := &workspaceRepoMemory{workspace: &domain.Workspace{ID: wsID, OwnerID: ownerID, Name: "Payments API"}}
	colRepo := &collectionRepoMemory{collection: &domain.Collection{ID: colID, WorkspaceID: wsID, Name: "Invoices"}}
	flowRepo := &flowRepoMemory{flow: &domain.Flow{ID: flowID, WorkspaceID: wsID, Name: "Flow"}}
	reqRepo := newRequestRepoMemory()
	_ = reqRepo.Create(context.Background(), &domain.Request{
		ID:           uuid.New(),
		FlowID:       flowID,
		CollectionID: &colID,
		Name:         "Create Invoice",
		Method:       "POST",
		URL:          "https://api.example.com/invoices",
		Headers:      `[{"key":"Content-Type","value":"application/json"}]`,
		Body:         `{"total": 500}`,
	})

	svc := NewDocsService(wsRepo, colRepo, flowRepo, reqRepo)

	t.Run("Generate Markdown", func(t *testing.T) {
		md, ct, err := svc.GenerateCollectionDocs(context.Background(), ownerID, colID, "markdown")
		require.NoError(t, err)
		assert.Equal(t, "text/markdown; charset=utf-8", ct)
		assert.Contains(t, md, "# Invoices")
		assert.Contains(t, md, "`POST` `https://api.example.com/invoices`")
	})

	t.Run("Generate HTML", func(t *testing.T) {
		htmlDoc, ct, err := svc.GenerateWorkspaceDocs(context.Background(), ownerID, wsID, "html")
		require.NoError(t, err)
		assert.Equal(t, "text/html; charset=utf-8", ct)
		assert.Contains(t, htmlDoc, "<title>Payments API - API Documentation</title>")
		assert.Contains(t, htmlDoc, "Create Invoice")
	})
}

func TestGraphQLService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"data": map[string]any{
				"__schema": map[string]any{
					"queryType": map[string]any{"name": "Query"},
					"types": []any{
						map[string]any{"name": "User", "kind": "OBJECT"},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := NewGraphQLService()
	res, err := svc.Introspect(context.Background(), server.URL, nil)
	require.NoError(t, err)
	schema, ok := res["__schema"].(map[string]any)
	require.True(t, ok)
	assert.NotNil(t, schema["queryType"])
}

func TestBenchmarkService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	ownerID := uuid.New()
	wsID := uuid.New()
	flowID := uuid.New()

	wsRepo := &workspaceRepoMemory{workspace: &domain.Workspace{ID: wsID, OwnerID: ownerID, Name: "Bench WS"}}
	flowRepo := &flowRepoMemory{flow: &domain.Flow{ID: flowID, WorkspaceID: wsID, Name: "Flow"}}
	reqRepo := newRequestRepoMemory()
	_ = reqRepo.Create(context.Background(), &domain.Request{
		ID:     uuid.New(),
		FlowID: flowID,
		Name:   "Fast Health",
		Method: "GET",
		URL:    server.URL,
	})

	reqSvc := NewRequestService(reqRepo, flowRepo, wsRepo, &collectionRepoMemory{}, &environmentRepoMemory{}, &multiHistoryRepoMemory{items: make(map[uuid.UUID]*domain.RequestHistory)}, true)
	benchSvc := NewBenchmarkService(flowRepo, wsRepo, reqSvc)

	result, err := benchSvc.RunBenchmark(context.Background(), ownerID, BenchmarkInput{
		FlowID:      flowID,
		Iterations:  6,
		Concurrency: 2,
	})
	require.NoError(t, err)
	assert.Equal(t, 6, result.TotalRuns)
	assert.Equal(t, 6, result.PassedRuns)
	assert.Equal(t, 0, result.FailedRuns)
	assert.True(t, result.P50LatencyMS > 0)
}
