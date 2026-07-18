package handlers

import (
	"backend/internal/application"
	"net/http"

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
		FlowID  uint   `json:"flow_id" binding:"required"`
		Method  string `json:"method" binding:"required"`
		URL     string `json:"url" binding:"required"`
		Headers string `json:"headers"`
		Body    string `json:"body"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requestResult, err := h.service.CreateRequest(c.Request.Context(), req.FlowID, req.Method, req.URL, req.Headers, req.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, requestResult)
}
