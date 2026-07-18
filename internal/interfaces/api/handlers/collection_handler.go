package handlers

import (
	"net/http"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
)

type CollectionHandler struct {
	service application.CollectionService
}

func NewCollectionHandler(service application.CollectionService) *CollectionHandler {
	return &CollectionHandler{service: service}
}

func (h *CollectionHandler) Create(c *gin.Context) {
	var req CreateCollectionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	collection, err := h.service.CreateCollection(c.Request.Context(), req.WorkspaceID, req.Name, req.Description)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Collection created successfully !!!", collection)
}
