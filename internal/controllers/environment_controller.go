package controllers

import (
	"net/http"

	"backend/internal/models"
	"backend/internal/services"

	"github.com/gin-gonic/gin"
)

// EnvironmentController handles environment-related HTTP endpoints
type EnvironmentController struct {
	service services.EnvironmentService
}

func NewEnvironmentController(service services.EnvironmentService) *EnvironmentController {
	return &EnvironmentController{service: service}
}

// CreateEnvironment POST /api/workspaces/:workspaceId/environments
func (c *EnvironmentController) CreateEnvironment(ctx *gin.Context) {
	workspaceID := ctx.Param("workspaceId")

	var dto models.CreateEnvironmentDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponseDTO{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	env, err := c.service.CreateEnvironment(ctx, workspaceID, &dto)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "CREATE_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusCreated, env)
}

// GetEnvironment GET /api/environments/:id
func (c *EnvironmentController) GetEnvironment(ctx *gin.Context) {
	id := ctx.Param("id")

	env, err := c.service.GetEnvironment(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.ErrorResponseDTO{
			Error:   "NOT_FOUND",
			Message: "Environment not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	ctx.JSON(http.StatusOK, env)
}

// ListEnvironments GET /api/workspaces/:workspaceId/environments
func (c *EnvironmentController) ListEnvironments(ctx *gin.Context) {
	workspaceID := ctx.Param("workspaceId")

	envs, err := c.service.ListEnvironments(ctx, workspaceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "LIST_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": envs})
}

// UpdateEnvironment PUT /api/environments/:id
func (c *EnvironmentController) UpdateEnvironment(ctx *gin.Context) {
	id := ctx.Param("id")

	var dto models.UpdateEnvironmentDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponseDTO{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	env, err := c.service.UpdateEnvironment(ctx, id, &dto)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "UPDATE_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, env)
}

// DeleteEnvironment DELETE /api/environments/:id
func (c *EnvironmentController) DeleteEnvironment(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.service.DeleteEnvironment(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "DELETE_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Environment deleted"})
}

// SetActiveEnvironment POST /api/workspaces/:workspaceId/environments/:id/activate
func (c *EnvironmentController) SetActiveEnvironment(ctx *gin.Context) {
	workspaceID := ctx.Param("workspaceId")
	id := ctx.Param("id")

	if err := c.service.SetActiveEnvironment(ctx, workspaceID, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "ACTIVATE_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Environment activated"})
}

// GetActiveEnvironment GET /api/workspaces/:workspaceId/environments/active
func (c *EnvironmentController) GetActiveEnvironment(ctx *gin.Context) {
	workspaceID := ctx.Param("workspaceId")

	env, err := c.service.GetActiveEnvironment(ctx, workspaceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
			Error:   "GET_ERROR",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.JSON(http.StatusOK, env)
}
