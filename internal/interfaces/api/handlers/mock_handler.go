package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"api.request.app.backend/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MockHandler struct {
	workspaceRepo domain.WorkspaceRepository
	flowRepo      domain.FlowRepository
	requestRepo   domain.RequestRepository
}

func NewMockHandler(workspaceRepo domain.WorkspaceRepository, flowRepo domain.FlowRepository, requestRepo domain.RequestRepository) *MockHandler {
	return &MockHandler{
		workspaceRepo: workspaceRepo,
		flowRepo:      flowRepo,
		requestRepo:   requestRepo,
	}
}

// ServeMock handles incoming mock API calls for a workspace: ANY /api/v1/mock/:workspaceId/*path
func (h *MockHandler) ServeMock(c *gin.Context) {
	workspaceIDStr := c.Param("workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace ID"})
		return
	}

	subPath := c.Param("path")
	if !strings.HasPrefix(subPath, "/") {
		subPath = "/" + subPath
	}
	method := strings.ToUpper(c.Request.Method)

	flows, err := h.flowRepo.ListByWorkspace(c.Request.Context(), workspaceID)
	if err != nil || len(flows) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":        "workspace not found or contains no flows",
			"workspace_id": workspaceID,
		})
		return
	}

	var matchedRequest *domain.Request

	for _, flow := range flows {
		requests, err := h.requestRepo.ListByFlow(c.Request.Context(), flow.ID)
		if err != nil {
			continue
		}
		for _, req := range requests {
			if !strings.EqualFold(req.Method, method) {
				continue
			}
			reqURL, err := url.Parse(req.URL)
			if err != nil {
				continue
			}
			// Match exact path or wildcard suffix match
			cleanReqPath := strings.TrimRight(reqURL.Path, "/")
			cleanSubPath := strings.TrimRight(subPath, "/")

			if strings.EqualFold(cleanReqPath, cleanSubPath) ||
				strings.HasSuffix(cleanReqPath, cleanSubPath) ||
				cleanSubPath == "" && cleanReqPath == "" {
				matchedReqCopy := req
				matchedRequest = &matchedReqCopy
				break
			}
		}
		if matchedRequest != nil {
			break
		}
	}

	if matchedRequest == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":        "no mock route found matching request",
			"method":       method,
			"path":         subPath,
			"workspace_id": workspaceID,
		})
		return
	}

	// Read mock settings if configured
	mockStatus := http.StatusOK
	mockContentType := "application/json"
	mockBody := ""

	if matchedRequest.Settings != "" {
		var settings map[string]any
		if err := json.Unmarshal([]byte(matchedRequest.Settings), &settings); err == nil {
			if s, ok := settings["mock_status"].(float64); ok && s > 0 {
				mockStatus = int(s)
			}
			if ct, ok := settings["mock_content_type"].(string); ok && ct != "" {
				mockContentType = ct
			}
			if mb, ok := settings["mock_body"].(string); ok && mb != "" {
				mockBody = mb
			}
			if headers, ok := settings["mock_headers"].(map[string]any); ok {
				for k, v := range headers {
					c.Header(k, fmt.Sprintf("%v", v))
				}
			}
		}
	}

	c.Header("X-Powered-By", "FlowAPI-Mock-Server")
	c.Header("X-Mock-Request-ID", matchedRequest.ID.String())

	if mockBody != "" {
		c.Data(mockStatus, mockContentType, []byte(mockBody))
		return
	}

	// Default intelligent mock response
	if matchedRequest.Body != "" && matchedRequest.Method != http.MethodGet {
		c.Data(mockStatus, mockContentType, []byte(matchedRequest.Body))
		return
	}

	c.JSON(mockStatus, gin.H{
		"mocked":       true,
		"request_id":   matchedRequest.ID,
		"name":         matchedRequest.Name,
		"method":       method,
		"path":         subPath,
		"workspace_id": workspaceID,
	})
}
