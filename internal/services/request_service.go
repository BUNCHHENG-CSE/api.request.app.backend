package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"backend/internal/entities"
	"backend/internal/models"
	"backend/internal/repositories"
)

// RequestService handles request operations
type RequestService interface {
	CreateRequest(ctx context.Context, dto *models.CreateRequestDTO) (*entities.RequestTab, error)
	GetRequest(ctx context.Context, id string) (*entities.RequestTab, error)
	ListRequests(ctx context.Context, collectionID string) ([]*entities.RequestTab, error)
	UpdateRequest(ctx context.Context, id string, dto *models.UpdateRequestDTO) (*entities.RequestTab, error)
	DeleteRequest(ctx context.Context, id string) error
	ExecuteRequest(ctx context.Context, dto *models.ExecuteRequestDTO) (*models.ExecuteResponseDTO, error)
	DuplicateRequest(ctx context.Context, id string) (*entities.RequestTab, error)
	SearchRequests(ctx context.Context, workspaceID string, query string) ([]*entities.RequestTab, error)
	GetRequestHistory(ctx context.Context, requestID string, limit int) ([]*entities.RequestHistory, error)
}

type requestService struct {
	repo            repositories.RequestRepository
	environmentRepo repositories.EnvironmentRepository
	responseRepo    repositories.ResponseRepository
	historyRepo     repositories.HistoryRepository
	httpClient      *http.Client
}

