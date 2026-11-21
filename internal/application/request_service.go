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
	"regexp"
	"strings"
	"time"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

const maxExecutionResponseBytes = 2 * 1024 * 1024

var variableRegex = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_\-\.]+)\s*\}\}`)

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
	HistoryID   uuid.UUID         `json:"history_id"`
	Status      int               `json:"status"`
	StatusText  string            `json:"statusText"`
	Time        int64             `json:"time"`
	Size        string            `json:"size"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	Truncated   bool              `json:"truncated"`
	Assertions  []AssertionResult `json:"assertions,omitempty"`
	TestsPassed int               `json:"tests_passed"`
	TestsFailed int               `json:"tests_failed"`
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
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 20
	transport.IdleConnTimeout = 90 * time.Second
	transport.ResponseHeaderTimeout = 30 * time.Second
	transport.ExpectContinueTimeout = 1 * time.Second
	transport.TLSHandshakeTimeout = 10 * time.Second

	return &requestService{
		repo:            repo,
		flowRepo:        flowRepo,
		workspaceRepo:   workspaceRepo,
		collectionRepo:  collectionRepo,
		environmentRepo: environmentRepo,
		historyRepo:     historyRepo,
		httpClient:      &http.Client{Timeout: 30 * time.Second, Transport: transport},
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
		var lastDialErr error
		hasValidIP := false
		for _, ip := range ips {
			if !allowPrivate && (ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast()) {
				continue
			}
			hasValidIP = true
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if err == nil {
				return conn, nil
			}
			lastDialErr = err
		}
		if !hasValidIP {
			return nil, errors.New("target resolves only to blocked private or local addresses")
		}
		if lastDialErr != nil {
			return nil, lastDialErr
		}
		return nil, errors.New("failed to connect to target host")
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
		FlowID:       input.FlowID,
		CollectionID: input.CollectionID,
		Name:         strings.TrimSpace(input.Name),
		Position:     input.Position,
		Method:       method,
		URL:          input.URL,
		Headers:      input.Headers,
		QueryParams:  input.QueryParams,
		Body:         input.Body,
		BodyType:     input.BodyType,
		Auth:         input.Auth,
		Scripts:      input.Scripts,
		Settings:     input.Settings,
	}
	if req.Position < 0 {
		return nil, errors.New("position cannot be negative")
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
		req.Name = strings.TrimSpace(*patch.Name)
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
	// Verify all requests belong to user first
	for _, id := range ids {
		if _, err := s.GetRequest(ctx, ownerID, id); err != nil {
			return err
		}
	}
	for _, id := range ids {
		if err := s.repo.Delete(ctx, id); err != nil {
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
	if len(variables) == 0 || value == "" {
		return value
	}
	return variableRegex.ReplaceAllStringFunc(value, func(match string) string {
		submatches := variableRegex.FindStringSubmatch(match)
		if len(submatches) > 1 {
			varName := submatches[1]
			if replacement, ok := variables[varName]; ok {
				return replacement
			}
		}
		return match
	})
}

// parseKeyValueData parses JSON into map[string]string supporting both:
// 1) Object format: {"key": "val", "num": 1}
// 2) Array format: [{"key": "k", "value": "v", "enabled": true}]
func parseKeyValueData(raw string, fieldLabel string) (map[string]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" || raw == "[]" {
		return map[string]string{}, nil
	}

	// Try object map format first
	var objMap map[string]any
	if err := json.Unmarshal([]byte(raw), &objMap); err == nil {
		result := make(map[string]string, len(objMap))
		for k, v := range objMap {
			if v != nil {
				result[k] = fmt.Sprintf("%v", v)
			}
		}
		return result, nil
	}

	// Try array of objects format
	var arr []map[string]any
	if err := json.Unmarshal([]byte(raw), &arr); err == nil {
		result := make(map[string]string, len(arr))
		for _, item := range arr {
			// Check if item is explicitly disabled
			if enabled, ok := item["enabled"].(bool); ok && !enabled {
				continue
			}
			key := ""
			if k, ok := item["key"].(string); ok {
				key = k
			} else if name, ok := item["name"].(string); ok {
				key = name
			}
			if key == "" {
				continue
			}
			val := ""
			if v, ok := item["value"]; ok && v != nil {
				val = fmt.Sprintf("%v", v)
			}
			result[key] = val
		}
		return result, nil
	}

	return nil, fmt.Errorf("stored %s must be a valid JSON object or array of key-value pairs", fieldLabel)
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
		parsed, err := parseKeyValueData(env.Variables, "environment variables")
		if err != nil {
			return nil, err
		}
		variables = parsed
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
	executable.Auth = substituteVariables(executable.Auth, variables)

	result, executionErr := s.executeStored(ctx, &executable)
	history := &domain.RequestHistory{
		UserID:      ownerID,
		WorkspaceID: flow.WorkspaceID,
		RequestID:   stored.ID,
		Method:      executable.Method,
		URL:         executable.URL,
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

	// Evaluate declarative assertions if defined in scripts or settings
	assertions := evaluateAssertions(stored.Scripts, result)
	if len(assertions) == 0 {
		assertions = evaluateAssertions(stored.Settings, result)
	}
	if len(assertions) > 0 {
		result.Assertions = assertions
		for _, a := range assertions {
			if a.Passed {
				result.TestsPassed++
			} else {
				result.TestsFailed++
			}
		}
	}

	return result, nil
}

func applyAuth(outbound *http.Request, targetURL *url.URL, authRaw string) {
	authRaw = strings.TrimSpace(authRaw)
	if authRaw == "" || authRaw == "{}" {
		return
	}
	var authMap map[string]any
	if err := json.Unmarshal([]byte(authRaw), &authMap); err != nil {
		return
	}

	authType := strings.ToLower(fmt.Sprintf("%v", authMap["type"]))

	// Bearer Token
	if authType == "bearer" || authMap["bearer"] != nil {
		token := ""
		if t, ok := authMap["token"].(string); ok && t != "" {
			token = t
		} else if b, ok := authMap["bearer"].(string); ok && b != "" {
			token = b
		} else if bMap, ok := authMap["bearer"].(map[string]any); ok {
			token = fmt.Sprintf("%v", bMap["token"])
		}
		if token != "" && outbound.Header.Get("Authorization") == "" {
			outbound.Header.Set("Authorization", "Bearer "+token)
		}
		return
	}

	// Basic Auth
	if authType == "basic" || authMap["basic"] != nil {
		username := fmt.Sprintf("%v", authMap["username"])
		password := fmt.Sprintf("%v", authMap["password"])
		if bMap, ok := authMap["basic"].(map[string]any); ok {
			username = fmt.Sprintf("%v", bMap["username"])
			password = fmt.Sprintf("%v", bMap["password"])
		}
		if username != "" && outbound.Header.Get("Authorization") == "" {
			outbound.SetBasicAuth(username, password)
		}
		return
	}

	// API Key
	if authType == "apikey" || authType == "api_key" {
		key := fmt.Sprintf("%v", authMap["key"])
		val := fmt.Sprintf("%v", authMap["value"])
		in := strings.ToLower(fmt.Sprintf("%v", authMap["in"]))
		if key != "" && val != "" {
			if in == "query" {
				q := targetURL.Query()
				q.Set(key, val)
				targetURL.RawQuery = q.Encode()
			} else if outbound.Header.Get(key) == "" {
				outbound.Header.Set(key, val)
			}
		}
	}
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

	// Query Parameters
	if stored.QueryParams != "" {
		params, err := parseKeyValueData(stored.QueryParams, "query_params")
		if err != nil {
			return nil, err
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

	// Headers
	if stored.Headers != "" {
		headers, err := parseKeyValueData(stored.Headers, "headers")
		if err != nil {
			return nil, err
		}
		for key, value := range headers {
			outbound.Header.Set(key, value)
		}
	}

	// Automatic Content-Type inference if not explicitly provided
	if outbound.Header.Get("Content-Type") == "" && stored.Body != "" {
		switch strings.ToLower(strings.TrimSpace(stored.BodyType)) {
		case "json":
			outbound.Header.Set("Content-Type", "application/json")
		case "xml":
			outbound.Header.Set("Content-Type", "application/xml")
		case "form", "form-urlencoded", "urlencoded":
			outbound.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		case "text":
			outbound.Header.Set("Content-Type", "text/plain")
		}
	}

	// Auth Configuration
	applyAuth(outbound, targetURL, stored.Auth)
	// Update request URL if Auth modified query params
	outbound.URL = targetURL

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
		Status:     res.StatusCode,
		StatusText: res.Status,
		Time:       time.Since(started).Milliseconds(),
		Size:       fmt.Sprintf("%dB", len(data)),
		Headers:    headers,
		Body:       bodyText,
		Truncated:  truncated,
	}, nil
}
