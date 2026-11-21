package application

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type MonitorInput struct {
	WorkspaceID uuid.UUID
	FlowID      uuid.UUID
	Name        string
	IntervalMin int
	WebhookURL  string
	IsActive    *bool
}

type MonitorPatch struct {
	Name        *string
	IntervalMin *int
	WebhookURL  *string
	IsActive    *bool
}

type MonitorService interface {
	CreateMonitor(ctx context.Context, ownerID uuid.UUID, input MonitorInput) (*domain.Monitor, error)
	GetByID(ctx context.Context, ownerID, id uuid.UUID) (*domain.Monitor, error)
	ListByWorkspace(ctx context.Context, ownerID, workspaceID uuid.UUID) ([]domain.Monitor, error)
	UpdateMonitor(ctx context.Context, ownerID, id uuid.UUID, patch MonitorPatch) (*domain.Monitor, error)
	DeleteMonitor(ctx context.Context, ownerID, id uuid.UUID) error
	RunMonitor(ctx context.Context, ownerID, id uuid.UUID) (*domain.MonitorRunLog, error)
	ListRunLogs(ctx context.Context, ownerID, id uuid.UUID, limit int) ([]domain.MonitorRunLog, error)
}

type monitorService struct {
	monitorRepo    domain.MonitorRepository
	workspaceRepo  domain.WorkspaceRepository
	flowRepo       domain.FlowRepository
	requestService RequestService
	httpClient     *http.Client
}

func NewMonitorService(
	monitorRepo domain.MonitorRepository,
	workspaceRepo domain.WorkspaceRepository,
	flowRepo domain.FlowRepository,
	requestService RequestService,
) MonitorService {
	return &monitorService{
		monitorRepo:    monitorRepo,
		workspaceRepo:  workspaceRepo,
		flowRepo:       flowRepo,
		requestService: requestService,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *monitorService) authorizeWorkspace(ctx context.Context, workspaceID, userID uuid.UUID) error {
	_, err := s.workspaceRepo.GetByID(ctx, workspaceID, userID)
	return err
}

func (s *monitorService) CreateMonitor(ctx context.Context, ownerID uuid.UUID, input MonitorInput) (*domain.Monitor, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, errors.New("monitor name is required")
	}
	if err := s.authorizeWorkspace(ctx, input.WorkspaceID, ownerID); err != nil {
		return nil, err
	}
	flow, err := s.flowRepo.GetByID(ctx, input.FlowID)
	if err != nil {
		return nil, errors.New("flow not found")
	}
	if flow.WorkspaceID != input.WorkspaceID {
		return nil, errors.New("flow does not belong to specified workspace")
	}

	interval := input.IntervalMin
	if interval <= 0 {
		interval = 60
	}
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	monitor := &domain.Monitor{
		WorkspaceID: input.WorkspaceID,
		FlowID:      input.FlowID,
		Name:        strings.TrimSpace(input.Name),
		IntervalMin: interval,
		WebhookURL:  strings.TrimSpace(input.WebhookURL),
		IsActive:    isActive,
	}

	if err := s.monitorRepo.Create(ctx, monitor); err != nil {
		return nil, err
	}

	return monitor, nil
}

func (s *monitorService) GetByID(ctx context.Context, ownerID, id uuid.UUID) (*domain.Monitor, error) {
	monitor, err := s.monitorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeWorkspace(ctx, monitor.WorkspaceID, ownerID); err != nil {
		return nil, err
	}
	return monitor, nil
}

func (s *monitorService) ListByWorkspace(ctx context.Context, ownerID, workspaceID uuid.UUID) ([]domain.Monitor, error) {
	if err := s.authorizeWorkspace(ctx, workspaceID, ownerID); err != nil {
		return nil, err
	}
	return s.monitorRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *monitorService) UpdateMonitor(ctx context.Context, ownerID, id uuid.UUID, patch MonitorPatch) (*domain.Monitor, error) {
	monitor, err := s.GetByID(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}

	if patch.Name != nil && strings.TrimSpace(*patch.Name) != "" {
		monitor.Name = strings.TrimSpace(*patch.Name)
	}
	if patch.IntervalMin != nil && *patch.IntervalMin > 0 {
		monitor.IntervalMin = *patch.IntervalMin
	}
	if patch.WebhookURL != nil {
		monitor.WebhookURL = strings.TrimSpace(*patch.WebhookURL)
	}
	if patch.IsActive != nil {
		monitor.IsActive = *patch.IsActive
	}

	if err := s.monitorRepo.Update(ctx, monitor); err != nil {
		return nil, err
	}

	return monitor, nil
}

