package handlers

import (
	"net/http"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/domain"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WorkspaceMemberHandler struct {
	service application.WorkspaceMemberService
}

func NewWorkspaceMemberHandler(service application.WorkspaceMemberService) *WorkspaceMemberHandler {
	return &WorkspaceMemberHandler{service: service}
}

type AddMemberRequest struct {
	Identifier string               `json:"identifier" binding:"required"`
	Role       domain.WorkspaceRole `json:"role"`
}

type UpdateMemberRoleRequest struct {
	Role domain.WorkspaceRole `json:"role" binding:"required"`
}

func (h *WorkspaceMemberHandler) AddMember(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	requesterID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	member, err := h.service.AddMember(c.Request.Context(), requesterID, workspaceID, req.Identifier, req.Role)
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Member added successfully", member)
}

func (h *WorkspaceMemberHandler) ListMembers(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	requesterID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	members, err := h.service.ListMembers(c.Request.Context(), requesterID, workspaceID)
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Members retrieved successfully", members)
}

func (h *WorkspaceMemberHandler) UpdateRole(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	targetUserIDStr := c.Param("userId")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid target user ID")
		return
	}
	requesterID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var req UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	member, err := h.service.UpdateRole(c.Request.Context(), requesterID, workspaceID, targetUserID, req.Role)
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Member role updated successfully", member)
}

func (h *WorkspaceMemberHandler) RemoveMember(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	targetUserIDStr := c.Param("userId")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid target user ID")
		return
	}
	requesterID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	if err := h.service.RemoveMember(c.Request.Context(), requesterID, workspaceID, targetUserID); err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Member removed successfully", gin.H{})
}
