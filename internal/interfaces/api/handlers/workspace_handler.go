package handlers

import (
	"net/http"
	"strconv"

	"backend/internal/application"

	"github.com/gin-gonic/gin"
)

type WorkspaceHandler struct {
	service application.WorkspaceService
}

func NewWorkspaceHandler(service application.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{service: service}
}

func (h *WorkspaceHandler) Create(c *gin.Context) {
	var req struct {
		Name    string `json:"name" binding:"required"`
		OwnerID uint   `json:"owner_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspace, err := h.service.CreateWorkspace(c.Request.Context(), req.Name, req.OwnerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, workspace)
}

func (h *WorkspaceHandler) Get(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	workspace, err := h.service.GetWorkspace(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}

	c.JSON(http.StatusOK, workspace)
}
