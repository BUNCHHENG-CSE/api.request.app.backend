package response

import (
	"github.com/gin-gonic/gin"
)

// BaseResponse Standard Response wrapper
type BaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Success formats a successful 2xx response
func Success(c *gin.Context, statusCode int, data any, message string) {
	c.JSON(statusCode, BaseResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error formats a 4xx or 5xx response
func Error(c *gin.Context, statusCode int, err string) {
	c.JSON(statusCode, BaseResponse{
		Success: false,
		Error:   err,
	})
}
