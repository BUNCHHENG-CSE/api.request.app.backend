package handlers

import (
	"errors"
	"net/http"
	"strings"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/interfaces/api/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
	if err == nil {
		return
	}

	// Check sentinel errors first
	if errors.Is(err, application.ErrEmailAlreadyInUse) {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	if errors.Is(err, application.ErrInvalidCurrentPassword) || errors.Is(err, application.ErrInvalidCredentials) {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	if errors.Is(err, application.ErrUserNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	msg := strings.ToLower(err.Error())

	// 404 Not Found patterns
	if strings.Contains(msg, "not found") || strings.Contains(msg, "record not found") {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	// 409 Conflict patterns
	if strings.Contains(msg, "already in use") || strings.Contains(msg, "already exists") || strings.Contains(msg, "duplicate key") {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}

	// 403 Forbidden patterns
	if strings.Contains(msg, "only access your own") || strings.Contains(msg, "only update your own") ||
		strings.Contains(msg, "ownership mismatch") || strings.Contains(msg, "belongs to another workspace") ||
		strings.Contains(msg, "escapes collection root") {
		response.Error(c, http.StatusForbidden, err.Error())
		return
	}

	// 400 Bad Request patterns (validation, malformed inputs, constraints)
	if strings.Contains(msg, "required") || strings.Contains(msg, "cannot be empty") ||
		strings.Contains(msg, "cannot be negative") || strings.Contains(msg, "must be") ||
		strings.Contains(msg, "must belong to") || strings.Contains(msg, "invalid") ||
		strings.Contains(msg, "at least") || strings.Contains(msg, "must contain") ||
		strings.Contains(msg, "confirmation do not match") || strings.Contains(msg, "blocked private") {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 502 Bad Gateway for network/dial failures
	if strings.Contains(msg, "connection refused") || strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "no such host") || strings.Contains(msg, "context deadline exceeded") {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}

	// Default internal server error
	response.Error(c, http.StatusInternalServerError, err.Error())
}
