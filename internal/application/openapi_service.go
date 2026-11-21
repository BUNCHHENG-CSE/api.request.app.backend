package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type OpenAPIImportResult struct {
	Collection domain.Collection `json:"collection"`
	Flow       domain.Flow       `json:"flow"`
	Requests   []domain.Request  `json:"requests"`
}

type OpenAPIService interface {
	ImportOpenAPI(ctx context.Context, ownerID, workspaceID uuid.UUID, specData []byte) (*OpenAPIImportResult, error)
	ExportWorkspaceOpenAPI(ctx context.Context, ownerID, workspaceID uuid.UUID) ([]byte, error)
	ExportCollectionOpenAPI(ctx context.Context, ownerID, collectionID uuid.UUID) ([]byte, error)
}

type openAPIService struct {
	workspaceRepo  domain.WorkspaceRepository
	collectionRepo domain.CollectionRepository
	flowRepo       domain.FlowRepository
	requestRepo    domain.RequestRepository
}

func NewOpenAPIService(
	workspaceRepo domain.WorkspaceRepository,
	collectionRepo domain.CollectionRepository,
	flowRepo domain.FlowRepository,
	requestRepo domain.RequestRepository,
) OpenAPIService {
	return &openAPIService{
		workspaceRepo:  workspaceRepo,
		collectionRepo: collectionRepo,
		flowRepo:       flowRepo,
		requestRepo:    requestRepo,
	}
}

// openAPISpec internal representation for parsing 2.0 / 3.0 specs
type openAPISpec struct {
	OpenAPI string                            `yaml:"openapi" json:"openapi"`
	Swagger string                            `yaml:"swagger" json:"swagger"`
	Info    openAPIInfo                       `yaml:"info" json:"info"`
	Servers []openAPIServer                   `yaml:"servers" json:"servers"`
	Host    string                            `yaml:"host" json:"host"`
	Paths   map[string]map[string]openAPIOp   `yaml:"paths" json:"paths"`
}

type openAPIInfo struct {
	Title       string `yaml:"title" json:"title"`
	Description string `yaml:"description" json:"description"`
	Version     string `yaml:"version" json:"version"`
}

type openAPIServer struct {
	URL string `yaml:"url" json:"url"`
}

type openAPIOp struct {
	Summary     string              `yaml:"summary" json:"summary"`
	Description string              `yaml:"description" json:"description"`
	OperationID string              `yaml:"operationId" json:"operationId"`
	Parameters  []openAPIParam      `yaml:"parameters" json:"parameters"`
	RequestBody *openAPIRequestBody `yaml:"requestBody" json:"requestBody"`
}

type openAPIParam struct {
	Name     string `yaml:"name" json:"name"`
	In       string `yaml:"in" json:"in"` // "query", "header", "path"
	Required bool   `yaml:"required" json:"required"`
	Schema   any    `yaml:"schema" json:"schema"`
	Default  any    `yaml:"default" json:"default"`
}

type openAPIRequestBody struct {
	Content map[string]openAPIMediaType `yaml:"content" json:"content"`
}

type openAPIMediaType struct {
	Example any `yaml:"example" json:"example"`
	Schema  any `yaml:"schema" json:"schema"`
}