func NewRequestService(
	repo repositories.RequestRepository,
	envRepo repositories.EnvironmentRepository,
	respRepo repositories.ResponseRepository,
	histRepo repositories.HistoryRepository,
) RequestService {
	return &requestService{
		repo:            repo,
		environmentRepo: envRepo,
		responseRepo:    respRepo,
		historyRepo:     histRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *requestService) CreateRequest(ctx context.Context, dto *models.CreateRequestDTO) (*entities.RequestTab, error) {
	req := &entities.RequestTab{
		Name:             dto.Name,
		Description:      dto.Description,
		Method:           dto.Method,
		URL:              dto.URL,
		Body:             dto.Body,
		BodyType:         dto.BodyType,
		GraphqlQuery:     dto.GraphqlQuery,
		GraphqlVariables: dto.GraphqlVariables,
		CollectionID:     dto.CollectionID,
		FolderID:         dto.FolderID,
	}

	// Convert DTOs to entities
	if dto.Auth != nil {
		req.Auth = &entities.AuthorizationData{
			Type:          dto.Auth.Type,
			BearerToken:   dto.Auth.BearerToken,
			BasicUsername: dto.Auth.BasicUsername,
			BasicPassword: dto.Auth.BasicPassword,
			APIKeyKey:     dto.Auth.APIKeyKey,
			APIKeyValue:   dto.Auth.APIKeyValue,
			APIKeyIn:      dto.Auth.APIKeyIn,
		}
	}

	if dto.Scripts != nil {
		req.Scripts = &entities.Scripts{
			PreRequest:   dto.Scripts.PreRequest,
			PostResponse: dto.Scripts.PostResponse,
		}
	}

	if dto.Settings != nil {
		req.Settings = &entities.RequestSettings{
			HTTPVersion:               dto.Settings.HTTPVersion,
			StrictSSL:                 dto.Settings.StrictSSL,
			FollowRedirects:           dto.Settings.FollowRedirects,
			FollowAuthorizationHeader: dto.Settings.FollowAuthorizationHeader,
			MaxRedirects:              dto.Settings.MaxRedirects,
		}
	}

	return s.repo.Create(ctx, req)
}

func (s *requestService) GetRequest(ctx context.Context, id string) (*entities.RequestTab, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *requestService) ListRequests(ctx context.Context, collectionID string) ([]*entities.RequestTab, error) {
	return s.repo.ListByCollection(ctx, collectionID)
}

func (s *requestService) UpdateRequest(ctx context.Context, id string, dto *models.UpdateRequestDTO) (*entities.RequestTab, error) {
	req, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if dto.Name != nil {
		req.Name = *dto.Name
	}
	if dto.Description != nil {
		req.Description = *dto.Description
	}
	if dto.Method != nil {
		req.Method = *dto.Method
	}
	if dto.URL != nil {
		req.URL = *dto.URL
	}
	if dto.Body != nil {
		req.Body = *dto.Body
	}
	if dto.BodyType != nil {
		req.BodyType = *dto.BodyType
	}
	if dto.GraphqlQuery != nil {
		req.GraphqlQuery = *dto.GraphqlQuery
	}
	if dto.GraphqlVariables != nil {
		req.GraphqlVariables = *dto.GraphqlVariables
	}

	if dto.Auth != nil {
		req.Auth = &entities.AuthorizationData{
			Type:          dto.Auth.Type,
			BearerToken:   dto.Auth.BearerToken,
			BasicUsername: dto.Auth.BasicUsername,
			BasicPassword: dto.Auth.BasicPassword,
			APIKeyKey:     dto.Auth.APIKeyKey,
			APIKeyValue:   dto.Auth.APIKeyValue,
			APIKeyIn:      dto.Auth.APIKeyIn,
		}
	}

	if dto.Settings != nil {
		req.Settings = &entities.RequestSettings{
			HTTPVersion:     dto.Settings.HTTPVersion,
			StrictSSL:       dto.Settings.StrictSSL,
			FollowRedirects: dto.Settings.FollowRedirects,
			MaxRedirects:    dto.Settings.MaxRedirects,
		}
	}

	return s.repo.Update(ctx, req)
}

func (s *requestService) DeleteRequest(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ExecuteRequest runs an HTTP request
func (s *requestService) ExecuteRequest(ctx context.Context, dto *models.ExecuteRequestDTO) (*models.ExecuteResponseDTO, error) {
	req, err := s.repo.GetByID(ctx, dto.RequestID)
	if err != nil {
		return nil, err
	}

	// Apply overrides if provided
	if dto.Override != nil {
		if dto.Override.URL != nil {
			req.URL = *dto.Override.URL
		}
		if dto.Override.Method != nil {
			req.Method = *dto.Override.Method
		}
		if dto.Override.Body != nil {
			req.Body = *dto.Override.Body
		}
	}

	// Load environment variables if specified
	var envVars map[string]string
	if dto.EnvironmentID != nil {
		env, err := s.environmentRepo.GetByID(ctx, *dto.EnvironmentID)
		if err == nil && len(env.Variables) > 0 {
			envVars = make(map[string]string)

			// 1. Unmarshal the []byte into a slice of maps
			var varsList []map[string]interface{}
			if err := json.Unmarshal(env.Variables, &varsList); err == nil {
				// 2. Iterate safely
				for _, kv := range varsList {
					if name, kOk := kv["name"].(string); kOk {
						if val, vOk := kv["value"].(string); vOk {
							envVars[name] = val
						}
					}
				}
			}
		}
	}

	// Replace environment variables in URL and body
	req.URL = s.replaceVariables(req.URL, envVars)
	req.Body = s.replaceVariables(req.Body, envVars)

	// Build HTTP request
	httpReq, err := s.buildHTTPRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// Execute request
	start := time.Now()
	httpResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request execution failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(httpResp.Body)

	duration := time.Since(start).Milliseconds()

	// Read response body
	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Convert response headers
	headers := make([]models.KeyValuePairDTO, 0)
	for k, v := range httpResp.Header {
		for _, val := range v {
			headers = append(headers, models.KeyValuePairDTO{
				Key:   k,
				Value: val,
			})
		}
	}

	response := &models.ExecuteResponseDTO{
		StatusCode: httpResp.StatusCode,
		StatusText: httpResp.Status,
		Headers:    headers,
		Body:       string(bodyBytes),
		Size:       len(bodyBytes),
		Time:       int(duration),
	}

	respEntity := &entities.Response{
		RequestID:  req.ID,
		StatusCode: httpResp.StatusCode,
		StatusText: httpResp.Status,
		Body:       string(bodyBytes),
		Size:       len(bodyBytes),
		Time:       int(duration),
	}

	// Check the error here
	if _, err := s.responseRepo.Create(ctx, respEntity); err != nil {
		return nil, fmt.Errorf("failed to save response to database: %w", err)
	}

	return response, nil
}

func (s *requestService) DuplicateRequest(ctx context.Context, id string) (*entities.RequestTab, error) {
	req, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	newReq := &entities.RequestTab{
		CollectionID:     req.CollectionID,
		FolderID:         req.FolderID,
		WorkspaceID:      req.WorkspaceID,
		Name:             req.Name + " (copy)",
		Description:      req.Description,
		Method:           req.Method,
		URL:              req.URL,
		Body:             req.Body,
		BodyType:         req.BodyType,
		GraphqlQuery:     req.GraphqlQuery,
		GraphqlVariables: req.GraphqlVariables,
		Auth:             req.Auth,
		Scripts:          req.Scripts,
		Settings:         req.Settings,
	}

	return s.repo.Create(ctx, newReq)
}

func (s *requestService) SearchRequests(ctx context.Context, workspaceID string, query string) ([]*entities.RequestTab, error) {
	return s.repo.Search(ctx, workspaceID, query)
}

func (s *requestService) GetRequestHistory(ctx context.Context, requestID string, limit int) ([]*entities.RequestHistory, error) {
	return s.historyRepo.ListByRequest(ctx, requestID, limit)
}

// buildHTTPRequest converts a RequestTab to an HTTP request
func (s *requestService) buildHTTPRequest(req *entities.RequestTab) (*http.Request, error) {
	var body io.Reader
	if req.Body != "" {
		body = bytes.NewBufferString(req.Body)
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, body)
	if err != nil {
		return nil, err
	}

	// Add headers
	if len(req.Headers) > 0 {
		var headersList []map[string]interface{}
		if err := json.Unmarshal(req.Headers, &headersList); err == nil {
			for _, kv := range headersList {
				enabled, ok := kv["enabled"].(bool)
				if !ok || enabled {
					key, kOk := kv["key"].(string)
					val, vOk := kv["value"].(string)
					if kOk && vOk {
						httpReq.Header.Add(key, val)
					}
				}
			}
		}
	}

	// Add authorization
	if req.Auth != nil && req.Auth.Type != "none" {
		s.applyAuthorization(httpReq, req.Auth)
	}

	// Add query parameters
	if len(req.Params) > 0 {
		var paramsList []map[string]interface{}
		if err := json.Unmarshal(req.Params, &paramsList); err == nil {
			q := httpReq.URL.Query()
			for _, kv := range paramsList {
				enabled, ok := kv["enabled"].(bool)
				if !ok || enabled {
					key, kOk := kv["key"].(string)
					val, vOk := kv["value"].(string)
					if kOk && vOk {
						q.Add(key, val)
					}
				}
			}
			httpReq.URL.RawQuery = q.Encode()
		}
	}

	// Set content type if not set
	if httpReq.Header.Get("Content-Type") == "" && req.Body != "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	return httpReq, nil
}

// applyAuthorization adds authorization headers
func (s *requestService) applyAuthorization(req *http.Request, auth *entities.AuthorizationData) {
	switch auth.Type {
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+auth.BearerToken)
	case "basic":
		req.SetBasicAuth(auth.BasicUsername, auth.BasicPassword)
	case "api_key":
		if auth.APIKeyIn == "header" {
			req.Header.Set(auth.APIKeyKey, auth.APIKeyValue)
		} else if auth.APIKeyIn == "query" {
			q := req.URL.Query()
			q.Add(auth.APIKeyKey, auth.APIKeyValue)
			req.URL.RawQuery = q.Encode()
		}
	}
}

// replaceVariables replaces {{variable}} with environment values
func (s *requestService) replaceVariables(text string, vars map[string]string) string {
	if len(vars) == 0 {
		return text
	}

	for k, v := range vars {
		placeholder := fmt.Sprintf("{{%s}}", k)
		text = strings.ReplaceAll(text, placeholder, v)
	}
	return text
}
