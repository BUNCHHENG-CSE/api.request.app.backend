package application

import (
	"context"
	"strings"
	"testing"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

func TestSnippetServiceGenerators(t *testing.T) {
	ownerID, workspaceID, flowID := uuid.New(), uuid.New(), uuid.New()
	reqRepo := newRequestRepoMemory()
	flows := &flowRepoMemory{flow: &domain.Flow{ID: flowID, WorkspaceID: workspaceID}}
	workspaces := &workspaceRepoMemory{workspace: &domain.Workspace{ID: workspaceID, OwnerID: ownerID}}
	reqService := NewRequestService(reqRepo, flows, workspaces, &collectionRepoMemory{}, &environmentRepoMemory{}, &historyRepoMemory{}, true)

	req := &domain.Request{
		ID:       uuid.New(),
		FlowID:   flowID,
		Name:     "Create Item",
		Method:   "POST",
		URL:      "https://api.example.com/items",
		Headers:  `{"Content-Type":"application/json","X-Key":"123"}`,
		Body:     `{"name":"Widget"}`,
		BodyType: "json",
		Auth:     `{"type":"bearer","token":"my-jwt"}`,
	}
	_ = reqRepo.Create(context.Background(), req)

	snippetSvc := NewSnippetService(reqService)

	// 1. cURL
	curlCode, err := snippetSvc.Generate(context.Background(), ownerID, req.ID, "curl")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(curlCode, "curl -X POST") || !strings.Contains(curlCode, "-H 'Authorization: Bearer my-jwt'") {
		t.Fatalf("unexpected curl snippet: %s", curlCode)
	}

	// 2. JavaScript / Fetch
	jsCode, err := snippetSvc.Generate(context.Background(), ownerID, req.ID, "javascript")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(jsCode, "fetch('https://api.example.com/items'") || !strings.Contains(jsCode, "method: 'POST'") {
		t.Fatalf("unexpected fetch snippet: %s", jsCode)
	}

	// 3. Python / requests
	pyCode, err := snippetSvc.Generate(context.Background(), ownerID, req.ID, "python")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(pyCode, "import requests") || !strings.Contains(pyCode, "requests.post") {
		t.Fatalf("unexpected python snippet: %s", pyCode)
	}

	// 4. Golang / net/http
	goCode, err := snippetSvc.Generate(context.Background(), ownerID, req.ID, "golang")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(goCode, "http.NewRequest(\"POST\"") || !strings.Contains(goCode, "net/http") {
		t.Fatalf("unexpected go snippet: %s", goCode)
	}
}
