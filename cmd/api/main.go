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
		&domain.WorkspaceMember{},
		&domain.Flow{},
		&domain.Request{},
		&domain.Environment{},
		&domain.Collection{},
		&domain.RequestHistory{},
		&domain.Monitor{},
		&domain.MonitorRunLog{},
		&domain.WorkspaceCookie{},
		&domain.WorkspaceActivity{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database schemas:", err)
	}

	userRepo := repositories.NewUserRepository(db)
	workspaceRepo := repositories.NewWorkspaceRepository(db)
	workspaceMemberRepo := repositories.NewWorkspaceMemberRepository(db)
	flowRepo := repositories.NewFlowRepository(db)
	requestRepo := repositories.NewRequestRepository(db)
	envRepo := repositories.NewEnvironmentRepository(db)
	collectionRepo := repositories.NewCollectionRepository(db)
	historyRepo := repositories.NewRequestHistoryRepository(db)
	searchRepo := repositories.NewSearchRepository(db)
	monitorRepo := repositories.NewMonitorRepository(db)
	cookieRepo := repositories.NewCookieRepository(db)
	activityRepo := repositories.NewActivityRepository(db)

	// 3. Initialize Services (Application Layer)
	userService := application.NewUserService(userRepo, cfg.JWTSecret)
	workspaceService := application.NewWorkspaceService(workspaceRepo)
	workspaceMemberService := application.NewWorkspaceMemberService(workspaceRepo, workspaceMemberRepo, userRepo)
	flowService := application.NewFlowService(flowRepo, workspaceRepo)
	requestService := application.NewRequestService(requestRepo, flowRepo, workspaceRepo, collectionRepo, envRepo, historyRepo, cfg.AllowPrivateRequests)
	requestService.SetCookieJar(cookieRepo)
	envService := application.NewEnvironmentService(envRepo, workspaceRepo)
	collectionService := application.NewCollectionService(collectionRepo, workspaceRepo)
	searchService := application.NewSearchService(searchRepo)
	localCollectionService, err := application.NewLocalCollectionService(cfg.LocalCollectionsRoot, workspaceRepo, flowRepo, requestRepo, envRepo, collectionRepo)
	if err != nil {
		log.Fatal("Failed to initialize local collection storage:", err)
	}
	importerService := application.NewImporterService(requestService, flowService, workspaceRepo, flowRepo, requestRepo, collectionRepo)
	snippetService := application.NewSnippetService(requestService)
	openAPIService := application.NewOpenAPIService(workspaceRepo, collectionRepo, flowRepo, requestRepo)
	monitorService := application.NewMonitorService(monitorRepo, workspaceRepo, flowRepo, requestService)
	cookieService := application.NewCookieService(cookieRepo, workspaceRepo)
	benchmarkService := application.NewBenchmarkService(flowRepo, workspaceRepo, requestService)
	graphQLService := application.NewGraphQLService()
	docsService := application.NewDocsService(workspaceRepo, collectionRepo, flowRepo, requestRepo)
	activityService := application.NewActivityService(activityRepo, workspaceRepo)

	// 4. Initialize Handlers (Interfaces Layer)
	userHandler := handlers.NewUserHandler(userService)
	workspaceHandler := handlers.NewWorkspaceHandler(workspaceService)
	workspaceMemberHandler := handlers.NewWorkspaceMemberHandler(workspaceMemberService)
	flowHandler := handlers.NewFlowHandler(flowService, requestService)
	requestHandler := handlers.NewRequestHandler(requestService)
	envHandler := handlers.NewEnvironmentHandler(envService)
	collectionHandler := handlers.NewCollectionHandler(collectionService, requestService)
	searchHandler := handlers.NewSearchHandler(searchService)
	localCollectionHandler := handlers.NewLocalCollectionHandler(localCollectionService)
	importerHandler := handlers.NewImporterHandler(importerService)
	snippetHandler := handlers.NewSnippetHandler(snippetService)
	mockHandler := handlers.NewMockHandler(workspaceRepo, flowRepo, requestRepo)
	openAPIHandler := handlers.NewOpenAPIHandler(openAPIService)
	monitorHandler := handlers.NewMonitorHandler(monitorService)
	cookieHandler := handlers.NewCookieHandler(cookieService)
	benchmarkHandler := handlers.NewBenchmarkHandler(benchmarkService)
	graphQLHandler := handlers.NewGraphQLHandler(graphQLService)
	docsHandler := handlers.NewDocsHandler(docsService)
	activityHandler := handlers.NewActivityHandler(activityService)
	webhookHandler := handlers.NewWebhookHandler(flowService, requestService)

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
		workspaceMemberHandler,
		openAPIHandler,
		monitorHandler,
		cookieHandler,
		benchmarkHandler,
		graphQLHandler,
		docsHandler,
		activityHandler,
		webhookHandler,
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
