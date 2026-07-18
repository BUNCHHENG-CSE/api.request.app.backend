package handlers

import (
	"net/http"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type RequestHandler struct {
	service application.RequestService
}

func NewRequestHandler(service application.RequestService) *RequestHandler {
	return &RequestHandler{service: service}
}

func (h *RequestHandler) Create(c *gin.Context) {
	var req struct {
		FlowID  uuid.UUID `json:"flow_id" binding:"required"`
		Method  string    `json:"method" binding:"required"`
		URL     string    `json:"url" binding:"required"`
		Headers string    `json:"headers"`
		Body    string    `json:"body"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	requestResult, err := h.service.CreateRequest(c.Request.Context(), req.FlowID, req.Method, req.URL, req.Headers, req.Body)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Request result created successfully !!!", requestResult)
}
