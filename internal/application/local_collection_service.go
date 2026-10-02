package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type LocalCollectionInfo struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type localSnapshot struct {
	Version          int                  `json:"version"`
	ExportedAt       time.Time            `json:"exported_at"`
	Workspace        domain.Workspace     `json:"workspace"`
	Flows            []domain.Flow        `json:"flows"`
	Requests         []domain.Request     `json:"requests"`
	Environments     []domain.Environment `json:"environments"`
	Collections      []domain.Collection  `json:"collections"`
	RequestFiles     map[string]string    `json:"request_files"`
	EnvironmentFiles map[string]string    `json:"environment_files"`
}

type LocalCollectionService interface {
	List(ctx context.Context, ownerID uuid.UUID) ([]LocalCollectionInfo, error)
	Get(ctx context.Context, ownerID, workspaceID uuid.UUID) (*LocalCollectionInfo, error)
	Export(ctx context.Context, ownerID, workspaceID uuid.UUID) (*LocalCollectionInfo, error)
	Import(ctx context.Context, ownerID, workspaceID uuid.UUID) (map[string]int, error)
	Delete(ctx context.Context, ownerID, workspaceID uuid.UUID) error
}

type localCollectionService struct {
	root            string
	workspaceRepo   domain.WorkspaceRepository
	flowRepo        domain.FlowRepository
	requestRepo     domain.RequestRepository
	environmentRepo domain.EnvironmentRepository
	collectionRepo  domain.CollectionRepository
}

func NewLocalCollectionService(root string, workspaceRepo domain.WorkspaceRepository, flowRepo domain.FlowRepository, requestRepo domain.RequestRepository, environmentRepo domain.EnvironmentRepository, collectionRepo domain.CollectionRepository) (LocalCollectionService, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, err
	}
	return &localCollectionService{root: abs, workspaceRepo: workspaceRepo, flowRepo: flowRepo, requestRepo: requestRepo, environmentRepo: environmentRepo, collectionRepo: collectionRepo}, nil
}

func (s *localCollectionService) workspaceDir(id uuid.UUID) string {
	return filepath.Join(s.root, id.String())
}

func (s *localCollectionService) safeWorkspaceDir(id uuid.UUID) (string, error) {
	dir := filepath.Clean(s.workspaceDir(id))
	rel, err := filepath.Rel(s.root, dir)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", errors.New("invalid local collection path")
	}
	return dir, nil
}

func safeLocalFile(root, relative string) (string, error) {
	if relative == "" || filepath.IsAbs(relative) {
		return "", errors.New("invalid local file path")
	}
	path := filepath.Clean(filepath.Join(root, filepath.FromSlash(relative)))
	rel, err := filepath.Rel(root, path)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", errors.New("local file path escapes collection root")
	}
	return path, nil
}

func atomicWrite(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".flowapi-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	_ = os.Remove(path)
	return os.Rename(tmpName, path)
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
		} else if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "item"
	}
	return result
}

