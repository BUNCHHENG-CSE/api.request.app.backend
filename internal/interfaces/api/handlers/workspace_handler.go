package handlers

import (
	"net/http"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/google/uuid"

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
		Name    string    `json:"name" binding:"required"`
		OwnerID uuid.UUID `json:"owner_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	workspace, err := h.service.CreateWorkspace(c.Request.Context(), req.Name, req.OwnerID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Workspace created successfully!!!", workspace)
}

func (h *WorkspaceHandler) Get(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return
	}

	workspace, err := h.service.GetWorkspace(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "workspace not found")
		return
	}

	response.Success(c, http.StatusOK, "Get workspace successfully!!!", workspace)
}
