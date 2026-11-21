package handlers

import (
	"io"
	"net/http"
	"strings"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OpenAPIHandler struct {
	service application.OpenAPIService
}

func NewOpenAPIHandler(service application.OpenAPIService) *OpenAPIHandler {
	return &OpenAPIHandler{service: service}
}

func (h *OpenAPIHandler) ImportOpenAPI(c *gin.Context) {
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var workspaceID uuid.UUID
	var specBytes []byte

	// Check if multipart form
	if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
		wsIDStr := c.PostForm("workspace_id")
		if wsIDStr == "" {
			response.Error(c, http.StatusBadRequest, "workspace_id form field is required")
			return
		}
		var err error
		workspaceID, err = uuid.Parse(wsIDStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid workspace_id UUID")
			return
		}

		file, err := c.FormFile("file")
		if err != nil {
			response.Error(c, http.StatusBadRequest, "OpenAPI file upload is required")
			return
		}
		opened, err := file.Open()
		if err != nil {
			response.Error(c, http.StatusBadRequest, "failed to read uploaded file")
			return
		}
		defer opened.Close()
		specBytes, err = io.ReadAll(opened)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "failed to read file content")
			return
		}
	} else {
		// JSON payload
		var req struct {
			WorkspaceID uuid.UUID `json:"workspace_id"`
			Spec        string    `json:"spec"`
		}
		if err := c.ShouldBindJSON(&req); err == nil && req.WorkspaceID != uuid.Nil && req.Spec != "" {
			workspaceID = req.WorkspaceID
			specBytes = []byte(req.Spec)
		} else {
			// Query parameter with raw body
			wsIDStr := c.Query("workspace_id")
			if wsIDStr != "" {
				if parsedID, err := uuid.Parse(wsIDStr); err == nil {
					workspaceID = parsedID
					bodyBytes, _ := io.ReadAll(c.Request.Body)
					specBytes = bodyBytes
				}
			}
		}
	}

	if workspaceID == uuid.Nil || len(specBytes) == 0 {
		response.Error(c, http.StatusBadRequest, "valid workspace_id and OpenAPI specification content are required")
		return
	}

	res, err := h.service.ImportOpenAPI(c.Request.Context(), ownerID, workspaceID, specBytes)
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "OpenAPI imported successfully", res)
}

func (h *OpenAPIHandler) ExportWorkspace(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	specData, err := h.service.ExportWorkspaceOpenAPI(c.Request.Context(), ownerID, workspaceID)
	if err != nil {
		serviceError(c, err)
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\"openapi-spec.json\"")
	c.Data(http.StatusOK, "application/json; charset=utf-8", specData)
}

func (h *OpenAPIHandler) ExportCollection(c *gin.Context) {
	collectionID, ok := pathUUID(c, "id", "collection")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	specData, err := h.service.ExportCollectionOpenAPI(c.Request.Context(), ownerID, collectionID)
	if err != nil {
		serviceError(c, err)
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\"openapi-spec.json\"")
	c.Data(http.StatusOK, "application/json; charset=utf-8", specData)
}
