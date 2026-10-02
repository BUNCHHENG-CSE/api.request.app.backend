package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"
	userID := uuid.New()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{name: "valid", authHeader: "Bearer " + token, wantStatus: http.StatusNoContent},
		{name: "missing", wantStatus: http.StatusUnauthorized},
		{name: "invalid", authHeader: "Bearer invalid", wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(AuthMiddleware(secret))
			router.GET("/protected", func(c *gin.Context) {
				if tt.name == "valid" && c.MustGet("user_id") != userID {
					t.Fatalf("unexpected user ID in context")
				}
				c.Status(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			if res.Code != tt.wantStatus {
				t.Fatalf("got status %d, want %d", res.Code, tt.wantStatus)
			}
		})
	}
}