func parseStringMap(raw string) map[string]string {
	values := map[string]string{}
	_ = json.Unmarshal([]byte(raw), &values)
	return values
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func renderRequest(req domain.Request) string {
	method := strings.ToLower(req.Method)
	bodyType := req.BodyType
	if bodyType == "" {
		if req.Body == "" {
			bodyType = "none"
		} else {
			bodyType = "text"
		}
	}
	var b strings.Builder
	name := strings.NewReplacer("\r", " ", "\n", " ").Replace(req.Name)
	fmt.Fprintf(&b, "meta {\n  name: %s\n  type: http\n  seq: %d\n}\n\n", name, req.Position+1)
	fmt.Fprintf(&b, "%s {\n  url: %s\n  body: %s\n  auth: none\n}\n", method, req.URL, bodyType)
	if headers := parseStringMap(req.Headers); len(headers) > 0 {
		b.WriteString("\nheaders {\n")
		for _, key := range sortedKeys(headers) {
			fmt.Fprintf(&b, "  %s: %s\n", key, headers[key])
		}
		b.WriteString("}\n")
	}
	if params := parseStringMap(req.QueryParams); len(params) > 0 {
		b.WriteString("\nparams:query {\n")
		for _, key := range sortedKeys(params) {
			fmt.Fprintf(&b, "  %s: %s\n", key, params[key])
		}
		b.WriteString("}\n")
	}
	if req.Body != "" {
		fmt.Fprintf(&b, "\nbody:%s {\n%s\n}\n", bodyType, req.Body)
	}
	return b.String()
}

func renderEnvironment(env domain.Environment) string {
	values := parseStringMap(env.Variables)
	var b strings.Builder
	b.WriteString("vars {\n")
	for _, key := range sortedKeys(values) {
		fmt.Fprintf(&b, "  %s: %s\n", key, values[key])
	}
	b.WriteString("}\n")
	return b.String()
}

func (s *localCollectionService) Export(ctx context.Context, ownerID, workspaceID uuid.UUID) (*LocalCollectionInfo, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID)
	if err != nil {
		return nil, err
	}
	flows, err := s.flowRepo.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	environments, err := s.environmentRepo.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	collections, err := s.collectionRepo.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	dir, err := s.safeWorkspaceDir(workspaceID)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	snapshot := localSnapshot{Version: 1, ExportedAt: time.Now().UTC(), Workspace: *workspace, Flows: flows, Environments: environments, Collections: collections, RequestFiles: map[string]string{}, EnvironmentFiles: map[string]string{}}
	for _, flow := range flows {
		requests, err := s.requestRepo.ListByFlow(ctx, flow.ID)
		if err != nil {
			return nil, err
		}
		snapshot.Requests = append(snapshot.Requests, requests...)
		for _, req := range requests {
			rel := filepath.Join("requests", flow.ID.String(), fmt.Sprintf("%04d-%s-%s.bru", req.Position+1, slug(req.Name), req.ID.String()[:8]))
			snapshot.RequestFiles[req.ID.String()] = filepath.ToSlash(rel)
			if err := atomicWrite(filepath.Join(dir, rel), []byte(renderRequest(req)), 0o640); err != nil {
				return nil, err
			}
		}
	}
	for _, env := range environments {
		rel := filepath.Join("environments", env.ID.String()+"-"+slug(env.Name)+".bru")
		snapshot.EnvironmentFiles[env.ID.String()] = filepath.ToSlash(rel)
		if err := atomicWrite(filepath.Join(dir, rel), []byte(renderEnvironment(env)), 0o640); err != nil {
			return nil, err
		}
	}
	brunoJSON, _ := json.MarshalIndent(map[string]any{"version": "1", "name": workspace.Name, "type": "collection", "ignore": []string{"node_modules", ".git"}}, "", "  ")
	if err := atomicWrite(filepath.Join(dir, "bruno.json"), append(brunoJSON, '\n'), 0o640); err != nil {
		return nil, err
	}
	manifest, _ := json.MarshalIndent(snapshot, "", "  ")
	if err := atomicWrite(filepath.Join(dir, "flowapi.json"), append(manifest, '\n'), 0o640); err != nil {
		return nil, err
	}
	return &LocalCollectionInfo{WorkspaceID: workspaceID, Name: workspace.Name, Path: dir, UpdatedAt: snapshot.ExportedAt}, nil
}

func (s *localCollectionService) readSnapshot(ownerID, workspaceID uuid.UUID) (*localSnapshot, string, error) {
	dir, err := s.safeWorkspaceDir(workspaceID)
	if err != nil {
		return nil, "", err
	}
	data, err := os.ReadFile(filepath.Join(dir, "flowapi.json"))
	if err != nil {
		return nil, "", err
	}
	var snapshot localSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, "", err
	}
	if snapshot.Workspace.ID != workspaceID || snapshot.Workspace.OwnerID != ownerID {
		return nil, "", errors.New("local collection ownership mismatch")
	}
	return &snapshot, dir, nil
}

func (s *localCollectionService) Get(ctx context.Context, ownerID, workspaceID uuid.UUID) (*LocalCollectionInfo, error) {
	if _, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID); err != nil {
		return nil, err
	}
	snapshot, dir, err := s.readSnapshot(ownerID, workspaceID)
	if err != nil {
		return nil, err
	}
	return &LocalCollectionInfo{WorkspaceID: workspaceID, Name: snapshot.Workspace.Name, Path: dir, UpdatedAt: snapshot.ExportedAt}, nil
}

func (s *localCollectionService) List(ctx context.Context, ownerID uuid.UUID) ([]LocalCollectionInfo, error) {
	workspaces, err := s.workspaceRepo.List(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	items := make([]LocalCollectionInfo, 0)
	for _, workspace := range workspaces {
		if item, err := s.Get(ctx, ownerID, workspace.ID); err == nil {
			items = append(items, *item)
		}
	}
	return items, nil
}

func findBlock(content, label string) string {
	start := strings.Index(content, label+" {")
	if start < 0 {
		return ""
	}
	open := strings.Index(content[start:], "{") + start
	depth := 0
	for i := open; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return strings.TrimSpace(content[open+1 : i])
			}
		}
	}
	return ""
}