func (s *openAPIService) ImportOpenAPI(ctx context.Context, ownerID, workspaceID uuid.UUID, specData []byte) (*OpenAPIImportResult, error) {
	if _, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID); err != nil {
		return nil, err
	}

	var spec openAPISpec
	// Try parsing as YAML/JSON
	if err := yaml.Unmarshal(specData, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI specification: %w", err)
	}

	if spec.OpenAPI == "" && spec.Swagger == "" && len(spec.Paths) == 0 {
		return nil, errors.New("invalid or empty OpenAPI / Swagger specification")
	}

	title := strings.TrimSpace(spec.Info.Title)
	if title == "" {
		title = "Imported OpenAPI Collection"
	}

	baseURL := "http://localhost:8080"
	if len(spec.Servers) > 0 && spec.Servers[0].URL != "" {
		baseURL = strings.TrimRight(spec.Servers[0].URL, "/")
	} else if spec.Host != "" {
		baseURL = "https://" + strings.TrimRight(spec.Host, "/")
	}

	// Create collection
	col := &domain.Collection{
		WorkspaceID: workspaceID,
		Name:        title,
		Description: spec.Info.Description,
	}
	if err := s.collectionRepo.Create(ctx, col); err != nil {
		return nil, err
	}

	// Create a companion flow
	flow := &domain.Flow{
		WorkspaceID: workspaceID,
		Name:        title + " Flow",
		Description: "Flow generated from OpenAPI import",
	}
	if err := s.flowRepo.Create(ctx, flow); err != nil {
		return nil, err
	}

	createdRequests := make([]domain.Request, 0)
	position := 0

	// Sort paths for deterministic ordering
	paths := make([]string, 0, len(spec.Paths))
	for p := range spec.Paths {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	validMethods := map[string]bool{
		"get": true, "post": true, "put": true, "delete": true,
		"patch": true, "options": true, "head": true,
	}

	for _, pathStr := range paths {
		methods := spec.Paths[pathStr]
		for method, op := range methods {
			methodLower := strings.ToLower(method)
			if !validMethods[methodLower] {
				continue
			}

			reqName := strings.TrimSpace(op.Summary)
			if reqName == "" {
				reqName = strings.TrimSpace(op.OperationID)
			}
			if reqName == "" {
				reqName = fmt.Sprintf("%s %s", strings.ToUpper(methodLower), pathStr)
			}

			// Headers and query params
			headers := make([]map[string]any, 0)
			queryParams := make([]map[string]any, 0)

			for _, p := range op.Parameters {
				switch strings.ToLower(p.In) {
				case "header":
					headers = append(headers, map[string]any{
						"key":     p.Name,
						"value":   "",
						"enabled": p.Required,
					})
				case "query":
					queryParams = append(queryParams, map[string]any{
						"key":     p.Name,
						"value":   "",
						"enabled": p.Required,
					})
				}
			}

			// Body
			bodyStr := ""
			bodyType := "none"
			if op.RequestBody != nil {
				if jsonContent, ok := op.RequestBody.Content["application/json"]; ok {
					bodyType = "json"
					headers = append(headers, map[string]any{
						"key":     "Content-Type",
						"value":   "application/json",
						"enabled": true,
					})
					if jsonContent.Example != nil {
						bBytes, _ := json.MarshalIndent(jsonContent.Example, "", "  ")
						bodyStr = string(bBytes)
					} else if jsonContent.Schema != nil {
						bBytes, _ := json.MarshalIndent(jsonContent.Schema, "", "  ")
						bodyStr = string(bBytes)
					}
				}
			}

			headersJSON, _ := json.Marshal(headers)
			queryJSON, _ := json.Marshal(queryParams)

			fullURL := baseURL + pathStr

			reqEntity := domain.Request{
				FlowID:       flow.ID,
				CollectionID: &col.ID,
				Name:         reqName,
				Position:     position,
				Method:       strings.ToUpper(methodLower),
				URL:          fullURL,
				Headers:      string(headersJSON),
				QueryParams:  string(queryJSON),
				Body:         bodyStr,
				BodyType:     bodyType,
			}

			if err := s.requestRepo.Create(ctx, &reqEntity); err != nil {
				return nil, err
			}

			createdRequests = append(createdRequests, reqEntity)
			position++
		}
	}

	return &OpenAPIImportResult{
		Collection: *col,
		Flow:       *flow,
		Requests:   createdRequests,
	}, nil
}

