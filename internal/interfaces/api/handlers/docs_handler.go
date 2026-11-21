package handlers

import (
	"net/http"

	"api.request.app.backend/internal/application"
	"github.com/gin-gonic/gin"
)

type DocsHandler struct {
	service application.DocsService
}

func NewDocsHandler(service application.DocsService) *DocsHandler {
	return &DocsHandler{service: service}
}

func (h *DocsHandler) GenerateCollectionDocs(c *gin.Context) {
	collectionID, ok := pathUUID(c, "id", "collection")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	format := c.DefaultQuery("format", "markdown")
	content, contentType, err := h.service.GenerateCollectionDocs(c.Request.Context(), ownerID, collectionID, format)
	if err != nil {
		serviceError(c, err)
		return
	}

	c.Data(http.StatusOK, contentType, []byte(content))
}

func (h *DocsHandler) GenerateWorkspaceDocs(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	format := c.DefaultQuery("format", "markdown")
	content, contentType, err := h.service.GenerateWorkspaceDocs(c.Request.Context(), ownerID, workspaceID, format)
	if err != nil {
		serviceError(c, err)
		return
	}

	c.Data(http.StatusOK, contentType, []byte(content))
}
