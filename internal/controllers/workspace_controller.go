package controllers

import (
	"net/http"

	"backend/internal/models"
	"backend/internal/services"

	"github.com/gin-gonic/gin"
)

// WorkspaceController handles workspace-related HTTP endpoints
type WorkspaceController struct {
	service services.WorkspaceService
}

func NewWorkspaceController(service services.WorkspaceService) *WorkspaceController {
	return &WorkspaceController{service: service}
}

// CreateWorkspace POST /api/workspaces
func (c *WorkspaceController) CreateWorkspace(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, models.ErrorResponseDTO{
			Error:   "UNAUTHORIZED",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	var dto models.CreateWorkspaceDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponseDTO{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	workspace, err := c.service.CreateWorkspace(ctx, userID.(string), &dto)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "CREATE_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusCreated, workspace)
}

// GetWorkspace GET /api/workspaces/:id
func (c *WorkspaceController) GetWorkspace(ctx *gin.Context) {
	id := ctx.Param("id")

	workspace, err := c.service.GetWorkspace(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.ErrorResponseDTO{
			Error:   "NOT_FOUND",
			Message: "Workspace not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	ctx.JSON(http.StatusOK, workspace)
}

// ListWorkspaces GET /api/workspaces
func (c *WorkspaceController) ListWorkspaces(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, models.ErrorResponseDTO{
			Error:   "UNAUTHORIZED",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	workspaces, err := c.service.ListWorkspaces(ctx, userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "LIST_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": workspaces})
}

// UpdateWorkspace PUT /api/workspaces/:id
func (c *WorkspaceController) UpdateWorkspace(ctx *gin.Context) {
	id := ctx.Param("id")

	var dto models.UpdateWorkspaceDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponseDTO{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	workspace, err := c.service.UpdateWorkspace(ctx, id, &dto)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "UPDATE_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, workspace)
}

// DeleteWorkspace DELETE /api/workspaces/:id
func (c *WorkspaceController) DeleteWorkspace(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.service.DeleteWorkspace(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "DELETE_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Workspace deleted"})
}

// AddMember POST /api/workspaces/:id/members
func (c *WorkspaceController) AddMember(ctx *gin.Context) {
	id := ctx.Param("id")

	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required,oneof=owner editor viewer"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponseDTO{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := c.service.AddMember(ctx, id, req.UserID, req.Role); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "ADD_MEMBER_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Member added"})
}

// RemoveMember DELETE /api/workspaces/:id/members/:userId
func (c *WorkspaceController) RemoveMember(ctx *gin.Context) {
	id := ctx.Param("id")
	userID := ctx.Param("userId")

	if err := c.service.RemoveMember(ctx, id, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "REMOVE_MEMBER_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Member removed"})
}

// GetMembers GET /api/workspaces/:id/members
func (c *WorkspaceController) GetMembers(ctx *gin.Context) {
	id := ctx.Param("id")

	members, err := c.service.GetMembers(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "GET_MEMBERS_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": members})
}

// UpdateMemberRole PUT /api/workspaces/:id/members/:userId/role
func (c *WorkspaceController) UpdateMemberRole(ctx *gin.Context) {
	id := ctx.Param("id")
	userID := ctx.Param("userId")

	var req struct {
		Role string `json:"role" binding:"required,oneof=owner editor viewer"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponseDTO{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := c.service.UpdateMemberRole(ctx, id, userID, req.Role); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "UPDATE_ROLE_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Role updated"})
}