func (s *openAPIService) ExportWorkspaceOpenAPI(ctx context.Context, ownerID, workspaceID uuid.UUID) ([]byte, error) {
	ws, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID)
	if err != nil {
		return nil, err
	}

	flows, err := s.flowRepo.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var allRequests []domain.Request
	for _, f := range flows {
		reqs, err := s.requestRepo.ListByFlow(ctx, f.ID)
		if err == nil {
			allRequests = append(allRequests, reqs...)
		}
	}

	return s.buildOpenAPISpec(ws.Name, allRequests)
}

func (s *openAPIService) ExportCollectionOpenAPI(ctx context.Context, ownerID, collectionID uuid.UUID) ([]byte, error) {
	col, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return nil, err
	}
	if _, err := s.workspaceRepo.GetByID(ctx, col.WorkspaceID, ownerID); err != nil {
		return nil, err
	}

	flows, err := s.flowRepo.ListByWorkspace(ctx, col.WorkspaceID)
	if err != nil {
		return nil, err
	}

	var colRequests []domain.Request
	for _, f := range flows {
		reqs, err := s.requestRepo.ListByFlow(ctx, f.ID)
		if err == nil {
			for _, r := range reqs {
				if r.CollectionID != nil && *r.CollectionID == collectionID {
					colRequests = append(colRequests, r)
				}
			}
		}
	}

	return s.buildOpenAPISpec(col.Name, colRequests)
}

func (s *openAPIService) buildOpenAPISpec(title string, requests []domain.Request) ([]byte, error) {
	pathsMap := make(map[string]map[string]any)
	baseURL := "http://localhost:8080"

	for _, r := range requests {
		parsedURL, err := url.Parse(r.URL)
		reqPath := "/"
		if err == nil {
			if parsedURL.Host != "" {
				baseURL = parsedURL.Scheme + "://" + parsedURL.Host
			}
			if parsedURL.Path != "" {
				reqPath = parsedURL.Path
			}
		} else if strings.HasPrefix(r.URL, "/") {
			reqPath = r.URL
		}

		if pathsMap[reqPath] == nil {
			pathsMap[reqPath] = make(map[string]any)
		}

		methodLower := strings.ToLower(r.Method)
		if methodLower == "" {
			methodLower = "get"
		}

		var params []map[string]any

		// Parse query params
		if r.QueryParams != "" {
			var qp []map[string]any
			if err := json.Unmarshal([]byte(r.QueryParams), &qp); err == nil {
				for _, item := range qp {
					if key, ok := item["key"].(string); ok && key != "" {
						params = append(params, map[string]any{
							"name":     key,
							"in":       "query",
							"required": false,
							"schema":   map[string]any{"type": "string"},
						})
					}
				}
			}
		}

		// Parse headers
		if r.Headers != "" {
			var hp []map[string]any
			if err := json.Unmarshal([]byte(r.Headers), &hp); err == nil {
				for _, item := range hp {
					if key, ok := item["key"].(string); ok && key != "" && !strings.EqualFold(key, "Content-Type") {
						params = append(params, map[string]any{
							"name":     key,
							"in":       "header",
							"required": false,
							"schema":   map[string]any{"type": "string"},
						})
					}
				}
			}
		}

		op := map[string]any{
			"summary":    r.Name,
			"parameters": params,
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Successful response",
				},
			},
		}

		if strings.TrimSpace(r.Body) != "" && (r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH") {
			var bodyObj any
			if err := json.Unmarshal([]byte(r.Body), &bodyObj); err == nil {
				op["requestBody"] = map[string]any{
					"required": true,
					"content": map[string]any{
						"application/json": map[string]any{
							"example": bodyObj,
						},
					},
				}
			} else {
				op["requestBody"] = map[string]any{
					"required": true,
					"content": map[string]any{
						"text/plain": map[string]any{
							"example": r.Body,
						},
					},
				}
			}
		}

		pathsMap[reqPath][methodLower] = op
	}

	doc := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":   title,
			"version": "1.0.0",
		},
		"servers": []map[string]any{
			{"url": baseURL},
		},
		"paths": pathsMap,
	}

	return json.MarshalIndent(doc, "", "  ")
}
