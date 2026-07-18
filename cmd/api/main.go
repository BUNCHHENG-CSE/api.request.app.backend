package main

import (
	"log"

	"api.request.app.backend/internal/domain"
	"api.request.app.backend/internal/infrastructure/config"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/infrastructure/repositories"
	"api.request.app.backend/internal/interfaces/api/handlers"
	"api.request.app.backend/internal/interfaces/api/routes"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 2. Load Configuration from .env
	cfg := config.LoadConfig()

	// Initialize Database using the configured URL
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 3. Run Database Migrations
	log.Println("Running database migrations...")
	err = db.AutoMigrate(
		&domain.User{},
		&domain.Workspace{},
		&domain.Flow{},
		&domain.Request{},
		&domain.Environment{},
		&domain.Collection{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database schemas:", err)
	}

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
