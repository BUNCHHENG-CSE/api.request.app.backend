package controllers

import (
	"net/http"

	"backend/internal/models"
	"backend/internal/services"

	"github.com/gin-gonic/gin"
)

// RequestController handles request-related HTTP endpoints
type RequestController struct {
	service services.RequestService
}

func NewRequestController(service services.RequestService) *RequestController {
	return &RequestController{service: service}
}

// CreateRequest POST /api/requests
func (c *RequestController) CreateRequest(ctx *gin.Context) {
	var dto models.CreateRequestDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponseDTO{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	req, err := c.service.CreateRequest(ctx, &dto)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "CREATE_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusCreated, req)
}

// GetRequest GET /api/requests/:id
func (c *RequestController) GetRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	req, err := c.service.GetRequest(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.ErrorResponseDTO{
			Error:   "NOT_FOUND",
			Message: "Request not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	ctx.JSON(http.StatusOK, req)
}

// ListRequests GET /api/collections/:collectionId/requests
func (c *RequestController) ListRequests(ctx *gin.Context) {
	collectionID := ctx.Param("collectionId")

	requests, err := c.service.ListRequests(ctx, collectionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "LIST_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": requests})
}

// UpdateRequest PUT /api/requests/:id
func (c *RequestController) UpdateRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	var dto models.UpdateRequestDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponseDTO{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	req, err := c.service.UpdateRequest(ctx, id, &dto)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "UPDATE_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, req)
}

// DeleteRequest DELETE /api/requests/:id
func (c *RequestController) DeleteRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.service.DeleteRequest(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "DELETE_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Request deleted"})
}

// ExecuteRequest POST /api/requests/:id/execute
func (c *RequestController) ExecuteRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	var dto models.ExecuteRequestDTO
	dto.RequestID = id

	// Allow optional request body for overrides
	ctx.ShouldBindJSON(&dto)

	resp, err := c.service.ExecuteRequest(ctx, &dto)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "EXECUTION_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// DuplicateRequest POST /api/requests/:id/duplicate
func (c *RequestController) DuplicateRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	req, err := c.service.DuplicateRequest(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "DUPLICATE_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusCreated, req)
}

// SearchRequests GET /api/workspaces/:workspaceId/requests/search?q=query
func (c *RequestController) SearchRequests(ctx *gin.Context) {
	workspaceID := ctx.Param("workspaceId")
	query := ctx.Query("q")

	requests, err := c.service.SearchRequests(ctx, workspaceID, query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "SEARCH_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": requests})
}

// GetRequestHistory GET /api/requests/:id/history?limit=20
func (c *RequestController) GetRequestHistory(ctx *gin.Context) {
	id := ctx.Param("id")
	limit := 20
	ctx.BindQuery(&limit)

	history, err := c.service.GetRequestHistory(ctx, id, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "HISTORY_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": history})
}
