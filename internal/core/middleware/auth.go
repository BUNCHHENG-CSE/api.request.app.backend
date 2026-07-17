package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates API keys or JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing or invalid token"})
			return
		}

		// Token validation logic would go here
		// token := strings.TrimPrefix(authHeader, "Bearer ")

		// Set user ID in context for downstream handlers to use
		c.Set("userID", "example-user-id")

		c.Next()
	}
}
