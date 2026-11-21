package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ImporterHandler struct {
	service application.ImporterService
}

func NewImporterHandler(service application.ImporterService) *ImporterHandler {
	return &ImporterHandler{service: service}
}

func (h *ImporterHandler) ImportCurl(c *gin.Context) {
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var req struct {
		FlowID uuid.UUID `json:"flow_id" binding:"required"`
		Curl   string    `json:"curl" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.service.ImportCurl(c.Request.Context(), ownerID, req.FlowID, req.Curl)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "cURL command imported successfully", created)
}

func (h *ImporterHandler) ImportPostman(c *gin.Context) {
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var workspaceID uuid.UUID
	var fileBytes []byte

	// Check if submitted as multipart/form-data
	file, fileHeader, err := c.Request.FormFile("file")
	if err == nil && file != nil {
		defer file.Close()
		wsIDStr := c.PostForm("workspace_id")
		parsedID, parseErr := uuid.Parse(wsIDStr)
		if parseErr != nil {
			response.Error(c, http.StatusBadRequest, "invalid workspace_id in form")
			return
		}
		workspaceID = parsedID
		data, readErr := io.ReadAll(file)
		if readErr != nil {
			response.Error(c, http.StatusBadRequest, "failed to read uploaded file")
			return
		}
		fileBytes = data
		_ = fileHeader
	} else {
		// Or submitted as raw JSON body
		var req struct {
			WorkspaceID uuid.UUID `json:"workspace_id" binding:"required"`
			Collection  any       `json:"collection" binding:"required"`
		}
		if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
			response.Error(c, http.StatusBadRequest, bindErr.Error())
			return
		}
		workspaceID = req.WorkspaceID
		rawJSON, marshalErr := json.Marshal(req.Collection)
		if marshalErr != nil {
			response.Error(c, http.StatusBadRequest, "invalid collection json")
			return
		}
		fileBytes = rawJSON
	}

	result, err := h.service.ImportPostman(c.Request.Context(), ownerID, workspaceID, fileBytes)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Postman collection imported successfully", result)
}
