package response

import (
	"github.com/gin-gonic/gin"
)

// Success standardized JSON success response
func Success(c *gin.Context, statusCode int, message string, data any) {
	c.JSON(statusCode, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// Error standardized JSON error response
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   message,
	})
}
