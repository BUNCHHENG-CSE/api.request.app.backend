package application

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

func TestLocalCollectionExportAndImport(t *testing.T) {
	ownerID, workspaceID, flowID, requestID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	workspaces := &workspaceRepoMemory{workspace: &domain.Workspace{ID: workspaceID, OwnerID: ownerID, Name: "Demo"}}
	flows := &flowRepoMemory{flow: &domain.Flow{ID: flowID, WorkspaceID: workspaceID, Name: "Users"}}
	requests := &requestRepoMemory{request: &domain.Request{ID: requestID, FlowID: flowID, Name: "List Users", Method: "GET", URL: "https://example.com/users", Headers: `{}`}}
	environments := &environmentRepoMemory{}
	collections := &collectionRepoMemory{}

	service, err := NewLocalCollectionService(t.TempDir(), workspaces, flows, requests, environments, collections)
	if err != nil {
		t.Fatal(err)
	}
	info, err := service.Export(context.Background(), ownerID, workspaceID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(info.Path, "bruno.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(info.Path, "flowapi.json")); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(filepath.Join(info.Path, "requests", flowID.String()))
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected one Bruno request file: %v", err)
	}
	requestFile := filepath.Join(info.Path, "requests", flowID.String(), entries[0].Name())
	data, err := os.ReadFile(requestFile)
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.Replace(string(data), "https://example.com/users", "https://example.com/v2/users", 1)
	if err := os.WriteFile(requestFile, []byte(updated), 0o640); err != nil {
		t.Fatal(err)
	}

	counts, err := service.Import(context.Background(), ownerID, workspaceID)
	if err != nil {
		t.Fatal(err)
	}
	if counts["requests"] != 1 || requests.request.URL != "https://example.com/v2/users" {
		t.Fatalf("local request edit was not imported: %+v", requests.request)
	}
}
