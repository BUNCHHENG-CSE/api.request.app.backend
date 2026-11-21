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
		handlers.NewWorkspaceMemberHandler(nil),
		handlers.NewOpenAPIHandler(nil),
		handlers.NewMonitorHandler(nil),
		handlers.NewCookieHandler(nil),
		handlers.NewBenchmarkHandler(nil),
		handlers.NewGraphQLHandler(nil),
		handlers.NewDocsHandler(nil),
		handlers.NewActivityHandler(nil),
		handlers.NewWebhookHandler(nil, nil),
		"test-secret",
		"http://localhost:3000",
	)

	routes := router.Routes()

	wanted := map[string]bool{
		"PATCH /api/v1/workspaces/:id":                          false,
		"POST /api/v1/requests/:id/execute":                     false,
		"DELETE /api/v1/collections/:id/requests/:requestId":    false,
		"POST /api/v1/requests/import-curl":                     false,
		"GET /api/v1/requests/:id/snippet":                      false,
		"POST /api/v1/collections/import-postman":               false,
		"GET /api/v1/local-collections/:workspaceId/export-zip":  false,
		"POST /api/v1/local-collections/:workspaceId/import-zip": false,
		"GET /api/v1/local-collections/:workspaceId/files":      false,
		"GET /api/v1/local-collections/:workspaceId/file":       false,
		"PUT /api/v1/local-collections/:workspaceId/file":       false,
		"GET /api/v1/mock/:workspaceId/*path":                   false,
		"GET /api/v1/workspaces/:id/members":                    false,
		"POST /api/v1/workspaces/:id/members":                   false,
		"PATCH /api/v1/workspaces/:id/members/:userId":          false,
		"DELETE /api/v1/workspaces/:id/members/:userId":         false,
		"GET /api/v1/workspaces/:id/export-openapi":             false,
		"POST /api/v1/collections/import-openapi":               false,
		"GET /api/v1/collections/:id/export-openapi":            false,
		"GET /api/v1/history/diff":                              false,
		"POST /api/v1/monitors":                                 false,
		"GET /api/v1/monitors":                                  false,
		"POST /api/v1/monitors/:id/run":                         false,
		"GET /api/v1/monitors/:id/logs":                         false,
		"POST /api/v1/webhooks/flows/:token":                    false,
		"GET /api/v1/workspaces/:id/cookies":                    false,
		"POST /api/v1/workspaces/:id/cookies":                   false,
		"GET /api/v1/workspaces/:id/activity":                   false,
		"GET /api/v1/workspaces/:id/docs":                       false,
		"POST /api/v1/flows/:id/benchmark":                      false,
		"POST /api/v1/flows/:id/webhook-token":                  false,
		"POST /api/v1/graphql/introspect":                       false,
		"GET /api/v1/collections/:id/docs":                      false,
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
