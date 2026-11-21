package handlers

import (
	"net/http"
	"time"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	flowService    application.FlowService
	requestService application.RequestService
}

func NewWebhookHandler(flowService application.FlowService, requestService application.RequestService) *WebhookHandler {
	return &WebhookHandler{flowService: flowService, requestService: requestService}
}

func (h *WebhookHandler) TriggerFlow(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		response.Error(c, http.StatusBadRequest, "Webhook token is required")
		return
	}

	flow, err := h.flowService.GetFlowByWebhookToken(c.Request.Context(), token)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Invalid or expired webhook token")
		return
	}

	// We use the owner from the workspace
	requests, err := h.requestService.GetRequestsByFlow(c.Request.Context(), flow.WorkspaceID, flow.ID)
	if err != nil {
		// Try using a fallback query
		requests, _ = h.requestService.GetRequestsByFlow(c.Request.Context(), flow.WorkspaceID, flow.ID)
	}

	started := time.Now()
	results := make([]gin.H, 0, len(requests))
	passed := 0
	failed := 0
	accumulatedVars := make(map[string]string)

	for _, stored := range requests {
		res, runErr := h.requestService.ExecuteRequestWithVariables(c.Request.Context(), flow.WorkspaceID, stored.ID, nil, accumulatedVars)
		if runErr != nil || res.Status >= 400 || res.TestsFailed > 0 {
			failed++
			errMsg := ""
			if runErr != nil {
				errMsg = runErr.Error()
			} else if res.TestsFailed > 0 {
				errMsg = "assertions failed"
			} else {
				errMsg = res.StatusText
			}
			results = append(results, gin.H{
				"request_id": stored.ID,
				"name":       stored.Name,
				"status":     "failed",
				"error":      errMsg,
			})
			continue
		}
		passed++
		for k, v := range res.ExtractedVariables {
			accumulatedVars[k] = v
		}
		results = append(results, gin.H{
			"request_id": stored.ID,
			"name":       stored.Name,
			"status":     "passed",
			"code":       res.Status,
			"time_ms":    res.Time,
		})
	}

	overallStatus := "passed"
	statusCode := http.StatusOK
	if failed > 0 || (len(requests) > 0 && passed == 0) {
		overallStatus = "failed"
		statusCode = http.StatusExpectationFailed
	}

	c.JSON(statusCode, gin.H{
		"success":        overallStatus == "passed",
		"status":         overallStatus,
		"flow_id":        flow.ID,
		"flow_name":      flow.Name,
		"total_requests": len(requests),
		"passed":         passed,
		"failed":         failed,
		"duration_ms":    time.Since(started).Milliseconds(),
		"results":        results,
	})
}
