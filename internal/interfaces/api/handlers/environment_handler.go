package handlers

import (
	"net/http"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type EnvironmentHandler struct {
	service application.EnvironmentService
}

func NewEnvironmentHandler(service application.EnvironmentService) *EnvironmentHandler {
	return &EnvironmentHandler{service: service}
}

func (h *EnvironmentHandler) Create(c *gin.Context) {
	var req struct {
		WorkspaceID uuid.UUID `json:"workspace_id" binding:"required"`
		Name        string    `json:"name" binding:"required"`
		Variables   string    `json:"variables" binding:"required"` // Assuming JSON string payload
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	env, err := h.service.CreateEnvironment(c.Request.Context(), req.WorkspaceID, req.Name, req.Variables)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Environment created successfully !!!", env)
}