func parseBlockMap(block string) map[string]string {
	values := map[string]string{}
	for _, line := range strings.Split(block, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(parts) == 2 {
			values[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return values
}

func parseRequestFile(content string, req *domain.Request) {
	meta := parseBlockMap(findBlock(content, "meta"))
	if name := meta["name"]; name != "" {
		req.Name = name
	}
	if seq, err := strconv.Atoi(meta["seq"]); err == nil && seq > 0 {
		req.Position = seq - 1
	}
	for _, method := range []string{"get", "post", "put", "patch", "delete", "head", "options"} {
		block := findBlock(content, method)
		if block == "" {
			continue
		}
		values := parseBlockMap(block)
		req.Method = strings.ToUpper(method)
		if values["url"] != "" {
			req.URL = values["url"]
		}
		if values["body"] != "" {
			req.BodyType = values["body"]
		}
		break
	}
	if values := parseBlockMap(findBlock(content, "headers")); len(values) > 0 {
		encoded, _ := json.Marshal(values)
		req.Headers = string(encoded)
	}
	if values := parseBlockMap(findBlock(content, "params:query")); len(values) > 0 {
		encoded, _ := json.Marshal(values)
		req.QueryParams = string(encoded)
	}
	if req.BodyType != "" && req.BodyType != "none" {
		if body := findBlock(content, "body:"+req.BodyType); body != "" {
			req.Body = body
		}
	}
}

func (s *localCollectionService) Import(ctx context.Context, ownerID, workspaceID uuid.UUID) (map[string]int, error) {
	if _, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID); err != nil {
		return nil, err
	}
	snapshot, dir, err := s.readSnapshot(ownerID, workspaceID)
	if err != nil {
		return nil, err
	}
	counts := map[string]int{"flows": 0, "requests": 0, "environments": 0, "collections": 0}

	for i := range snapshot.Flows {
		flow := &snapshot.Flows[i]
		if flow.WorkspaceID != workspaceID {
			return nil, errors.New("invalid flow workspace in local collection")
		}
		existing, lookupErr := s.flowRepo.GetByID(ctx, flow.ID)
		if lookupErr == nil && existing.WorkspaceID != workspaceID {
			return nil, errors.New("local flow ID belongs to another workspace")
		}
		if lookupErr != nil {
			err = s.flowRepo.Create(ctx, flow)
		} else {
			err = s.flowRepo.Update(ctx, flow)
		}
		if err != nil {
			return nil, err
		}
		counts["flows"]++
	}

	for i := range snapshot.Collections {
		item := &snapshot.Collections[i]
		if item.WorkspaceID != workspaceID {
			return nil, errors.New("invalid collection workspace in local collection")
		}
		existing, lookupErr := s.collectionRepo.GetByID(ctx, item.ID)
		if lookupErr == nil && existing.WorkspaceID != workspaceID {
			return nil, errors.New("local collection ID belongs to another workspace")
		}
		if lookupErr != nil {
			err = s.collectionRepo.Create(ctx, item)
		} else {
			err = s.collectionRepo.Update(ctx, item)
		}
		if err != nil {
			return nil, err
		}
		counts["collections"]++
	}

	for i := range snapshot.Environments {
		env := &snapshot.Environments[i]
		if env.WorkspaceID != workspaceID {
			return nil, errors.New("invalid environment workspace in local collection")
		}
		if rel := snapshot.EnvironmentFiles[env.ID.String()]; rel != "" {
			localFile, pathErr := safeLocalFile(dir, rel)
			if pathErr != nil {
				return nil, pathErr
			}
			if data, readErr := os.ReadFile(localFile); readErr == nil {
				values := parseBlockMap(findBlock(string(data), "vars"))
				encoded, _ := json.Marshal(values)
				env.Variables = string(encoded)
			}
		}
		existing, lookupErr := s.environmentRepo.GetByID(ctx, env.ID)
		if lookupErr == nil && existing.WorkspaceID != workspaceID {
			return nil, errors.New("local environment ID belongs to another workspace")
		}
		if lookupErr != nil {
			err = s.environmentRepo.Create(ctx, env)
		} else {
			err = s.environmentRepo.Update(ctx, env)
		}
		if err != nil {
			return nil, err
		}
		counts["environments"]++
	}

	flowIDs := map[uuid.UUID]bool{}
	for _, flow := range snapshot.Flows {
		flowIDs[flow.ID] = true
	}
	for i := range snapshot.Requests {
		req := &snapshot.Requests[i]
		if !flowIDs[req.FlowID] {
			return nil, errors.New("invalid request flow in local collection")
		}
		if rel := snapshot.RequestFiles[req.ID.String()]; rel != "" {
			localFile, pathErr := safeLocalFile(dir, rel)
			if pathErr != nil {
				return nil, pathErr
			}
			if data, readErr := os.ReadFile(localFile); readErr == nil {
				parseRequestFile(string(data), req)
			}
		}
		existing, lookupErr := s.requestRepo.GetByID(ctx, req.ID)
		if lookupErr == nil {
			existingFlow, flowErr := s.flowRepo.GetByID(ctx, existing.FlowID)
			if flowErr != nil || existingFlow.WorkspaceID != workspaceID {
				return nil, errors.New("local request ID belongs to another workspace")
			}
			err = s.requestRepo.Update(ctx, req)
		} else {
			err = s.requestRepo.Create(ctx, req)
		}
		if err != nil {
			return nil, err
		}
		counts["requests"]++
	}
	return counts, nil
}

func (s *localCollectionService) Delete(ctx context.Context, ownerID, workspaceID uuid.UUID) error {
	if _, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID); err != nil {
		return err
	}
	dir, err := s.safeWorkspaceDir(workspaceID)
	if err != nil {
		return err
	}
	return os.RemoveAll(dir)
}
