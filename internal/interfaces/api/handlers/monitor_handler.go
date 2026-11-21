package handlers

import (
	"net/http"
	"strconv"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MonitorHandler struct {
	service application.MonitorService
}

func NewMonitorHandler(service application.MonitorService) *MonitorHandler {
	return &MonitorHandler{service: service}
}

type CreateMonitorRequest struct {
	WorkspaceID uuid.UUID `json:"workspace_id" binding:"required"`
	FlowID      uuid.UUID `json:"flow_id" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	IntervalMin int       `json:"interval_minutes"`
	WebhookURL  string    `json:"webhook_url"`
	IsActive    *bool     `json:"is_active"`
}

type UpdateMonitorRequest struct {
	Name        *string `json:"name"`
	IntervalMin *int    `json:"interval_minutes"`
	WebhookURL  *string `json:"webhook_url"`
	IsActive    *bool   `json:"is_active"`
}

func (h *MonitorHandler) Create(c *gin.Context) {
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var req CreateMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	monitor, err := h.service.CreateMonitor(c.Request.Context(), ownerID, application.MonitorInput{
		WorkspaceID: req.WorkspaceID,
		FlowID:      req.FlowID,
		Name:        req.Name,
		IntervalMin: req.IntervalMin,
		WebhookURL:  req.WebhookURL,
		IsActive:    req.IsActive,
	})
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Monitor created successfully", monitor)
}

func (h *MonitorHandler) List(c *gin.Context) {
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	workspaceIDStr := c.Query("workspace_id")
	if workspaceIDStr == "" {
		response.Error(c, http.StatusBadRequest, "workspace_id query parameter is required")
		return
	}
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid workspace_id UUID")
		return
	}

	monitors, err := h.service.ListByWorkspace(c.Request.Context(), ownerID, workspaceID)
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Monitors retrieved successfully", monitors)
}

func (h *MonitorHandler) Get(c *gin.Context) {
	id, ok := pathUUID(c, "id", "monitor")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	monitor, err := h.service.GetByID(c.Request.Context(), ownerID, id)
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Monitor retrieved successfully", monitor)
}

func (h *MonitorHandler) Update(c *gin.Context) {
	id, ok := pathUUID(c, "id", "monitor")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var req UpdateMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	monitor, err := h.service.UpdateMonitor(c.Request.Context(), ownerID, id, application.MonitorPatch{
		Name:        req.Name,
		IntervalMin: req.IntervalMin,
		WebhookURL:  req.WebhookURL,
		IsActive:    req.IsActive,
	})
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Monitor updated successfully", monitor)
}

func (h *MonitorHandler) Delete(c *gin.Context) {
	id, ok := pathUUID(c, "id", "monitor")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	if err := h.service.DeleteMonitor(c.Request.Context(), ownerID, id); err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Monitor deleted successfully", gin.H{})
}

func (h *MonitorHandler) Run(c *gin.Context) {
	id, ok := pathUUID(c, "id", "monitor")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	log, err := h.service.RunMonitor(c.Request.Context(), ownerID, id)
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Monitor executed successfully", log)
}

func (h *MonitorHandler) ListLogs(c *gin.Context) {
	id, ok := pathUUID(c, "id", "monitor")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	logs, err := h.service.ListRunLogs(c.Request.Context(), ownerID, id, limit)
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Monitor run logs retrieved successfully", logs)
}
