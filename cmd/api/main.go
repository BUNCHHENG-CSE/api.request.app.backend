package main

import (
	"log"

	"backend/internal/application"
	"backend/internal/infrastructure/repositories"
	"backend/internal/interfaces/api/handlers"
	"backend/internal/interfaces/api/routes"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Initialize Infrastructure (Database)
	// In production, load this from environment variables
	dsn := "host=localhost user=postgres password=secret dbname=mydb port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 2. Initialize Repositories (Infrastructure Layer)
	userRepo := repositories.NewUserRepository(db)
	workspaceRepo := repositories.NewWorkspaceRepository(db)
	flowRepo := repositories.NewFlowRepository(db)
	requestRepo := repositories.NewRequestRepository(db)
	envRepo := repositories.NewEnvironmentRepository(db)

	// 3. Initialize Services (Application Layer)
	userService := application.NewUserService(userRepo)
	workspaceService := application.NewWorkspaceService(workspaceRepo)
	flowService := application.NewFlowService(flowRepo)
	requestService := application.NewRequestService(requestRepo)
	envService := application.NewEnvironmentService(envRepo)

	// 4. Initialize Handlers (Interfaces Layer)
	userHandler := handlers.NewUserHandler(userService)
	workspaceHandler := handlers.NewWorkspaceHandler(workspaceService)
	flowHandler := handlers.NewFlowHandler(flowService)
	requestHandler := handlers.NewRequestHandler(requestService)
	envHandler := handlers.NewEnvironmentHandler(envService)

	// 5. Setup Router and Inject Handlers
	router := routes.SetupRouter(
		userHandler,
		workspaceHandler,
		flowHandler,
		requestHandler,
		envHandler,
	)

	// 6. Start the Server
	log.Println("Starting server on :8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
