package application

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

const maxExecutionResponseBytes = 2 * 1024 * 1024

type RequestInput struct {
	FlowID       uuid.UUID
	CollectionID *uuid.UUID
	Name         string
	Position     int
	Method       string
	URL          string
	Headers      string
	QueryParams  string
	Body         string
	BodyType     string
	Auth         string
	Scripts      string
	Settings     string
}

type RequestPatch struct {
	FlowID      *uuid.UUID
	Name        *string
	Position    *int
	Method      *string
	URL         *string
	Headers     *string
	QueryParams *string
	Body        *string
	BodyType    *string
	Auth        *string
	Scripts     *string
	Settings    *string
}

type ExecutionResult struct {
	HistoryID  uuid.UUID         `json:"history_id"`
	Status     int               `json:"status"`
	StatusText string            `json:"statusText"`
	Time       int64             `json:"time"`
	Size       string            `json:"size"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Truncated  bool              `json:"truncated"`
}

type RequestService interface {
	CreateRequest(ctx context.Context, ownerID uuid.UUID, input RequestInput) (*domain.Request, error)
	GetRequest(ctx context.Context, ownerID, id uuid.UUID) (*domain.Request, error)
	GetRequestsByFlow(ctx context.Context, ownerID, flowID uuid.UUID) ([]domain.Request, error)
	UpdateRequest(ctx context.Context, ownerID, id uuid.UUID, patch RequestPatch) (*domain.Request, error)
	DeleteRequest(ctx context.Context, ownerID, id uuid.UUID) error
	SetCollection(ctx context.Context, ownerID, requestID uuid.UUID, collectionID *uuid.UUID) (*domain.Request, error)
	DuplicateRequest(ctx context.Context, ownerID, id uuid.UUID) (*domain.Request, error)
	BulkDeleteRequests(ctx context.Context, ownerID uuid.UUID, ids []uuid.UUID) error
	ExecuteRequest(ctx context.Context, ownerID, id uuid.UUID, environmentID *uuid.UUID) (*ExecutionResult, error)
	ListHistory(ctx context.Context, ownerID uuid.UUID, workspaceID *uuid.UUID, page, limit int) ([]domain.RequestHistory, int64, error)
	GetHistory(ctx context.Context, ownerID, id uuid.UUID) (*domain.RequestHistory, error)
	ClearHistory(ctx context.Context, ownerID uuid.UUID) error
}

type requestService struct {
	repo            domain.RequestRepository
	flowRepo        domain.FlowRepository
	workspaceRepo   domain.WorkspaceRepository
	collectionRepo  domain.CollectionRepository
	environmentRepo domain.EnvironmentRepository
	historyRepo     domain.RequestHistoryRepository
	httpClient      *http.Client
}

func NewRequestService(repo domain.RequestRepository, flowRepo domain.FlowRepository, workspaceRepo domain.WorkspaceRepository, collectionRepo domain.CollectionRepository, environmentRepo domain.EnvironmentRepository, historyRepo domain.RequestHistoryRepository, allowPrivateRequests bool) RequestService {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = guardedDialContext(allowPrivateRequests)
	return &requestService{
		repo: repo, flowRepo: flowRepo, workspaceRepo: workspaceRepo, collectionRepo: collectionRepo,
		environmentRepo: environmentRepo, historyRepo: historyRepo,
		httpClient: &http.Client{Timeout: 30 * time.Second, Transport: transport},
	}
}

func guardedDialContext(allowPrivate bool) func(context.Context, string, string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
		if err != nil {
			return nil, err
		}
		for _, ip := range ips {
			if !allowPrivate && (ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast()) {
				continue
			}
			return dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		}
		return nil, errors.New("target resolves only to blocked private or local addresses")
	}
}

func validateRequest(method, rawURL string) (string, error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" || strings.ContainsAny(method, " \t\r\n") {
		return "", errors.New("valid method is required")
	}
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", errors.New("URL must be an absolute HTTP or HTTPS URL")
	}
	return method, nil
}

func (s *requestService) authorizeFlow(ctx context.Context, ownerID, flowID uuid.UUID) (*domain.Flow, error) {
	flow, err := s.flowRepo.GetByID(ctx, flowID)
	if err != nil {
		return nil, err
	}
	if _, err := s.workspaceRepo.GetByID(ctx, flow.WorkspaceID, ownerID); err != nil {
		return nil, err
	}
	return flow, nil
}

func (s *requestService) CreateRequest(ctx context.Context, ownerID uuid.UUID, input RequestInput) (*domain.Request, error) {
	method, err := validateRequest(input.Method, input.URL)
	if err != nil {
		return nil, err
	}
	if _, err := s.authorizeFlow(ctx, ownerID, input.FlowID); err != nil {
		return nil, err
	}
	req := &domain.Request{
		FlowID: input.FlowID, CollectionID: input.CollectionID, Name: input.Name, Position: input.Position, Method: method, URL: input.URL,
		Headers: input.Headers, QueryParams: input.QueryParams, Body: input.Body, BodyType: input.BodyType,
		Auth: input.Auth, Scripts: input.Scripts, Settings: input.Settings,
	}
	if input.CollectionID != nil {
		if err := s.validateCollection(ctx, ownerID, req, *input.CollectionID); err != nil {
			return nil, err
		}
	}
	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *requestService) GetRequest(ctx context.Context, ownerID, id uuid.UUID) (*domain.Request, error) {
	req, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if _, err := s.authorizeFlow(ctx, ownerID, req.FlowID); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *requestService) GetRequestsByFlow(ctx context.Context, ownerID, flowID uuid.UUID) ([]domain.Request, error) {
	if _, err := s.authorizeFlow(ctx, ownerID, flowID); err != nil {
		return nil, err
	}
	return s.repo.ListByFlow(ctx, flowID)
}

func (s *requestService) UpdateRequest(ctx context.Context, ownerID, id uuid.UUID, patch RequestPatch) (*domain.Request, error) {
	req, err := s.GetRequest(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	if patch.FlowID != nil {
		if _, err := s.authorizeFlow(ctx, ownerID, *patch.FlowID); err != nil {
			return nil, err
		}
		req.FlowID = *patch.FlowID
	}
	if patch.Name != nil {
		req.Name = *patch.Name
	}
	if patch.Position != nil {
		if *patch.Position < 0 {
			return nil, errors.New("position cannot be negative")
		}
		req.Position = *patch.Position
	}
	if patch.Method != nil {
		req.Method = *patch.Method
	}
	if patch.URL != nil {
		req.URL = *patch.URL
	}
	if normalized, err := validateRequest(req.Method, req.URL); err != nil {
		return nil, err
	} else {
		req.Method = normalized
	}
	if patch.Headers != nil {
		req.Headers = *patch.Headers
	}
	if patch.QueryParams != nil {
		req.QueryParams = *patch.QueryParams
	}
	if patch.Body != nil {
		req.Body = *patch.Body
	}
	if patch.BodyType != nil {
		req.BodyType = *patch.BodyType
	}
	if patch.Auth != nil {
		req.Auth = *patch.Auth
	}
	if patch.Scripts != nil {
		req.Scripts = *patch.Scripts
	}
	if patch.Settings != nil {
		req.Settings = *patch.Settings
	}
	if req.CollectionID != nil {
		if err := s.validateCollection(ctx, ownerID, req, *req.CollectionID); err != nil {
			return nil, err
		}
	}
	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *requestService) DeleteRequest(ctx context.Context, ownerID, id uuid.UUID) error {
	if _, err := s.GetRequest(ctx, ownerID, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *requestService) validateCollection(ctx context.Context, ownerID uuid.UUID, req *domain.Request, collectionID uuid.UUID) error {
	collection, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return err
	}
	flow, err := s.authorizeFlow(ctx, ownerID, req.FlowID)
	if err != nil {
		return err
	}
	if collection.WorkspaceID != flow.WorkspaceID {
		return errors.New("request and collection must belong to the same workspace")
	}
	if _, err := s.workspaceRepo.GetByID(ctx, collection.WorkspaceID, ownerID); err != nil {
		return err
	}
	return nil
}

func (s *requestService) SetCollection(ctx context.Context, ownerID, requestID uuid.UUID, collectionID *uuid.UUID) (*domain.Request, error) {
	req, err := s.GetRequest(ctx, ownerID, requestID)
	if err != nil {
		return nil, err
	}
	if collectionID != nil {
		if err := s.validateCollection(ctx, ownerID, req, *collectionID); err != nil {
			return nil, err
		}
	}
	req.CollectionID = collectionID
	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *requestService) DuplicateRequest(ctx context.Context, ownerID, id uuid.UUID) (*domain.Request, error) {
	original, err := s.GetRequest(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	duplicate := *original
	duplicate.ID = uuid.Nil
	duplicate.Name = strings.TrimSpace(original.Name + " Copy")
	duplicate.CreatedAt = time.Time{}
	duplicate.UpdatedAt = time.Time{}
	if err := s.repo.Create(ctx, &duplicate); err != nil {
		return nil, err
	}
	return &duplicate, nil
}

func (s *requestService) BulkDeleteRequests(ctx context.Context, ownerID uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return errors.New("at least one request ID is required")
	}
	for _, id := range ids {
		if err := s.DeleteRequest(ctx, ownerID, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *requestService) ListHistory(ctx context.Context, ownerID uuid.UUID, workspaceID *uuid.UUID, page, limit int) ([]domain.RequestHistory, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if workspaceID != nil {
		if _, err := s.workspaceRepo.GetByID(ctx, *workspaceID, ownerID); err != nil {
			return nil, 0, err
		}
	}
	return s.historyRepo.List(ctx, ownerID, workspaceID, limit, (page-1)*limit)
}

func (s *requestService) GetHistory(ctx context.Context, ownerID, id uuid.UUID) (*domain.RequestHistory, error) {
	return s.historyRepo.GetByID(ctx, id, ownerID)
}

func (s *requestService) ClearHistory(ctx context.Context, ownerID uuid.UUID) error {
	return s.historyRepo.DeleteByUser(ctx, ownerID)
}

func substituteVariables(value string, variables map[string]string) string {
	for key, replacement := range variables {
		value = strings.ReplaceAll(value, "{{"+key+"}}", replacement)
	}
	return value
}

func (s *requestService) executionVariables(ctx context.Context, ownerID uuid.UUID, flow *domain.Flow, environmentID *uuid.UUID) (map[string]string, error) {
	if environmentID == nil {
		return map[string]string{}, nil
	}
	env, err := s.environmentRepo.GetByID(ctx, *environmentID)
	if err != nil {
		return nil, err
	}
	if env.WorkspaceID != flow.WorkspaceID {
		return nil, errors.New("environment and request must belong to the same workspace")
	}
	if _, err := s.workspaceRepo.GetByID(ctx, env.WorkspaceID, ownerID); err != nil {
		return nil, err
	}
	variables := map[string]string{}
	if env.Variables != "" {
		if err := json.Unmarshal([]byte(env.Variables), &variables); err != nil {
			return nil, errors.New("environment variables must be a JSON object of string values")
		}
	}
	return variables, nil
}

func (s *requestService) ExecuteRequest(ctx context.Context, ownerID, id uuid.UUID, environmentID *uuid.UUID) (*ExecutionResult, error) {
	stored, err := s.GetRequest(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	flow, err := s.authorizeFlow(ctx, ownerID, stored.FlowID)
	if err != nil {
		return nil, err
	}
	variables, err := s.executionVariables(ctx, ownerID, flow, environmentID)
	if err != nil {
		return nil, err
	}
	executable := *stored
	executable.URL = substituteVariables(executable.URL, variables)
	executable.Headers = substituteVariables(executable.Headers, variables)
	executable.QueryParams = substituteVariables(executable.QueryParams, variables)
	executable.Body = substituteVariables(executable.Body, variables)
	result, executionErr := s.executeStored(ctx, &executable)
	history := &domain.RequestHistory{
		UserID: ownerID, WorkspaceID: flow.WorkspaceID, RequestID: stored.ID,
		Method: executable.Method, URL: executable.URL,
	}
	if executionErr != nil {
		history.Error = executionErr.Error()
		_ = s.historyRepo.Create(ctx, history)
		return nil, executionErr
	}
	history.Status = result.Status
	history.DurationMS = result.Time
	history.SizeBytes = len(result.Body)
	encodedHeaders, _ := json.Marshal(result.Headers)
	history.ResponseHeaders = string(encodedHeaders)
	history.ResponseBody = result.Body
	if err := s.historyRepo.Create(ctx, history); err != nil {
		return nil, err
	}
	result.HistoryID = history.ID
	return result, nil
}

func (s *requestService) executeStored(ctx context.Context, stored *domain.Request) (*ExecutionResult, error) {
	method, err := validateRequest(stored.Method, stored.URL)
	if err != nil {
		return nil, err
	}
	stored.Method = method
	var body io.Reader
	if stored.Method != http.MethodGet && stored.Method != http.MethodHead {
		body = bytes.NewBufferString(stored.Body)
	}
	targetURL, err := url.Parse(stored.URL)
	if err != nil {
		return nil, err
	}
	if stored.QueryParams != "" {
		var params map[string]string
		if err := json.Unmarshal([]byte(stored.QueryParams), &params); err != nil {
			return nil, errors.New("stored query_params must be a JSON object of string values")
		}
		query := targetURL.Query()
		for key, value := range params {
			query.Set(key, value)
		}
		targetURL.RawQuery = query.Encode()
	}
	outbound, err := http.NewRequestWithContext(ctx, stored.Method, targetURL.String(), body)
	if err != nil {
		return nil, err
	}
	if stored.Headers != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(stored.Headers), &headers); err != nil {
			return nil, errors.New("stored headers must be a JSON object of string values")
		}
		for key, value := range headers {
			outbound.Header.Set(key, value)
		}
	}
	started := time.Now()
	res, err := s.httpClient.Do(outbound)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, maxExecutionResponseBytes+1))
	if err != nil {
		return nil, err
	}
	truncated := len(data) > maxExecutionResponseBytes
	if truncated {
		data = data[:maxExecutionResponseBytes]
	}
	headers := make(map[string]string, len(res.Header))
	for key, values := range res.Header {
		headers[key] = strings.Join(values, ", ")
	}
	bodyText := string(data)
	if truncated {
		bodyText += "\n\n[Response truncated at 2 MB]"
	}
	return &ExecutionResult{
		Status: res.StatusCode, StatusText: res.Status, Time: time.Since(started).Milliseconds(),
		Size: fmt.Sprintf("%dB", len(data)), Headers: headers, Body: bodyText, Truncated: truncated,
	}, nil
}
