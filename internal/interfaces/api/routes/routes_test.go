package routes

import (
	"testing"

	"api.request.app.backend/internal/interfaces/api/handlers"
	"github.com/gin-gonic/gin"
)

func TestSetupRouterRegistersAllRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := SetupRouter(
		handlers.NewUserHandler(nil),
		handlers.NewWorkspaceHandler(nil),
		handlers.NewFlowHandler(nil, nil),
		handlers.NewRequestHandler(nil),
		handlers.NewEnvironmentHandler(nil),
		handlers.NewCollectionHandler(nil, nil),
		handlers.NewSearchHandler(nil),
		handlers.NewLocalCollectionHandler(nil),
		handlers.NewImporterHandler(nil),
		handlers.NewSnippetHandler(nil),
		handlers.NewMockHandler(nil, nil, nil),
		"test-secret",
		"http://localhost:3000",
	)

	routes := router.Routes()

	wanted := map[string]bool{
		"PATCH /api/v1/workspaces/:id":                         false,
		"POST /api/v1/requests/:id/execute":                    false,
		"DELETE /api/v1/collections/:id/requests/:requestId":   false,
		"POST /api/v1/requests/import-curl":                    false,
		"GET /api/v1/requests/:id/snippet":                     false,
		"POST /api/v1/collections/import-postman":              false,
		"GET /api/v1/local-collections/:workspaceId/export-zip": false,
		"POST /api/v1/local-collections/:workspaceId/import-zip": false,
		"GET /api/v1/local-collections/:workspaceId/files":     false,
		"GET /api/v1/local-collections/:workspaceId/file":      false,
		"PUT /api/v1/local-collections/:workspaceId/file":      false,
		"GET /api/v1/mock/:workspaceId/*path":                  false,
	}
	for _, route := range routes {
		key := route.Method + " " + route.Path
		if _, ok := wanted[key]; ok {
			wanted[key] = true
		}
	}
	for route, found := range wanted {
		if !found {
			t.Errorf("route %s was not registered", route)
		}
	}
}
