package handlers

import (
	"net/http"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
)

type FlowHandler struct {
	service application.FlowService
}

func NewFlowHandler(service application.FlowService) *FlowHandler {
	return &FlowHandler{service: service}
}
func (h *FlowHandler) Create(c *gin.Context) {
	// Optimized: A Flow is usually created inside a Workspace, requiring a Name.
	var req CreateFlowRequest

	// 1. Bind and validate the JSON payload
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 2. Pass to the Application Service layer
	flow, err := h.service.CreateFlow(c.Request.Context(), req.WorkspaceID, req.Name, req.Description)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 3. Return standardized response
	response.Success(c, http.StatusCreated, "Flow created successfully !!!", flow)
}
