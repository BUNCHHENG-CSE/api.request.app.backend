package routes

import (
	"api.request.app.backend/internal/interfaces/api/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures the Gin engine with all required routes
func SetupRouter(
	userHandler *handlers.UserHandler,
	workspaceHandler *handlers.WorkspaceHandler,
	flowHandler *handlers.FlowHandler,
	requestHandler *handlers.RequestHandler,
	envHandler *handlers.EnvironmentHandler,
) *gin.Engine {

	router := gin.Default()

	// API Version 1 Group
	v1 := router.Group("/api/v1")
	{
		// User routes
		users := v1.Group("/users")
		{
			users.POST("/register", userHandler.RegisterUser)
			users.GET("/:id", userHandler.GetUser)
		}

		// Workspace routes (Continuing from the previous step)
		workspaces := v1.Group("/workspaces")
		{
			workspaces.POST("/", workspaceHandler.Create)
			workspaces.GET("/:id", workspaceHandler.Get)
		}
		// Flow routes
		flows := v1.Group("/flows")
		{
			flows.POST("/", flowHandler.Create)
			// Add GET, PUT, DELETE here as you implement them in the handler
		}

		// Request routes
		requests := v1.Group("/requests")
		{
			requests.POST("/", requestHandler.Create)
			// Add GET, PUT, DELETE here as you implement them in the handler
		}

		// Environment routes
		environments := v1.Group("/environments")
		{
			environments.POST("/", envHandler.Create)
			// Add GET, PUT, DELETE here as you implement them in the handler
		}
	}

	return router
}
