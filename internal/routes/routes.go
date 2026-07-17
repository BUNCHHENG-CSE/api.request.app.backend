package routes

import (
	"backend/internal/controllers"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

type RouteConfig struct {
	RequestController     *controllers.RequestController
	WorkspaceController   *controllers.WorkspaceController
	EnvironmentController *controllers.EnvironmentController
	FlowController        *controllers.FlowController
}

// RegisterRoutes sets up all API routes
func RegisterRoutes(r *gin.Engine, cfg *RouteConfig) {
	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")

	// Public routes (no auth required)
	_ = api.Group("")
	{
		// Auth routes would go here (login, register, refresh token)
		// For brevity, they're omitted in this example
	}

	// Protected routes (auth required)
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// Workspaces
		protected.POST("/workspaces", cfg.WorkspaceController.CreateWorkspace)
		protected.GET("/workspaces", cfg.WorkspaceController.ListWorkspaces)
		protected.GET("/workspaces/:id", cfg.WorkspaceController.GetWorkspace)
		protected.PUT("/workspaces/:id", cfg.WorkspaceController.UpdateWorkspace)
		protected.DELETE("/workspaces/:id", cfg.WorkspaceController.DeleteWorkspace)

		// Workspace Members
		protected.POST("/workspaces/:id/members", cfg.WorkspaceController.AddMember)
		protected.GET("/workspaces/:id/members", cfg.WorkspaceController.GetMembers)
		protected.DELETE("/workspaces/:id/members/:userId", cfg.WorkspaceController.RemoveMember)
		protected.PUT("/workspaces/:id/members/:userId/role", cfg.WorkspaceController.UpdateMemberRole)

		// Environments
		protected.POST("/workspaces/:workspaceId/environments", cfg.EnvironmentController.CreateEnvironment)
		protected.GET("/workspaces/:workspaceId/environments", cfg.EnvironmentController.ListEnvironments)
		protected.GET("/workspaces/:workspaceId/environments/active", cfg.EnvironmentController.GetActiveEnvironment)
		protected.GET("/environments/:id", cfg.EnvironmentController.GetEnvironment)
		protected.PUT("/environments/:id", cfg.EnvironmentController.UpdateEnvironment)
		protected.DELETE("/environments/:id", cfg.EnvironmentController.DeleteEnvironment)
		protected.POST("/workspaces/:workspaceId/environments/:id/activate", cfg.EnvironmentController.SetActiveEnvironment)

		// Requests
		protected.POST("/requests", cfg.RequestController.CreateRequest)
		protected.GET("/requests/:id", cfg.RequestController.GetRequest)
		protected.PUT("/requests/:id", cfg.RequestController.UpdateRequest)
		protected.DELETE("/requests/:id", cfg.RequestController.DeleteRequest)
		protected.POST("/requests/:id/execute", cfg.RequestController.ExecuteRequest)
		protected.POST("/requests/:id/duplicate", cfg.RequestController.DuplicateRequest)
		protected.GET("/requests/:id/history", cfg.RequestController.GetRequestHistory)
		protected.GET("/collections/:collectionId/requests", cfg.RequestController.ListRequests)
		protected.GET("/workspaces/:workspaceId/requests/search", cfg.RequestController.SearchRequests)

		// Flows
		protected.POST("/workspaces/:workspaceId/flows", cfg.FlowController.CreateFlow)
		protected.GET("/workspaces/:workspaceId/flows", cfg.FlowController.ListFlows)
		protected.GET("/flows/:id", cfg.FlowController.GetFlow)
		protected.PUT("/flows/:id", cfg.FlowController.UpdateFlow)
		protected.DELETE("/flows/:id", cfg.FlowController.DeleteFlow)
		protected.POST("/flows/:id/execute", cfg.FlowController.ExecuteFlow)
	}
}
