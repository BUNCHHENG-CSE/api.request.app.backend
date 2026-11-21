package handlers

import (
	"fmt"
	"io"
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
		serviceError(c, err)
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
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Workspace synchronized to local Bruno collection", item)
}

func (h *LocalCollectionHandler) ExportZip(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "workspaceId", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	zipBytes, filename, err := h.service.ExportZip(c.Request.Context(), ownerID, workspaceID)
	if err != nil {
		serviceError(c, err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Data(http.StatusOK, "application/zip", zipBytes)
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
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Local Bruno collection imported successfully", counts)
}

func (h *LocalCollectionHandler) ImportZip(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "workspaceId", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "file upload is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to read uploaded zip file")
		return
	}

	counts, err := h.service.ImportZip(c.Request.Context(), ownerID, workspaceID, data)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Bruno zip collection imported successfully", counts)
}

func (h *LocalCollectionHandler) GetFiles(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "workspaceId", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	entries, err := h.service.GetFileTree(c.Request.Context(), ownerID, workspaceID)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Local collection files retrieved successfully", entries)
}

func (h *LocalCollectionHandler) ReadFile(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "workspaceId", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	relPath := c.Query("path")
	if relPath == "" {
		response.Error(c, http.StatusBadRequest, "path query parameter is required")
		return
	}
	content, err := h.service.ReadLocalFile(c.Request.Context(), ownerID, workspaceID, relPath)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Local file read successfully", gin.H{
		"path":    relPath,
		"content": content,
	})
}

func (h *LocalCollectionHandler) WriteFile(c *gin.Context) {
	workspaceID, ok := pathUUID(c, "workspaceId", "workspace")
	if !ok {
		return
	}
	ownerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	var req struct {
		Path    string `json:"path" binding:"required"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	err := h.service.WriteLocalFile(c.Request.Context(), ownerID, workspaceID, req.Path, req.Content)
	if err != nil {
		serviceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Local file saved successfully", gin.H{
		"path": req.Path,
	})
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
