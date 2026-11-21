package application

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type DocsService interface {
	GenerateCollectionDocs(ctx context.Context, ownerID, collectionID uuid.UUID, format string) (string, string, error)
	GenerateWorkspaceDocs(ctx context.Context, ownerID, workspaceID uuid.UUID, format string) (string, string, error)
}

type docsService struct {
	workspaceRepo  domain.WorkspaceRepository
	collectionRepo domain.CollectionRepository
	flowRepo       domain.FlowRepository
	requestRepo    domain.RequestRepository
}

func NewDocsService(
	workspaceRepo domain.WorkspaceRepository,
	collectionRepo domain.CollectionRepository,
	flowRepo domain.FlowRepository,
	requestRepo domain.RequestRepository,
) DocsService {
	return &docsService{
		workspaceRepo:  workspaceRepo,
		collectionRepo: collectionRepo,
		flowRepo:       flowRepo,
		requestRepo:    requestRepo,
	}
}

func (s *docsService) GenerateCollectionDocs(ctx context.Context, ownerID, collectionID uuid.UUID, format string) (string, string, error) {
	col, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return "", "", err
	}
	if _, err := s.workspaceRepo.GetByID(ctx, col.WorkspaceID, ownerID); err != nil {
		return "", "", err
	}

	flows, err := s.flowRepo.ListByWorkspace(ctx, col.WorkspaceID)
	if err != nil {
		return "", "", err
	}

	var reqs []domain.Request
	for _, f := range flows {
		list, _ := s.requestRepo.ListByFlow(ctx, f.ID)
		for _, r := range list {
			if r.CollectionID != nil && *r.CollectionID == collectionID {
				reqs = append(reqs, r)
			}
		}
	}

	return s.renderDocs(col.Name, col.Description, reqs, format)
}

func (s *docsService) GenerateWorkspaceDocs(ctx context.Context, ownerID, workspaceID uuid.UUID, format string) (string, string, error) {
	ws, err := s.workspaceRepo.GetByID(ctx, workspaceID, ownerID)
	if err != nil {
		return "", "", err
	}

	flows, err := s.flowRepo.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return "", "", err
	}

	var reqs []domain.Request
	for _, f := range flows {
		list, _ := s.requestRepo.ListByFlow(ctx, f.ID)
		reqs = append(reqs, list...)
	}

	return s.renderDocs(ws.Name, "Complete workspace API documentation", reqs, format)
}

func (s *docsService) renderDocs(title, description string, requests []domain.Request, format string) (string, string, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "html" {
		return s.renderHTML(title, description, requests), "text/html; charset=utf-8", nil
	}
	return s.renderMarkdown(title, description, requests), "text/markdown; charset=utf-8", nil
}

func (s *docsService) renderMarkdown(title, description string, requests []domain.Request) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))
	if description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", description))
	}
	sb.WriteString(fmt.Sprintf("*Generated %d API endpoints*\n\n---\n\n", len(requests)))

	for _, r := range requests {
		sb.WriteString(fmt.Sprintf("## %s\n\n", r.Name))
		sb.WriteString(fmt.Sprintf("`%s` `%s`\n\n", r.Method, r.URL))

		// Headers
		headers := parseKeyValueArray(r.Headers)
		if len(headers) > 0 {
			sb.WriteString("### Headers\n\n")
			sb.WriteString("| Key | Value |\n|---|---|\n")
			for k, v := range headers {
				sb.WriteString(fmt.Sprintf("| `%s` | `%s` |\n", k, v))
			}
			sb.WriteString("\n")
		}

		// Query Params
		queryParams := parseKeyValueArray(r.QueryParams)
		if len(queryParams) > 0 {
			sb.WriteString("### Query Parameters\n\n")
			sb.WriteString("| Parameter | Example |\n|---|---|\n")
			for k, v := range queryParams {
				sb.WriteString(fmt.Sprintf("| `%s` | `%s` |\n", k, v))
			}
			sb.WriteString("\n")
		}

		// Body
		if strings.TrimSpace(r.Body) != "" {
			sb.WriteString("### Request Body\n\n")
			sb.WriteString("```json\n")
			sb.WriteString(r.Body)
			sb.WriteString("\n```\n\n")
		}

		sb.WriteString("---\n\n")
	}

	return sb.String()
}

