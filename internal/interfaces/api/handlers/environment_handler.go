package handlers

import (
	"net/http"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		Variables   string    `json:"variables"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	env, err := h.service.CreateEnvironment(c.Request.Context(), ownerID, req.WorkspaceID, req.Name, req.Variables)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Environment created successfully", env)
}

func (h *EnvironmentHandler) ListByWorkspace(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	environments, err := h.service.GetEnvironmentsByWorkspace(c.Request.Context(), ownerID, workspaceID)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Environments retrieved successfully", environments)
}

func (h *EnvironmentHandler) Get(c *gin.Context) {
	id, ok := pathUUID(c, "id", "environment")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	env, err := h.service.GetEnvironment(c.Request.Context(), ownerID, id)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Environment retrieved successfully", env)
}

func (h *EnvironmentHandler) Update(c *gin.Context) {
	id, ok := pathUUID(c, "id", "environment")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	var req struct {
		Name      *string `json:"name"`
		Variables *string `json:"variables"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	env, err := h.service.UpdateEnvironment(c.Request.Context(), ownerID, id, req.Name, req.Variables)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Environment updated successfully", env)
}

func (h *EnvironmentHandler) Delete(c *gin.Context) {
	id, ok := pathUUID(c, "id", "environment")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteEnvironment(c.Request.Context(), ownerID, id); err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Environment deleted successfully", gin.H{})
}
