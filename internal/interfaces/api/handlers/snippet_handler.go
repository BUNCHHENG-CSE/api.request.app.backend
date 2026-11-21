package handlers

import (
	"net/http"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
)

type SnippetHandler struct {
	service application.SnippetService
}

func NewSnippetHandler(service application.SnippetService) *SnippetHandler {
	return &SnippetHandler{service: service}
}

func (h *SnippetHandler) Generate(c *gin.Context) {
	id, ok := pathUUID(c, "id", "request")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	language := c.DefaultQuery("language", "curl")
	snippet, err := h.service.Generate(c.Request.Context(), ownerID, id, language)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Code snippet generated successfully", gin.H{
		"request_id": id,
		"language":   language,
		"snippet":    snippet,
	})
}
