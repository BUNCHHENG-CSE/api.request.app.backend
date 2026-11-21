package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api.request.app.backend/internal/domain"
	"api.request.app.backend/internal/infrastructure/config"
	"api.request.app.backend/internal/infrastructure/database"

	"api.request.app.backend/internal/application"
	"api.request.app.backend/internal/infrastructure/repositories"
	"api.request.app.backend/internal/interfaces/api/handlers"
	"api.request.app.backend/internal/interfaces/api/routes"
)

func main() {
	// 2. Load Configuration from .env
	cfg := config.LoadConfig()

	// Initialize Database using the configured URL
	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
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
		&domain.RequestHistory{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database schemas:", err)
	}

	userRepo := repositories.NewUserRepository(db)
	workspaceRepo := repositories.NewWorkspaceRepository(db)
	flowRepo := repositories.NewFlowRepository(db)
	requestRepo := repositories.NewRequestRepository(db)
	envRepo := repositories.NewEnvironmentRepository(db)
	collectionRepo := repositories.NewCollectionRepository(db)
	historyRepo := repositories.NewRequestHistoryRepository(db)
	searchRepo := repositories.NewSearchRepository(db)

	// 3. Initialize Services (Application Layer)
	userService := application.NewUserService(userRepo, cfg.JWTSecret)
	workspaceService := application.NewWorkspaceService(workspaceRepo)
	flowService := application.NewFlowService(flowRepo, workspaceRepo)
	requestService := application.NewRequestService(requestRepo, flowRepo, workspaceRepo, collectionRepo, envRepo, historyRepo, cfg.AllowPrivateRequests)
	envService := application.NewEnvironmentService(envRepo, workspaceRepo)
	collectionService := application.NewCollectionService(collectionRepo, workspaceRepo)
	searchService := application.NewSearchService(searchRepo)
	localCollectionService, err := application.NewLocalCollectionService(cfg.LocalCollectionsRoot, workspaceRepo, flowRepo, requestRepo, envRepo, collectionRepo)
	if err != nil {
		log.Fatal("Failed to initialize local collection storage:", err)
	}
	importerService := application.NewImporterService(requestService, flowService, workspaceRepo, flowRepo, requestRepo, collectionRepo)
	snippetService := application.NewSnippetService(requestService)

	// 4. Initialize Handlers (Interfaces Layer)
	userHandler := handlers.NewUserHandler(userService)
	workspaceHandler := handlers.NewWorkspaceHandler(workspaceService)
	flowHandler := handlers.NewFlowHandler(flowService, requestService)
	requestHandler := handlers.NewRequestHandler(requestService)
	envHandler := handlers.NewEnvironmentHandler(envService)
	collectionHandler := handlers.NewCollectionHandler(collectionService, requestService)
	searchHandler := handlers.NewSearchHandler(searchService)
	localCollectionHandler := handlers.NewLocalCollectionHandler(localCollectionService)
	importerHandler := handlers.NewImporterHandler(importerService)
	snippetHandler := handlers.NewSnippetHandler(snippetService)
	mockHandler := handlers.NewMockHandler(workspaceRepo, flowRepo, requestRepo)

	// 5. Setup Router and Inject Handlers
	router := routes.SetupRouter(
		userHandler,
		workspaceHandler,
		flowHandler,
		requestHandler,
		envHandler,
		collectionHandler,
		searchHandler,
		localCollectionHandler,
		importerHandler,
		snippetHandler,
		mockHandler,
		cfg.JWTSecret,
		cfg.AllowedOrigins,
	)

	address := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Starting server on %s...", address)
	server := &http.Server{
		Addr: address, Handler: router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      35 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}
