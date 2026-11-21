package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"api.request.app.backend/internal/application"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestServiceErrorStatusMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "not found error",
			err:        errors.New("workspace not found"),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "gorm record not found",
			err:        gorm.ErrRecordNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "user not found sentinel",
			err:        application.ErrUserNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "duplicate email conflict",
			err:        application.ErrEmailAlreadyInUse,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "invalid credentials unauthorized",
			err:        application.ErrInvalidCredentials,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid current password unauthorized",
			err:        application.ErrInvalidCurrentPassword,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "validation required bad request",
			err:        errors.New("flow name is required"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "validation empty bad request",
			err:        errors.New("flow name cannot be empty"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "forbidden ownership mismatch",
			err:        errors.New("ownership mismatch"),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "bad gateway connection refused",
			err:        errors.New("dial tcp 127.0.0.1:8080: connect: connection refused"),
			wantStatus: http.StatusBadGateway,
		},
		{
			name:       "internal server error fallback",
			err:        errors.New("unexpected database driver disk error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			serviceError(c, tt.err)

			if w.Code != tt.wantStatus {
				t.Fatalf("for error %q: got status %d, want %d", tt.err, w.Code, tt.wantStatus)
			}
		})
	}
}
