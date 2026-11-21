package application

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type monitorRepoMemory struct {
	monitors map[uuid.UUID]*domain.Monitor
	runLogs  map[uuid.UUID][]domain.MonitorRunLog
}

func newMonitorRepoMemory() *monitorRepoMemory {
	return &monitorRepoMemory{
		monitors: make(map[uuid.UUID]*domain.Monitor),
		runLogs:  make(map[uuid.UUID][]domain.MonitorRunLog),
	}
}

func (r *monitorRepoMemory) Create(_ context.Context, m *domain.Monitor) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	copy := *m
	r.monitors[m.ID] = &copy
	return nil
}

func (r *monitorRepoMemory) GetByID(_ context.Context, id uuid.UUID) (*domain.Monitor, error) {
	m, ok := r.monitors[id]
	if !ok {
		return nil, errors.New("monitor not found")
	}
	return m, nil
}

func (r *monitorRepoMemory) ListByWorkspace(_ context.Context, wsID uuid.UUID) ([]domain.Monitor, error) {
	var list []domain.Monitor
	for _, m := range r.monitors {
		if m.WorkspaceID == wsID {
			list = append(list, *m)
		}
	}
	return list, nil
}

func (r *monitorRepoMemory) Update(_ context.Context, m *domain.Monitor) error {
	copy := *m
	r.monitors[m.ID] = &copy
	return nil
}

func (r *monitorRepoMemory) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.monitors, id)
	delete(r.runLogs, id)
	return nil
}

func (r *monitorRepoMemory) CreateRunLog(_ context.Context, log *domain.MonitorRunLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	r.runLogs[log.MonitorID] = append([]domain.MonitorRunLog{*log}, r.runLogs[log.MonitorID]...)
	return nil
}

func (r *monitorRepoMemory) ListRunLogs(_ context.Context, mID uuid.UUID, limit int) ([]domain.MonitorRunLog, error) {
	logs := r.runLogs[mID]
	if len(logs) > limit {
		return logs[:limit], nil
	}
	return logs, nil
}

func TestMonitorService_RunAndWebhook(t *testing.T) {
	var webhookPayloadMu sync.Mutex
	var webhookPayload string
	webhookReceived := make(chan struct{}, 1)

	webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		webhookPayloadMu.Lock()
		webhookPayload = string(body)
		webhookPayloadMu.Unlock()
		select {
		case webhookReceived <- struct{}{}:
		default:
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer webhookServer.Close()

	apiTargetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer apiTargetServer.Close()

	ownerID := uuid.New()
	wsID := uuid.New()
	flowID := uuid.New()

	wsRepo := &workspaceRepoMemory{workspace: &domain.Workspace{ID: wsID, OwnerID: ownerID, Name: "Test WS"}}
	flowRepo := &flowRepoMemory{flow: &domain.Flow{ID: flowID, WorkspaceID: wsID, Name: "Flow 1"}}
	reqRepo := newRequestRepoMemory()
	reqEntity := &domain.Request{
		ID:     uuid.New(),
		FlowID: flowID,
		Name:   "Check Health",
		Method: "GET",
		URL:    apiTargetServer.URL,
	}
	_ = reqRepo.Create(context.Background(), reqEntity)

	mRepo := newMonitorRepoMemory()
	reqSvc := NewRequestService(reqRepo, flowRepo, wsRepo, &collectionRepoMemory{}, &environmentRepoMemory{}, &multiHistoryRepoMemory{items: make(map[uuid.UUID]*domain.RequestHistory)}, true)

	svc := NewMonitorService(mRepo, wsRepo, flowRepo, reqSvc)

	created, err := svc.CreateMonitor(context.Background(), ownerID, MonitorInput{
		WorkspaceID: wsID,
		FlowID:      flowID,
		Name:        "API Health Monitor",
		IntervalMin: 15,
		WebhookURL:  webhookServer.URL,
	})
	require.NoError(t, err)
	assert.Equal(t, "API Health Monitor", created.Name)
	assert.Equal(t, 15, created.IntervalMin)

	runLog, err := svc.RunMonitor(context.Background(), ownerID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "passed", runLog.Status)
	assert.Equal(t, 1, runLog.TotalSteps)
	assert.Equal(t, 1, runLog.PassedSteps)
	assert.Equal(t, 0, runLog.FailedSteps)

	// Wait for webhook dispatch
	<-webhookReceived
	webhookPayloadMu.Lock()
	assert.Contains(t, webhookPayload, "monitor.run")
	assert.Contains(t, webhookPayload, "API Health Monitor")
	assert.Contains(t, webhookPayload, "passed")
	webhookPayloadMu.Unlock()
}
