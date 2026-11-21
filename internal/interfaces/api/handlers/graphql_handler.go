package handlers

import (
	"net/http"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
)

type GraphQLHandler struct {
	service application.GraphQLService
}

func NewGraphQLHandler(service application.GraphQLService) *GraphQLHandler {
	return &GraphQLHandler{service: service}
}

type IntrospectRequest struct {
	URL     string            `json:"url" binding:"required"`
	Headers map[string]string `json:"headers"`
}

func (h *GraphQLHandler) Introspect(c *gin.Context) {
	var req IntrospectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	schema, err := h.service.Introspect(c.Request.Context(), req.URL, req.Headers)
	if err != nil {
		serviceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "GraphQL schema introspected successfully", schema)
}
