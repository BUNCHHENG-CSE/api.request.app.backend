package routes

import (
	"api.request.app.backend/internal/interfaces/api/handlers"
	"api.request.app.backend/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures the Gin engine with all required routes
func SetupRouter(
	userHandler *handlers.UserHandler,
	workspaceHandler *handlers.WorkspaceHandler,
	flowHandler *handlers.FlowHandler,
	requestHandler *handlers.RequestHandler,
	envHandler *handlers.EnvironmentHandler,
	collectionHandler *handlers.CollectionHandler,
	searchHandler *handlers.SearchHandler,
	localCollectionHandler *handlers.LocalCollectionHandler,
	importerHandler *handlers.ImporterHandler,
	snippetHandler *handlers.SnippetHandler,
	mockHandler *handlers.MockHandler,
	jwtSecret string,
	allowedOrigins string,
) *gin.Engine {

	router := gin.Default()
	_ = router.SetTrustedProxies(nil)
	router.Use(middleware.HTTPDefaults(allowedOrigins))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public Mock Server Route: matches simulated requests for any workspace
	router.Any("/api/v1/mock/:workspaceId/*path", mockHandler.ServeMock)

	v1 := router.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/register", userHandler.RegisterUser)
			users.POST("/login", userHandler.Login)
		}

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		{
			protected.GET("/search", searchHandler.Search)
			protected.GET("/users/:id", userHandler.GetUser)
			protected.PATCH("/users/:id", userHandler.Update)
			protected.POST("/users/refresh", userHandler.Refresh)
			protected.POST("/users/logout", userHandler.Logout)
			protected.POST("/users/change-password", userHandler.ChangePassword)

			workspaces := protected.Group("/workspaces")
			{
				workspaces.POST("", workspaceHandler.Create)
				workspaces.GET("", workspaceHandler.List)
				workspaces.GET("/:id", workspaceHandler.Get)
				workspaces.PATCH("/:id", workspaceHandler.Update)
				workspaces.DELETE("/:id", workspaceHandler.Delete)
				workspaces.GET("/:id/flows", flowHandler.ListByWorkspace)
				workspaces.GET("/:id/environments", envHandler.ListByWorkspace)
				workspaces.GET("/:id/collections", collectionHandler.ListByWorkspace)
				workspaces.POST("/:id/local/export", localCollectionHandler.Export)
			}

			flows := protected.Group("/flows")
			{
				flows.POST("", flowHandler.Create)
				flows.GET("/:id", flowHandler.Get)
				flows.PATCH("/:id", flowHandler.Update)
				flows.DELETE("/:id", flowHandler.Delete)
				flows.GET("/:id/requests", requestHandler.ListByFlow)
				flows.POST("/:id/run", flowHandler.Run)
			}

			requests := protected.Group("/requests")
			{
				requests.POST("", requestHandler.Create)
				requests.GET("/:id", requestHandler.Get)
				requests.PATCH("/:id", requestHandler.Update)
				requests.DELETE("/:id", requestHandler.Delete)
				requests.POST("/:id/execute", requestHandler.Execute)
				requests.POST("/:id/duplicate", requestHandler.Duplicate)
				requests.POST("/bulk-delete", requestHandler.BulkDelete)
				requests.POST("/import-curl", importerHandler.ImportCurl)
				requests.GET("/:id/snippet", snippetHandler.Generate)
			}

			history := protected.Group("/history")
			{
				history.GET("", requestHandler.ListHistory)
				history.GET("/:id", requestHandler.GetHistory)
				history.DELETE("", requestHandler.ClearHistory)
			}

			environments := protected.Group("/environments")
			{
				environments.POST("", envHandler.Create)
				environments.GET("/:id", envHandler.Get)
				environments.PATCH("/:id", envHandler.Update)
				environments.DELETE("/:id", envHandler.Delete)
			}

			collections := protected.Group("/collections")
			{
				collections.POST("", collectionHandler.Create)
				collections.GET("/:id", collectionHandler.Get)
				collections.PATCH("/:id", collectionHandler.Update)
				collections.DELETE("/:id", collectionHandler.Delete)
				collections.POST("/:id/requests/:requestId", collectionHandler.AddRequest)
				collections.DELETE("/:id/requests/:requestId", collectionHandler.RemoveRequest)
				collections.POST("/import-postman", importerHandler.ImportPostman)
			}

			localCollections := protected.Group("/local-collections")
			{
				localCollections.GET("", localCollectionHandler.List)
				localCollections.GET("/:workspaceId", localCollectionHandler.Get)
				localCollections.POST("/:workspaceId/import", localCollectionHandler.Import)
				localCollections.DELETE("/:workspaceId", localCollectionHandler.Delete)
				localCollections.GET("/:workspaceId/export-zip", localCollectionHandler.ExportZip)
				localCollections.POST("/:workspaceId/import-zip", localCollectionHandler.ImportZip)
				localCollections.GET("/:workspaceId/files", localCollectionHandler.GetFiles)
				localCollections.GET("/:workspaceId/file", localCollectionHandler.ReadFile)
				localCollections.PUT("/:workspaceId/file", localCollectionHandler.WriteFile)
			}
		}
	}

	return router
}
