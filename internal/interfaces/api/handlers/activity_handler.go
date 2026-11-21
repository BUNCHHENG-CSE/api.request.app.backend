package handlers

import (
	"net/http"
	"strconv"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
)

type ActivityHandler struct {
	service application.ActivityService
}

func NewActivityHandler(service application.ActivityService) *ActivityHandler {
	return &ActivityHandler{service: service}
}

func (h *ActivityHandler) List(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	limit := 30
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	activities, err := h.service.List(c.Request.Context(), ownerID, workspaceID, limit)
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Activities retrieved successfully", activities)
}
