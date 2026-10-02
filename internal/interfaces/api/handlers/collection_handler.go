package handlers

import (
	"net/http"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
)

type CollectionHandler struct {
	service        application.CollectionService
	requestService application.RequestService
}

func NewCollectionHandler(service application.CollectionService, requestService application.RequestService) *CollectionHandler {
	return &CollectionHandler{service: service, requestService: requestService}
}

func (h *CollectionHandler) Create(c *gin.Context) {
	var req CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	collection, err := h.service.CreateCollection(c.Request.Context(), ownerID, req.WorkspaceID, req.Name, req.Description)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Collection created successfully", collection)
}

func (h *CollectionHandler) ListByWorkspace(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "id", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	collections, err := h.service.GetCollectionsByWorkspace(c.Request.Context(), ownerID, workspaceID)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Collections retrieved successfully", collections)
}

func (h *CollectionHandler) Get(c *gin.Context) {
	id, ok := pathUUID(c, "id", "collection")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	collection, err := h.service.GetCollection(c.Request.Context(), ownerID, id)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Collection retrieved successfully", collection)
}

func (h *CollectionHandler) Update(c *gin.Context) {
	id, ok := pathUUID(c, "id", "collection")
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
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	collection, err := h.service.UpdateCollection(c.Request.Context(), ownerID, id, req.Name, req.Description)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Collection updated successfully", collection)
}

func (h *CollectionHandler) Delete(c *gin.Context) {
	id, ok := pathUUID(c, "id", "collection")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteCollection(c.Request.Context(), ownerID, id); err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Collection deleted successfully", gin.H{})
}

func (h *CollectionHandler) AddRequest(c *gin.Context) {
	collectionID, ok := pathUUID(c, "id", "collection")
	if !ok {
		return
	}
	requestID, ok := pathUUID(c, "requestId", "request")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	req, err := h.requestService.SetCollection(c.Request.Context(), ownerID, requestID, &collectionID)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Request associated with collection", req)
}

func (h *CollectionHandler) RemoveRequest(c *gin.Context) {
	collectionID, ok := pathUUID(c, "id", "collection")
	if !ok {
		return
	}
	requestID, ok := pathUUID(c, "requestId", "request")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	existing, err := h.requestService.GetRequest(c.Request.Context(), ownerID, requestID)
	if err != nil {
		serviceError(c, err)
		return
	}
	if existing.CollectionID == nil || *existing.CollectionID != collectionID {
		response.Error(c, http.StatusNotFound, "request is not associated with this collection")
		return
	}
	req, err := h.requestService.SetCollection(c.Request.Context(), ownerID, requestID, nil)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Request removed from collection", req)
}
