package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseDTO{
				Error:   "UNAUTHORIZED",
				Message: "Missing authorization header",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseDTO{
				Error:   "UNAUTHORIZED",
				Message: "Invalid authorization header format",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		token := parts[1]
		claims := &jwt.MapClaims{}

		// Parse token (you'll need to implement actual JWT parsing with your secret)
		_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("your-secret-key"), nil // Replace with actual secret
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseDTO{
				Error:   "UNAUTHORIZED",
				Message: "Invalid token",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Extract user ID from claims
		if userID, ok := (*claims)["sub"].(string); ok {
			c.Set("user_id", userID)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseDTO{
				Error:   "UNAUTHORIZED",
				Message: "Invalid token claims",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
		}
	}
}

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// ErrorHandlingMiddleware handles panics and errors
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponseDTO{
					Error:   "INTERNAL_ERROR",
					Message: fmt.Sprintf("%v", err),
					Code:    http.StatusInternalServerError,
				})
			}
		}()
		c.Next()
	}
}
