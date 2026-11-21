package handlers

import (
	"net/http"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BenchmarkHandler struct {
	service application.BenchmarkService
}

func NewBenchmarkHandler(service application.BenchmarkService) *BenchmarkHandler {
	return &BenchmarkHandler{service: service}
}

type RunBenchmarkRequest struct {
	EnvironmentID *uuid.UUID `json:"environment_id"`
	Iterations    int        `json:"iterations"`
	Concurrency   int        `json:"concurrency"`
	DelayMS       int        `json:"delay_ms"`
}

func (h *BenchmarkHandler) RunBenchmark(c *gin.Context) {
	flowID, ok := pathUUID(c, "id", "flow")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var req RunBenchmarkRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	}

	result, err := h.service.RunBenchmark(c.Request.Context(), ownerID, application.BenchmarkInput{
		FlowID:        flowID,
		EnvironmentID: req.EnvironmentID,
		Iterations:    req.Iterations,
		Concurrency:   req.Concurrency,
		DelayMS:       req.DelayMS,
	})
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Benchmark completed successfully", result)
}
