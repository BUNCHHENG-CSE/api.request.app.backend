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
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	ownerID, ok := c.Get("user_id")
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authenticated user is required")
		return
	}
	workspace, err := h.service.CreateWorkspace(c.Request.Context(), req.Name, ownerID.(uuid.UUID))
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

	ownerID, ok := c.Get("user_id")
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authenticated user is required")
		return
	}
	workspace, err := h.service.GetWorkspace(c.Request.Context(), id, ownerID.(uuid.UUID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "workspace not found")
		return
	}

	response.Success(c, http.StatusOK, "Get workspace successfully!!!", workspace)
}

func (h *WorkspaceHandler) List(c *gin.Context) {
	ownerID, ok := c.Get("user_id")
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authenticated user is required")
		return
	}
	workspaces, err := h.service.ListWorkspaces(c.Request.Context(), ownerID.(uuid.UUID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Workspaces retrieved successfully", workspaces)
}

func (h *WorkspaceHandler) Update(c *gin.Context) {
	id, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	workspace, err := h.service.UpdateWorkspace(c.Request.Context(), id, ownerID, req.Name)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Workspace updated successfully", workspace)
}

func (h *WorkspaceHandler) Delete(c *gin.Context) {
	id, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteWorkspace(c.Request.Context(), id, ownerID); err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Workspace deleted successfully", gin.H{})
}
