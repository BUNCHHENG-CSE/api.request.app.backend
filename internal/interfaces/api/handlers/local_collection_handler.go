package handlers

import (
	"net/http"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
)

type LocalCollectionHandler struct {
	service application.LocalCollectionService
}

func NewLocalCollectionHandler(service application.LocalCollectionService) *LocalCollectionHandler {
	return &LocalCollectionHandler{service: service}
}

func (h *LocalCollectionHandler) List(c *gin.Context) {
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	items, err := h.service.List(c.Request.Context(), ownerID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Local collections retrieved successfully", items)
}

func (h *LocalCollectionHandler) Get(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "workspaceId", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	item, err := h.service.Get(c.Request.Context(), ownerID, workspaceID)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Local collection retrieved successfully", item)
}

func (h *LocalCollectionHandler) Export(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	item, err := h.service.Export(c.Request.Context(), ownerID, workspaceID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Workspace synchronized to local Bruno collection", item)
}

func (h *LocalCollectionHandler) Import(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "workspaceId", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	counts, err := h.service.Import(c.Request.Context(), ownerID, workspaceID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Local Bruno collection imported successfully", counts)
}

func (h *LocalCollectionHandler) Delete(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "workspaceId", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), ownerID, workspaceID); err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Local collection deleted successfully", gin.H{})
}
