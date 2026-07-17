package main

import (
	"log"
	"os"

	"backend/internal/middleware"
	// import repositories here once implemented
	// "backend/internal/repositories"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=flowapi password=password dbname=flowapi port=5432 sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database
	_, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// TODO: Run migrations
	// db.AutoMigrate(&entities.User{}, &entities.Workspace{}, &entities.Collection{}, ...)

	// TODO: Initialize repositories
	// requestRepo := repositories.NewRequestRepository(db)
	// workspaceRepo := repositories.NewWorkspaceRepository(db)
	// environmentRepo := repositories.NewEnvironmentRepository(db)
	// etc.

	// TODO: Initialize services
	// requestService := services.NewRequestService(requestRepo, environmentRepo, responseRepo, historyRepo)
	// workspaceService := services.NewWorkspaceService(workspaceRepo, userRepo)
	// environmentService := services.NewEnvironmentService(environmentRepo)
	// flowService := services.NewFlowService(flowRepo, requestService)
	// etc.

	// TODO: Initialize controllers
	// requestController := controllers.NewRequestController(requestService)
	// workspaceController := controllers.NewWorkspaceController(workspaceService)
	// environmentController := controllers.NewEnvironmentController(environmentService)
	// flowController := controllers.NewFlowController(flowService)

	// Setup Gin router
	r := gin.Default()

	// Apply middleware
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.ErrorHandlingMiddleware())

	// TODO: Register routes
	// routeConfig := &routes.RouteConfig{
	// 	RequestController:    requestController,
	// 	WorkspaceController:  workspaceController,
	// 	EnvironmentController: environmentController,
	// 	FlowController:       flowController,
	// }
	// routes.RegisterRoutes(r, routeConfig)

	// Start server
	log.Printf("Starting FlowAPI server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
