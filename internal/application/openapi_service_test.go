package application

import (
	"context"
	"encoding/json"
	"testing"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIService_ImportAndExport(t *testing.T) {
	ownerID := uuid.New()
	wsID := uuid.New()

	wsRepo := &workspaceRepoMemory{workspace: &domain.Workspace{ID: wsID, OwnerID: ownerID, Name: "Test API"}}
	colRepo := &collectionRepoMemory{}
	flowRepo := &flowRepoMemory{}
	reqRepo := newRequestRepoMemory()

	svc := NewOpenAPIService(wsRepo, colRepo, flowRepo, reqRepo)

	openAPISpecYAML := `
openapi: 3.0.0
info:
  title: Petstore API
  description: Sample Petstore OpenAPI spec
  version: 1.0.0
servers:
  - url: https://petstore.example.com/v1
paths:
  /pets:
    get:
      summary: List all pets
      parameters:
        - name: limit
          in: query
          required: false
      responses:
        '200':
          description: A paged array of pets
    post:
      summary: Create a pet
      requestBody:
        content:
          application/json:
            example:
              name: Fluffy
              tag: dog
      responses:
        '201':
          description: Null response
`

	t.Run("Import OpenAPI 3.0 YAML", func(t *testing.T) {
		res, err := svc.ImportOpenAPI(context.Background(), ownerID, wsID, []byte(openAPISpecYAML))
		require.NoError(t, err)
		assert.Equal(t, "Petstore API", res.Collection.Name)
		assert.Equal(t, "Petstore API Flow", res.Flow.Name)
		assert.Len(t, res.Requests, 2)

		reqMethods := map[string]string{}
		for _, r := range res.Requests {
			reqMethods[r.Method] = r.URL
		}
		assert.Equal(t, "https://petstore.example.com/v1/pets", reqMethods["GET"])
		assert.Equal(t, "https://petstore.example.com/v1/pets", reqMethods["POST"])
	})

	t.Run("Export Workspace to OpenAPI 3.0", func(t *testing.T) {
		exported, err := svc.ExportWorkspaceOpenAPI(context.Background(), ownerID, wsID)
		require.NoError(t, err)

		var parsed map[string]any
		err = json.Unmarshal(exported, &parsed)
		require.NoError(t, err)

		assert.Equal(t, "3.0.3", parsed["openapi"])
		paths, ok := parsed["paths"].(map[string]any)
		require.True(t, ok)
		assert.Contains(t, paths, "/v1/pets")
	})
}
