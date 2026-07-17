package controllers

import (
	"net/http"

	"backend/internal/models"
	"backend/internal/services"

	"github.com/gin-gonic/gin"
)

// FlowController handles flow-related HTTP endpoints
type FlowController struct {
	service services.FlowService
}

func NewFlowController(service services.FlowService) *FlowController {
	return &FlowController{service: service}
}

// CreateFlow POST /api/workspaces/:workspaceId/flows
func (c *FlowController) CreateFlow(ctx *gin.Context) {
	workspaceID := ctx.Param("workspaceId")

	var dto models.CreateFlowDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponseDTO{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	flow, err := c.service.CreateFlow(ctx, workspaceID, &dto)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "CREATE_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusCreated, flow)
}

// GetFlow GET /api/flows/:id
func (c *FlowController) GetFlow(ctx *gin.Context) {
	id := ctx.Param("id")

	flow, err := c.service.GetFlow(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.ErrorResponseDTO{
			Error:   "NOT_FOUND",
			Message: "Flow not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	ctx.JSON(http.StatusOK, flow)
}

// ListFlows GET /api/workspaces/:workspaceId/flows
func (c *FlowController) ListFlows(ctx *gin.Context) {
	workspaceID := ctx.Param("workspaceId")

	flows, err := c.service.ListFlows(ctx, workspaceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "LIST_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": flows})
}

// UpdateFlow PUT /api/flows/:id
func (c *FlowController) UpdateFlow(ctx *gin.Context) {
	id := ctx.Param("id")

	var dto models.UpdateFlowDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponseDTO{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	flow, err := c.service.UpdateFlow(ctx, id, &dto)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "UPDATE_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, flow)
}

// DeleteFlow DELETE /api/flows/:id
func (c *FlowController) DeleteFlow(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.service.DeleteFlow(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "DELETE_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Flow deleted"})
}

// ExecuteFlow POST /api/flows/:id/execute
func (c *FlowController) ExecuteFlow(ctx *gin.Context) {
	id := ctx.Param("id")

	var executionCtx map[string]interface{}
	ctx.ShouldBindJSON(&executionCtx)

	results, err := c.service.ExecuteFlow(ctx, id, executionCtx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "EXECUTION_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"results": results})
}