func (s *docsService) renderHTML(title, description string, requests []domain.Request) string {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>` + html.EscapeString(title) + ` - API Documentation</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1e293b; background: #f8fafc; margin: 0; padding: 40px 20px; }
  .container { max-width: 900px; margin: 0 auto; background: #ffffff; padding: 40px; border-radius: 8px; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); }
  h1 { color: #0f172a; margin-top: 0; border-bottom: 2px solid #e2e8f0; padding-bottom: 12px; }
  .endpoint { border: 1px solid #e2e8f0; border-radius: 6px; padding: 20px; margin-bottom: 24px; background: #ffffff; }
  .method { display: inline-block; padding: 4px 10px; border-radius: 4px; font-weight: bold; font-size: 14px; text-transform: uppercase; color: #fff; margin-right: 10px; }
  .method-get { background: #10b981; }
  .method-post { background: #3b82f6; }
  .method-put { background: #f59e0b; }
  .method-delete { background: #ef4444; }
  .method-patch { background: #8b5cf6; }
  .url { font-family: monospace; font-size: 15px; color: #334155; }
  pre { background: #0f172a; color: #f8fafc; padding: 14px; border-radius: 6px; overflow-x: auto; font-family: monospace; font-size: 13px; }
  table { width: 100%; border-collapse: collapse; margin-top: 10px; margin-bottom: 16px; }
  th, td { text-align: left; padding: 8px 12px; border-bottom: 1px solid #e2e8f0; font-size: 14px; }
  th { background: #f1f5f9; color: #475569; }
</style>
</head>
<body>
<div class="container">
<h1>` + html.EscapeString(title) + `</h1>
<p>` + html.EscapeString(description) + `</p>
`)

	for _, r := range requests {
		mLower := strings.ToLower(r.Method)
		sb.WriteString(`<div class="endpoint">
<h3>` + html.EscapeString(r.Name) + `</h3>
<div style="margin-bottom: 12px;">
<span class="method method-` + mLower + `">` + html.EscapeString(r.Method) + `</span>
<span class="url">` + html.EscapeString(r.URL) + `</span>
</div>`)

		headers := parseKeyValueArray(r.Headers)
		if len(headers) > 0 {
			sb.WriteString(`<h4>Headers</h4><table><tr><th>Key</th><th>Value</th></tr>`)
			for k, v := range headers {
				sb.WriteString(`<tr><td><code>` + html.EscapeString(k) + `</code></td><td>` + html.EscapeString(v) + `</td></tr>`)
			}
			sb.WriteString(`</table>`)
		}

		if strings.TrimSpace(r.Body) != "" {
			sb.WriteString(`<h4>Body</h4><pre>` + html.EscapeString(r.Body) + `</pre>`)
		}

		sb.WriteString(`</div>`)
	}

	sb.WriteString(`</div></body></html>`)
	return sb.String()
}

func parseKeyValueArray(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" || raw == "[]" {
		return nil
	}
	res := make(map[string]string)
	var arr []map[string]any
	if err := json.Unmarshal([]byte(raw), &arr); err == nil {
		for _, item := range arr {
			if enabled, ok := item["enabled"].(bool); ok && !enabled {
				continue
			}
			key := fmt.Sprintf("%v", item["key"])
			if key != "" && key != "<nil>" {
				res[key] = fmt.Sprintf("%v", item["value"])
			}
		}
		return res
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err == nil {
		for k, v := range m {
			res[k] = fmt.Sprintf("%v", v)
		}
		return res
	}
	return nil
}