func (s *monitorService) DeleteMonitor(ctx context.Context, ownerID, id uuid.UUID) error {
	if _, err := s.GetByID(ctx, ownerID, id); err != nil {
		return err
	}
	return s.monitorRepo.Delete(ctx, id)
}

func (s *monitorService) RunMonitor(ctx context.Context, ownerID, id uuid.UUID) (*domain.MonitorRunLog, error) {
	monitor, err := s.GetByID(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}

	requests, err := s.requestService.GetRequestsByFlow(ctx, ownerID, monitor.FlowID)
	if err != nil {
		return nil, err
	}

	started := time.Now()
	passedSteps := 0
	failedSteps := 0
	var firstErrMsg string

	accumulatedVars := make(map[string]string)

	for _, req := range requests {
		result, runErr := s.requestService.ExecuteRequestWithVariables(ctx, ownerID, req.ID, nil, accumulatedVars)
		if runErr != nil {
			failedSteps++
			if firstErrMsg == "" {
				firstErrMsg = runErr.Error()
			}
			continue
		}
		if result.Status >= 400 || result.TestsFailed > 0 {
			failedSteps++
			if firstErrMsg == "" {
				if result.TestsFailed > 0 {
					firstErrMsg = "assertions failed"
				} else {
					firstErrMsg = result.StatusText
				}
			}
		} else {
			passedSteps++
		}

		for k, v := range result.ExtractedVariables {
			accumulatedVars[k] = v
		}
	}

	status := "passed"
	if failedSteps > 0 || (len(requests) > 0 && passedSteps == 0) {
		status = "failed"
	}

	duration := time.Since(started).Milliseconds()
	runLog := &domain.MonitorRunLog{
		MonitorID:   monitor.ID,
		Status:      status,
		TotalSteps:  len(requests),
		PassedSteps: passedSteps,
		FailedSteps: failedSteps,
		DurationMS:  duration,
		Error:       firstErrMsg,
		CreatedAt:   time.Now(),
	}

	_ = s.monitorRepo.CreateRunLog(ctx, runLog)

	now := time.Now()
	monitor.LastRunAt = &now
	monitor.LastStatus = status
	_ = s.monitorRepo.Update(ctx, monitor)

	// Send Alert Webhook if configured
	if monitor.WebhookURL != "" {
		go s.dispatchWebhook(monitor.WebhookURL, monitor, runLog)
	}

	return runLog, nil
}

func (s *monitorService) ListRunLogs(ctx context.Context, ownerID, id uuid.UUID, limit int) ([]domain.MonitorRunLog, error) {
	if _, err := s.GetByID(ctx, ownerID, id); err != nil {
		return nil, err
	}
	return s.monitorRepo.ListRunLogs(ctx, id, limit)
}

func (s *monitorService) dispatchWebhook(webhookURL string, monitor *domain.Monitor, runLog *domain.MonitorRunLog) {
	payload := map[string]any{
		"event":        "monitor.run",
		"monitor_id":   monitor.ID,
		"monitor_name": monitor.Name,
		"flow_id":      monitor.FlowID,
		"status":       runLog.Status,
		"total_steps":  runLog.TotalSteps,
		"passed_steps": runLog.PassedSteps,
		"failed_steps": runLog.FailedSteps,
		"duration_ms":  runLog.DurationMS,
		"error":        runLog.Error,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(data))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "API-Request-App-Monitor/1.0")

	resp, err := s.httpClient.Do(req)
	if err == nil && resp != nil {
		_ = resp.Body.Close()
	}
}
