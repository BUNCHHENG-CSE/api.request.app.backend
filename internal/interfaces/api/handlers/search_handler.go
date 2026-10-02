package handlers

import (
	"net/http"
	"strconv"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SearchHandler struct{ service application.SearchService }

func NewSearchHandler(service application.SearchService) *SearchHandler {
	return &SearchHandler{service: service}
}

func (h *SearchHandler) Search(c *gin.Context) {
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	var workspaceID *uuid.UUID
	if raw := c.Query("workspace_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid workspace ID")
			return
		}
		workspaceID = &parsed
	}
	results, err := h.service.Search(c.Request.Context(), ownerID, workspaceID, c.Query("q"), limit)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Search completed successfully", results)
}
