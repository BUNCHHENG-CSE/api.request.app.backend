package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type ImporterService interface {
	ImportCurl(ctx context.Context, ownerID, flowID uuid.UUID, curlCommand string) (*domain.Request, error)
	ImportPostman(ctx context.Context, ownerID, workspaceID uuid.UUID, data []byte) (*ImportPostmanResult, error)
}

type ImportPostmanResult struct {
	FlowID         uuid.UUID `json:"flow_id"`
	FlowName       string    `json:"flow_name"`
	RequestsCount  int       `json:"requests_count"`
	CollectionName string    `json:"collection_name"`
}

type importerService struct {
	requestService RequestService
	flowService    FlowService
	workspaceRepo  domain.WorkspaceRepository
	flowRepo       domain.FlowRepository
	requestRepo    domain.RequestRepository
	collectionRepo domain.CollectionRepository
}

func NewImporterService(requestService RequestService, flowService FlowService, workspaceRepo domain.WorkspaceRepository, flowRepo domain.FlowRepository, requestRepo domain.RequestRepository, collectionRepo domain.CollectionRepository) ImporterService {
	return &importerService{
		requestService: requestService,
		flowService:    flowService,
		workspaceRepo:  workspaceRepo,
		flowRepo:       flowRepo,
		requestRepo:    requestRepo,
		collectionRepo: collectionRepo,
	}
}

// ImportCurl parses a standard curl command into a stored Request
func (s *importerService) ImportCurl(ctx context.Context, ownerID, flowID uuid.UUID, curlCommand string) (*domain.Request, error) {
	curlCommand = strings.TrimSpace(curlCommand)
	if curlCommand == "" {
		return nil, errors.New("cURL command cannot be empty")
	}

	method, targetURL, headers, body, auth, err := parseCurl(curlCommand)
	if err != nil {
		return nil, err
	}

	name := "Imported Request"
	if parsedURL, err := url.Parse(targetURL); err == nil && parsedURL.Path != "" {
		parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
		if len(parts) > 0 && parts[len(parts)-1] != "" {
			name = strings.Title(strings.ReplaceAll(parts[len(parts)-1], "-", " "))
		}
	}

	bodyType := "none"
	if body != "" {
		if strings.HasPrefix(strings.TrimSpace(body), "{") || strings.HasPrefix(strings.TrimSpace(body), "[") {
			bodyType = "json"
		} else {
			bodyType = "text"
		}
	}

	encodedHeaders := "{}"
	if len(headers) > 0 {
		hBytes, _ := json.Marshal(headers)
		encodedHeaders = string(hBytes)
	}

	input := RequestInput{
		FlowID:   flowID,
		Name:     name,
		Method:   method,
		URL:      targetURL,
		Headers:  encodedHeaders,
		Body:     body,
		BodyType: bodyType,
		Auth:     auth,
	}

	return s.requestService.CreateRequest(ctx, ownerID, input)
}

func parseCurl(cmd string) (method, rawURL string, headers map[string]string, body, auth string, err error) {
	tokens := tokenizeCommand(cmd)
	if len(tokens) == 0 {
		return "", "", nil, "", "", errors.New("invalid cURL command")
	}

	method = "GET"
	headers = make(map[string]string)
	hasData := false

	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		switch t {
		case "-X", "--request":
			if i+1 < len(tokens) {
				method = strings.ToUpper(tokens[i+1])
				i++
			}
		case "-H", "--header":
			if i+1 < len(tokens) {
				h := tokens[i+1]
				parts := strings.SplitN(h, ":", 2)
				if len(parts) == 2 {
					k := strings.TrimSpace(parts[0])
					v := strings.TrimSpace(parts[1])
					if strings.EqualFold(k, "Authorization") && strings.HasPrefix(strings.ToLower(v), "bearer ") {
						token := strings.TrimSpace(v[7:])
						authBytes, _ := json.Marshal(map[string]string{"type": "bearer", "token": token})
						auth = string(authBytes)
					}
					headers[k] = v
				}
				i++
			}
		case "-d", "--data", "--data-raw", "--data-binary":
			if i+1 < len(tokens) {
				body = tokens[i+1]
				hasData = true
				i++
			}
		case "-u", "--user":
			if i+1 < len(tokens) {
				userPass := tokens[i+1]
				parts := strings.SplitN(userPass, ":", 2)
				if len(parts) == 2 {
					authBytes, _ := json.Marshal(map[string]string{
						"type":     "basic",
						"username": parts[0],
						"password": parts[1],
					})
					auth = string(authBytes)
				}
				i++
			}
		default:
			if strings.HasPrefix(t, "http://") || strings.HasPrefix(t, "https://") {
				rawURL = t
			}
		}
	}

	if rawURL == "" {
		return "", "", nil, "", "", errors.New("cURL command must contain an HTTP or HTTPS URL")
	}
	if hasData && method == "GET" {
		method = "POST"
	}

	return method, rawURL, headers, body, auth, nil
}

