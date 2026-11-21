package application

import (
	"context"
	"strings"
	"testing"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

func TestImporterServiceCurl(t *testing.T) {
	ownerID, workspaceID, flowID := uuid.New(), uuid.New(), uuid.New()
	reqRepo := newRequestRepoMemory()
	flowRepo := &flowRepoMemory{flow: &domain.Flow{ID: flowID, WorkspaceID: workspaceID}}
	wsRepo := &workspaceRepoMemory{workspace: &domain.Workspace{ID: workspaceID, OwnerID: ownerID}}
	reqService := NewRequestService(reqRepo, flowRepo, wsRepo, &collectionRepoMemory{}, &environmentRepoMemory{}, &historyRepoMemory{}, true)
	flowService := NewFlowService(flowRepo, wsRepo)

	importer := NewImporterService(reqService, flowService, wsRepo, flowRepo, reqRepo, &collectionRepoMemory{})

	curlCmd := `curl -X POST 'https://api.example.com/v1/orders' -H 'Content-Type: application/json' -H 'Authorization: Bearer test-token' -d '{"item_id": 99}'`

	created, err := importer.ImportCurl(context.Background(), ownerID, flowID, curlCmd)
	if err != nil {
		t.Fatal(err)
	}

	if created.Method != "POST" {
		t.Fatalf("expected POST, got %s", created.Method)
	}
	if created.URL != "https://api.example.com/v1/orders" {
		t.Fatalf("expected URL, got %s", created.URL)
	}
	if !strings.Contains(created.Headers, "application/json") {
		t.Fatalf("headers not parsed: %s", created.Headers)
	}
	if !strings.Contains(created.Auth, "test-token") {
		t.Fatalf("bearer auth not parsed: %s", created.Auth)
	}
	if created.Body != `{"item_id": 99}` {
		t.Fatalf("body not parsed: %s", created.Body)
	}
}

func TestImporterServicePostmanCollection(t *testing.T) {
	ownerID, workspaceID := uuid.New(), uuid.New()
	reqRepo := newRequestRepoMemory()
	flowRepo := &flowRepoMemory{}
	wsRepo := &workspaceRepoMemory{workspace: &domain.Workspace{ID: workspaceID, OwnerID: ownerID}}
	colRepo := &collectionRepoMemory{}
	reqService := NewRequestService(reqRepo, flowRepo, wsRepo, colRepo, &environmentRepoMemory{}, &historyRepoMemory{}, true)
	flowService := NewFlowService(flowRepo, wsRepo)

	importer := NewImporterService(reqService, flowService, wsRepo, flowRepo, reqRepo, colRepo)

	postmanJSON := `{
		"info": {
			"name": "E-Commerce API",
			"description": "API collection exported from Postman",
			"schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
		},
		"item": [
			{
				"name": "Get Products",
				"request": {
					"method": "GET",
					"header": [
						{"key": "Accept", "value": "application/json", "disabled": false}
					],
					"url": {
						"raw": "https://api.example.com/products"
					}
				}
			},
			{
				"name": "Create Product",
				"request": {
					"method": "POST",
					"header": [
						{"key": "Content-Type", "value": "application/json"}
					],
					"body": {
						"mode": "raw",
						"raw": "{\"title\":\"Laptop\"}"
					},
					"url": "https://api.example.com/products",
					"auth": {
						"type": "bearer",
						"bearer": [{"key": "token", "value": "postman-secret"}]
					}
				}
			}
		]
	}`

	result, err := importer.ImportPostman(context.Background(), ownerID, workspaceID, []byte(postmanJSON))
	if err != nil {
		t.Fatal(err)
	}

	if result.CollectionName != "E-Commerce API" {
		t.Fatalf("expected collection name 'E-Commerce API', got %s", result.CollectionName)
	}
	if result.RequestsCount != 2 {
		t.Fatalf("expected 2 requests imported, got %d", result.RequestsCount)
	}
}
