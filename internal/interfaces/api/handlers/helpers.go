package handlers

import (
	"net/http"

	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func authenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	value, ok := c.Get("user_id")
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authenticated user is required")
		return uuid.Nil, false
	}
	id, ok := value.(uuid.UUID)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "invalid authenticated user")
		return uuid.Nil, false
	}
	return id, true
}

func pathUUID(c *gin.Context, name, label string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid "+label+" ID")
		return uuid.Nil, false
	}
	return id, true
}

func serviceError(c *gin.Context, err error) {
	response.Error(c, http.StatusNotFound, err.Error())
}
