package application

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubstituteVariables_Dynamic(t *testing.T) {
	t.Run("substitutes dynamic variables", func(t *testing.T) {
		input := "GUID: {{$guid}}, TS: {{$timestamp}}, ISO: {{$isoTimestamp}}, INT: {{$randomInt}}, EMAIL: {{$randomEmail}}, NAME: {{$randomName}}"
		res := substituteVariables(input, nil)

		assert.NotContains(t, res, "{{$guid}}")
		assert.NotContains(t, res, "{{$timestamp}}")
		assert.NotContains(t, res, "{{$isoTimestamp}}")
		assert.NotContains(t, res, "{{$randomInt}}")
		assert.NotContains(t, res, "{{$randomEmail}}")
		assert.NotContains(t, res, "{{$randomName}}")

		assert.Contains(t, res, "GUID: ")
		assert.Contains(t, res, "@example.com")
		assert.Contains(t, res, "User_")
	})

	t.Run("substitutes runtime variables with precedence over dynamic", func(t *testing.T) {
		vars := map[string]string{
			"custom": "hello",
			"$guid":  "fixed-guid",
		}
		res := substituteVariables("{{custom}} - {{$guid}}", vars)
		assert.Equal(t, "hello - fixed-guid", res)
	})
}

func TestExtractVariables(t *testing.T) {
	headers := map[string]string{
		"Authorization": "Bearer sample-token-123",
		"X-Request-ID":   "req-xyz",
	}
	body := `{
		"code": 200,
		"data": {
			"token": "jwt-token-abc",
			"user": {
				"id": 42,
				"name": "Alice"
			},
			"items": [
				{"id": "item-1"},
				{"id": "item-2"}
			]
		}
	}`

	t.Run("extract from JSON config array", func(t *testing.T) {
		rules := `[
			{"name": "authToken", "source": "body", "path": "data.token"},
			{"name": "uid", "source": "body", "path": "data.user.id"},
			{"name": "authHeader", "source": "header", "path": "Authorization"},
			{"name": "firstItem", "source": "body", "path": "data.items[0].id"}
		]`
		extracted := ExtractVariables(rules, "", body, headers)
		assert.Equal(t, "jwt-token-abc", extracted["authToken"])
		assert.Equal(t, "42", extracted["uid"])
		assert.Equal(t, "Bearer sample-token-123", extracted["authHeader"])
		assert.Equal(t, "item-1", extracted["firstItem"])
	})

	t.Run("extract from Bruno and Postman script syntax", func(t *testing.T) {
		scripts := `
			// Bruno syntax
			bru.setVar("tokenFromBru", res.body.data.token)
			// Postman syntax
			pm.environment.set("userIdFromPm", pm.response.json().data.user.id)
			// @extract directive
			@extract authDirect = header.Authorization
		`
		extracted := ExtractVariables(scripts, "", body, headers)
		assert.Equal(t, "jwt-token-abc", extracted["tokenFromBru"])
		assert.Equal(t, "42", extracted["userIdFromPm"])
		assert.Equal(t, "Bearer sample-token-123", extracted["authDirect"])
	})
}

type multiHistoryRepoMemory struct {
	mu    sync.RWMutex
	items map[uuid.UUID]*domain.RequestHistory
}

func (r *multiHistoryRepoMemory) Create(_ context.Context, history *domain.RequestHistory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if history.ID == uuid.Nil {
		history.ID = uuid.New()
	}
	copy := *history
	r.items[history.ID] = &copy
	return nil
}

func (r *multiHistoryRepoMemory) GetByID(_ context.Context, id, userID uuid.UUID) (*domain.RequestHistory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.items[id]
	if !ok || h.UserID != userID {
		return nil, errors.New("history not found")
	}
	return h, nil
}

func (r *multiHistoryRepoMemory) List(context.Context, uuid.UUID, *uuid.UUID, int, int) ([]domain.RequestHistory, int64, error) {
	return nil, 0, nil
}

func (r *multiHistoryRepoMemory) DeleteByUser(context.Context, uuid.UUID) error {
	return nil
}

func TestCompareHistory(t *testing.T) {
	ownerID := uuid.New()
	wsID := uuid.New()
	reqID := uuid.New()

	hRepo := &multiHistoryRepoMemory{items: make(map[uuid.UUID]*domain.RequestHistory)}
	h1 := &domain.RequestHistory{
		UserID:          ownerID,
		WorkspaceID:     wsID,
		RequestID:       reqID,
		Method:          "GET",
		URL:             "https://example.com/api",
		Status:          200,
		DurationMS:      120,
		SizeBytes:       50,
		ResponseHeaders: `{"Content-Type": "application/json", "X-Version": "1"}`,
		ResponseBody:    `{"status": "ok"}`,
	}
	h2 := &domain.RequestHistory{
		UserID:          ownerID,
		WorkspaceID:     wsID,
		RequestID:       reqID,
		Method:          "GET",
		URL:             "https://example.com/api",
		Status:          500,
		DurationMS:      250,
		SizeBytes:       80,
		ResponseHeaders: `{"Content-Type": "application/json", "X-Version": "2"}`,
		ResponseBody:    `{"error": "internal error"}`,
	}
	require.NoError(t, hRepo.Create(context.Background(), h1))
	require.NoError(t, hRepo.Create(context.Background(), h2))

	svc := &requestService{
		historyRepo: hRepo,
	}

	diff, err := svc.CompareHistory(context.Background(), ownerID, h1.ID, h2.ID)
	require.NoError(t, err)
	assert.False(t, diff.StatusMatch)
	assert.False(t, diff.BodyMatch)
	assert.Equal(t, int64(130), diff.TimeDiffMS)
	assert.Equal(t, 30, diff.SizeDiff)
	assert.Contains(t, diff.HeaderDiff, "X-Version")
	assert.Equal(t, [2]string{"1", "2"}, diff.HeaderDiff["X-Version"])
}

func TestExecuteRequestWithVariables(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer my-runtime-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"token":"extracted-next-token"}}`))
	}))
	defer server.Close()

	ownerID := uuid.New()
	wsID := uuid.New()
	flowID := uuid.New()

	wsRepo := &workspaceRepoMemory{workspace: &domain.Workspace{ID: wsID, OwnerID: ownerID, Name: "WS"}}
	flowRepo := &flowRepoMemory{flow: &domain.Flow{ID: flowID, WorkspaceID: wsID, Name: "Flow"}}
	reqRepo := newRequestRepoMemory()
	reqEntity := &domain.Request{
		ID:       uuid.New(),
		FlowID:   flowID,
		Name:     "Step 1",
		Method:   "GET",
		URL:      server.URL,
		Headers:  `[{"key": "Authorization", "value": "Bearer {{customToken}}", "enabled": true}]`,
		Scripts:  `bru.setVar("nextToken", res.body.data.token)`,
	}
	_ = reqRepo.Create(context.Background(), reqEntity)

	hRepo := &multiHistoryRepoMemory{items: make(map[uuid.UUID]*domain.RequestHistory)}
	envRepo := &environmentRepoMemory{}
	colRepo := &collectionRepoMemory{}

	svc := NewRequestService(reqRepo, flowRepo, wsRepo, colRepo, envRepo, hRepo, true)

	res, err := svc.ExecuteRequestWithVariables(context.Background(), ownerID, reqEntity.ID, nil, map[string]string{
		"customToken": "my-runtime-token",
	})
	require.NoError(t, err)
	assert.Equal(t, 200, res.Status)
	assert.Equal(t, "extracted-next-token", res.ExtractedVariables["nextToken"])
}
