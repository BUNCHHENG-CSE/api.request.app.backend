package handlers

import (
	"net/http"
	"strconv"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RequestHandler struct{ service application.RequestService }

func NewRequestHandler(service application.RequestService) *RequestHandler {
	return &RequestHandler{service: service}
}

func (h *RequestHandler) Create(c *gin.Context) {
	var req CreateStoredRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	created, err := h.service.CreateRequest(c.Request.Context(), ownerID, application.RequestInput{
		FlowID:       req.FlowID,
		CollectionID: req.CollectionID,
		Name:         req.Name,
		Position:     req.Position,
		Method:       req.Method,
		URL:          req.URL,
		Headers:      req.Headers,
		QueryParams:  req.QueryParams,
		Body:         req.Body,
		BodyType:     req.BodyType,
		Auth:         req.Auth,
		Scripts:      req.Scripts,
		Settings:     req.Settings,
	})
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Request created successfully", created)
}

func (h *RequestHandler) ListByFlow(c *gin.Context) {
	flowID, ok := pathUUID(c, "id", "flow")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	requests, err := h.service.GetRequestsByFlow(c.Request.Context(), ownerID, flowID)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Requests retrieved successfully", requests)
}

func (h *RequestHandler) Get(c *gin.Context) {
	id, ok := pathUUID(c, "id", "request")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	req, err := h.service.GetRequest(c.Request.Context(), ownerID, id)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Request retrieved successfully", req)
}

func (h *RequestHandler) Update(c *gin.Context) {
	id, ok := pathUUID(c, "id", "request")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	var req PatchStoredRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	updated, err := h.service.UpdateRequest(c.Request.Context(), ownerID, id, application.RequestPatch{
		FlowID:      req.FlowID,
		Name:        req.Name,
		Position:    req.Position,
		Method:      req.Method,
		URL:         req.URL,
		Headers:     req.Headers,
		QueryParams: req.QueryParams,
		Body:        req.Body,
		BodyType:    req.BodyType,
		Auth:        req.Auth,
		Scripts:     req.Scripts,
		Settings:    req.Settings,
	})
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Request updated successfully", updated)
}

func (h *RequestHandler) Delete(c *gin.Context) {
	id, ok := pathUUID(c, "id", "request")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteRequest(c.Request.Context(), ownerID, id); err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Request deleted successfully", gin.H{})
}

func (h *RequestHandler) Execute(c *gin.Context) {
	id, ok := pathUUID(c, "id", "request")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	var req struct {
		EnvironmentID *uuid.UUID `json:"environment_id"`
	}
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	}
	result, err := h.service.ExecuteRequest(c.Request.Context(), ownerID, id, req.EnvironmentID)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Request executed successfully", result)
}

func (h *RequestHandler) Duplicate(c *gin.Context) {
	id, ok := pathUUID(c, "id", "request")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	req, err := h.service.DuplicateRequest(c.Request.Context(), ownerID, id)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Request duplicated successfully", req)
}

func (h *RequestHandler) BulkDelete(c *gin.Context) {
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	var req struct {
		IDs []uuid.UUID `json:"ids" binding:"required,min=1,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.service.BulkDeleteRequests(c.Request.Context(), ownerID, req.IDs); err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Requests deleted successfully", gin.H{"deleted": len(req.IDs)})
}

func (h *RequestHandler) ListHistory(c *gin.Context) {
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var workspaceID *uuid.UUID
	if raw := c.Query("workspace_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid workspace ID")
			return
		}
		workspaceID = &parsed
	}
	items, total, err := h.service.ListHistory(c.Request.Context(), ownerID, workspaceID, page, limit)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "History retrieved successfully", gin.H{"items": items, "total": total, "page": page, "limit": limit})
}

func (h *RequestHandler) GetHistory(c *gin.Context) {
	id, ok := pathUUID(c, "id", "history")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	history, err := h.service.GetHistory(c.Request.Context(), ownerID, id)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "History entry retrieved successfully", history)
}

func (h *RequestHandler) ClearHistory(c *gin.Context) {
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	if err := h.service.ClearHistory(c.Request.Context(), ownerID); err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "History cleared successfully", gin.H{})
}

func (h *RequestHandler) CompareHistory(c *gin.Context) {
	id1Str := c.Query("id1")
	id2Str := c.Query("id2")
	if id1Str == "" || id2Str == "" {
		response.Error(c, http.StatusBadRequest, "Both id1 and id2 query parameters are required")
		return
	}
	id1, err := uuid.Parse(id1Str)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid id1 UUID")
		return
	}
	id2, err := uuid.Parse(id2Str)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid id2 UUID")
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	diff, err := h.service.CompareHistory(c.Request.Context(), ownerID, id1, id2)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "History comparison retrieved successfully", diff)
}
