package handlers

import (
	"net/http"

	"backend/internal/application"

	"github.com/gin-gonic/gin"
)

type CollectionHandler struct {
	service application.CollectionService
}

func NewCollectionHandler(service application.CollectionService) *CollectionHandler {
	return &CollectionHandler{service: service}
}

func (h *CollectionHandler) Create(c *gin.Context) {
	var req struct {
		WorkspaceID uint   `json:"workspace_id" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection, err := h.service.CreateCollection(c.Request.Context(), req.WorkspaceID, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, collection)
}
