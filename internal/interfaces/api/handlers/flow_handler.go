package handlers

import (
	"net/http"
	"time"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FlowHandler struct {
	service        application.FlowService
	requestService application.RequestService
}

func NewFlowHandler(service application.FlowService, requestService application.RequestService) *FlowHandler {
	return &FlowHandler{service: service, requestService: requestService}
}

func (h *FlowHandler) Run(c *gin.Context) {
	id, ok := pathUUID(c, "id", "flow")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	if _, err := h.service.GetFlow(c.Request.Context(), ownerID, id); err != nil {
		serviceError(c, err)
		return
	}
	var req struct {
		EnvironmentID *uuid.UUID `json:"environment_id"`
		StopOnError   *bool      `json:"stop_on_error"`
	}
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	}
	stopOnError := true
	if req.StopOnError != nil {
		stopOnError = *req.StopOnError
	}
	requests, err := h.requestService.GetRequestsByFlow(c.Request.Context(), ownerID, id)
	if err != nil {
		serviceError(c, err)
		return
	}

	started := time.Now()
	results := make([]gin.H, 0, len(requests))
	passed := 0
	failed := 0
	totalTestsPassed := 0
	totalTestsFailed := 0

	for _, stored := range requests {
		result, runErr := h.requestService.ExecuteRequest(c.Request.Context(), ownerID, stored.ID, req.EnvironmentID)
		if runErr != nil {
			failed++
			results = append(results, gin.H{
				"request_id": stored.ID,
				"name":       stored.Name,
				"error":      runErr.Error(),
			})
			if stopOnError {
				break
			}
			continue
		}
		passed++
		totalTestsPassed += result.TestsPassed
		totalTestsFailed += result.TestsFailed
		results = append(results, gin.H{
			"request_id": stored.ID,
			"name":       stored.Name,
			"result":     result,
		})
	}

	response.Success(c, http.StatusOK, "Flow run completed", gin.H{
		"flow_id":            id,
		"results":            results,
		"total_requests":     len(requests),
		"passed":             passed,
		"failed":             failed,
		"total_tests_passed": totalTestsPassed,
		"total_tests_failed": totalTestsFailed,
		"duration_ms":        time.Since(started).Milliseconds(),
	})
}

func (h *FlowHandler) Create(c *gin.Context) {
	var req CreateFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	flow, err := h.service.CreateFlow(c.Request.Context(), ownerID, req.WorkspaceID, req.Name, req.Description, req.Nodes, req.Edges)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Flow created successfully", flow)
}

func (h *FlowHandler) ListByWorkspace(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	flows, err := h.service.GetFlowsByWorkspace(c.Request.Context(), ownerID, workspaceID)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Flows retrieved successfully", flows)
}

func (h *FlowHandler) Get(c *gin.Context) {
	id, ok := pathUUID(c, "id", "flow")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	flow, err := h.service.GetFlow(c.Request.Context(), ownerID, id)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Flow retrieved successfully", flow)
}

func (h *FlowHandler) Update(c *gin.Context) {
	id, ok := pathUUID(c, "id", "flow")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Nodes       *string `json:"nodes"`
		Edges       *string `json:"edges"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	flow, err := h.service.UpdateFlow(c.Request.Context(), ownerID, id, req.Name, req.Description, req.Nodes, req.Edges)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Flow updated successfully", flow)
}

func (h *FlowHandler) Delete(c *gin.Context) {
	id, ok := pathUUID(c, "id", "flow")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteFlow(c.Request.Context(), ownerID, id); err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Flow deleted successfully", gin.H{})
}
