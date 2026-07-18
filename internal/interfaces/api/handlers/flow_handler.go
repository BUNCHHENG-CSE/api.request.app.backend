package handlers

import (
	"backend/internal/application"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FlowHandler struct {
	service application.FlowService
}

func NewFlowHandler(service application.FlowService) *FlowHandler {
	return &FlowHandler{service: service}
}
func (h *FlowHandler) Create(c *gin.Context) {
	// Optimized: A Flow is usually created inside a Workspace, requiring a Name.
	var req struct {
		WorkspaceID uint   `json:"workspace_id" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	// 1. Bind and validate the JSON payload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Pass to the Application Service layer
	flow, err := h.service.CreateFlow(c.Request.Context(), req.WorkspaceID, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. Return standardized response
	c.JSON(http.StatusCreated, flow)
}
