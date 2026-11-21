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
	requests := newRequestRepoMemory()
	_ = requests.Create(context.Background(), &domain.Request{ID: requestID, FlowID: flowID, Name: "List Users", Method: "GET", URL: "https://example.com/users", Headers: `{}`})
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
	req, err := requests.GetByID(context.Background(), requestID)
	if err != nil || counts["requests"] != 1 || req.URL != "https://example.com/v2/users" {
		t.Fatalf("local request edit was not imported: %+v, err: %v", req, err)
	}
}

func TestLocalCollectionZipExportAndImport(t *testing.T) {
	ownerID, workspaceID, flowID, requestID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	workspaces := &workspaceRepoMemory{workspace: &domain.Workspace{ID: workspaceID, OwnerID: ownerID, Name: "ZipDemo"}}
	flows := &flowRepoMemory{flow: &domain.Flow{ID: flowID, WorkspaceID: workspaceID, Name: "Orders"}}
	requests := newRequestRepoMemory()
	_ = requests.Create(context.Background(), &domain.Request{ID: requestID, FlowID: flowID, Name: "Get Order", Method: "GET", URL: "https://api.example.com/orders", Headers: `{}`})

	service, err := NewLocalCollectionService(t.TempDir(), workspaces, flows, requests, &environmentRepoMemory{}, &collectionRepoMemory{})
	if err != nil {
		t.Fatal(err)
	}

	zipBytes, filename, err := service.ExportZip(context.Background(), ownerID, workspaceID)
	if err != nil {
		t.Fatal(err)
	}
	if len(zipBytes) == 0 || !strings.HasSuffix(filename, ".zip") {
		t.Fatalf("invalid zip export: filename=%s, len=%d", filename, len(zipBytes))
	}

	// Now import that zip into another service instance
	newDir := t.TempDir()
	service2, err := NewLocalCollectionService(newDir, workspaces, flows, requests, &environmentRepoMemory{}, &collectionRepoMemory{})
	if err != nil {
		t.Fatal(err)
	}

	counts, err := service2.ImportZip(context.Background(), ownerID, workspaceID, zipBytes)
	if err != nil {
		t.Fatal(err)
	}
	if counts["requests"] != 1 {
		t.Fatalf("expected 1 imported request, got: %+v", counts)
	}

	// Verify file tree on local disk
	tree, err := service2.GetFileTree(context.Background(), ownerID, workspaceID)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree) == 0 {
		t.Fatal("expected file tree to contain files")
	}

	// Read and write local file
	brunoContent, err := service2.ReadLocalFile(context.Background(), ownerID, workspaceID, "bruno.json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(brunoContent, "ZipDemo") {
		t.Fatalf("expected ZipDemo in bruno.json: %s", brunoContent)
	}
}