func tokenizeCommand(cmd string) []string {
	var tokens []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	runes := []rune(cmd)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' && !inSingle {
			escaped = true
			continue
		}
		if r == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}
		if r == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}
		if (r == ' ' || r == '\t' || r == '\n' || r == '\r') && !inSingle && !inDouble {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteRune(r)
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

// Postman Collection v2.1 models
type postmanCollection struct {
	Info struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Schema      string `json:"schema"`
	} `json:"info"`
	Item []postmanItem `json:"item"`
}

type postmanItem struct {
	Name    string          `json:"name"`
	Item    []postmanItem   `json:"item,omitempty"` // nested folders
	Request *postmanRequest `json:"request,omitempty"`
}

type postmanRequest struct {
	Method string `json:"method"`
	Header []struct {
		Key      string `json:"key"`
		Value    string `json:"value"`
		Disabled bool   `json:"disabled"`
	} `json:"header"`
	Body struct {
		Mode string `json:"mode"`
		Raw  string `json:"raw"`
	} `json:"body"`
	URL any `json:"url"` // string or object
	Auth *struct {
		Type   string `json:"type"`
		Bearer []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"bearer"`
	} `json:"auth"`
}

func (s *importerService) ImportPostman(ctx context.Context, ownerID, workspaceID uuid.UUID, data []byte) (*ImportPostmanResult, error) {
	if _, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID); err != nil {
		return nil, err
	}

	var collection postmanCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("failed to parse Postman collection JSON: %w", err)
	}

	flowName := strings.TrimSpace(collection.Info.Name)
	if flowName == "" {
		flowName = "Imported Postman Flow"
	}

	flow, err := s.flowService.CreateFlow(ctx, ownerID, workspaceID, flowName, collection.Info.Description, "[]", "[]")
	if err != nil {
		return nil, err
	}

	// Also create a matching Collection container
	col := &domain.Collection{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        flowName,
		Description: collection.Info.Description,
	}
	_ = s.collectionRepo.Create(ctx, col)

	var flattened []postmanItem
	flattenItems(collection.Item, &flattened)

	importedCount := 0
	for pos, item := range flattened {
		if item.Request == nil {
			continue
		}
		rawURL := extractPostmanURL(item.Request.URL)
		if rawURL == "" {
			continue
		}
		method := strings.ToUpper(strings.TrimSpace(item.Request.Method))
		if method == "" {
			method = "GET"
		}

		headersMap := make(map[string]string)
		for _, h := range item.Request.Header {
			if !h.Disabled && h.Key != "" {
				headersMap[h.Key] = h.Value
			}
		}
		headersJSON, _ := json.Marshal(headersMap)

		body := item.Request.Body.Raw
		bodyType := "none"
		if body != "" {
			bodyType = "text"
			if strings.HasPrefix(strings.TrimSpace(body), "{") || strings.HasPrefix(strings.TrimSpace(body), "[") {
				bodyType = "json"
			}
		}

		authJSON := "{}"
		if item.Request.Auth != nil && item.Request.Auth.Type == "bearer" && len(item.Request.Auth.Bearer) > 0 {
			token := ""
			for _, b := range item.Request.Auth.Bearer {
				if b.Key == "token" {
					token = b.Value
				}
			}
			if token != "" {
				authBytes, _ := json.Marshal(map[string]string{"type": "bearer", "token": token})
				authJSON = string(authBytes)
			}
		}

		reqInput := RequestInput{
			FlowID:       flow.ID,
			CollectionID: &col.ID,
			Name:         item.Name,
			Position:     pos,
			Method:       method,
			URL:          rawURL,
			Headers:      string(headersJSON),
			Body:         body,
			BodyType:     bodyType,
			Auth:         authJSON,
		}

		if _, err := s.requestService.CreateRequest(ctx, ownerID, reqInput); err == nil {
			importedCount++
		}
	}

	return &ImportPostmanResult{
		FlowID:         flow.ID,
		FlowName:       flow.Name,
		RequestsCount:  importedCount,
		CollectionName: col.Name,
	}, nil
}

func flattenItems(items []postmanItem, out *[]postmanItem) {
	for _, item := range items {
		if len(item.Item) > 0 {
			flattenItems(item.Item, out)
		} else if item.Request != nil {
			*out = append(*out, item)
		}
	}
}

func extractPostmanURL(raw any) string {
	if s, ok := raw.(string); ok {
		return s
	}
	if m, ok := raw.(map[string]any); ok {
		if rawStr, ok := m["raw"].(string); ok && rawStr != "" {
			return rawStr
		}
	}
	return ""
}
