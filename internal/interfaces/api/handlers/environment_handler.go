package handlers

import (
	"backend/internal/application"
	"net/http"

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
		WorkspaceID uint   `json:"workspace_id" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Variables   string `json:"variables" binding:"required"` // Assuming JSON string payload
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	env, err := h.service.CreateEnvironment(c.Request.Context(), req.WorkspaceID, req.Name, req.Variables)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, env)
}
