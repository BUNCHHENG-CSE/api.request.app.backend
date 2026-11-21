package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type SnippetService interface {
	Generate(ctx context.Context, ownerID, requestID uuid.UUID, language string) (string, error)
}

type snippetService struct {
	requestService RequestService
}

func NewSnippetService(requestService RequestService) SnippetService {
	return &snippetService{requestService: requestService}
}

func (s *snippetService) Generate(ctx context.Context, ownerID, requestID uuid.UUID, language string) (string, error) {
	req, err := s.requestService.GetRequest(ctx, ownerID, requestID)
	if err != nil {
		return "", err
	}

	lang := strings.ToLower(strings.TrimSpace(language))
	if lang == "" {
		lang = "curl"
	}

	headers, _ := parseKeyValueData(req.Headers, "headers")
	params, _ := parseKeyValueData(req.QueryParams, "query_params")

	// Apply auth to headers or params if defined
	if req.Auth != "" && req.Auth != "{}" {
		var authMap map[string]any
		if err := json.Unmarshal([]byte(req.Auth), &authMap); err == nil {
			t := strings.ToLower(fmt.Sprintf("%v", authMap["type"]))
			if t == "bearer" || authMap["bearer"] != nil {
				token := fmt.Sprintf("%v", authMap["token"])
				if token == "" || token == "<nil>" {
					token = fmt.Sprintf("%v", authMap["bearer"])
				}
				if headers == nil {
					headers = make(map[string]string)
				}
				if headers["Authorization"] == "" {
					headers["Authorization"] = "Bearer " + token
				}
			}
		}
	}

	switch lang {
	case "curl", "bash", "shell":
		return generateCurlSnippet(req, headers, params), nil
	case "javascript", "js", "fetch":
		return generateFetchSnippet(req, headers, params), nil
	case "python", "py", "requests":
		return generatePythonSnippet(req, headers, params), nil
	case "golang", "go":
		return generateGoSnippet(req, headers, params), nil
	default:
		return generateCurlSnippet(req, headers, params), nil
	}
}

func generateCurlSnippet(req *domain.Request, headers, params map[string]string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("curl -X %s \\\n  '%s'", req.Method, req.URL))

	for k, v := range headers {
		b.WriteString(fmt.Sprintf(" \\\n  -H '%s: %s'", k, v))
	}
	if req.Body != "" && req.Method != "GET" && req.Method != "HEAD" {
		escaped := strings.ReplaceAll(req.Body, "'", `'\''`)
		b.WriteString(fmt.Sprintf(" \\\n  -d '%s'", escaped))
	}
	return b.String()
}

func generateFetchSnippet(req *domain.Request, headers, params map[string]string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("fetch('%s', {\n", req.URL))
	b.WriteString(fmt.Sprintf("  method: '%s',\n", req.Method))

	if len(headers) > 0 {
		b.WriteString("  headers: {\n")
		for k, v := range headers {
			b.WriteString(fmt.Sprintf("    '%s': '%s',\n", k, v))
		}
		b.WriteString("  },\n")
	}

	if req.Body != "" && req.Method != "GET" && req.Method != "HEAD" {
		if strings.HasPrefix(strings.TrimSpace(req.Body), "{") || strings.HasPrefix(strings.TrimSpace(req.Body), "[") {
			b.WriteString(fmt.Sprintf("  body: JSON.stringify(%s)\n", req.Body))
		} else {
			b.WriteString(fmt.Sprintf("  body: %q\n", req.Body))
		}
	}
	b.WriteString("})\n.then(response => response.json())\n.then(data => console.log(data))\n.catch(error => console.error(error));")
	return b.String()
}

func generatePythonSnippet(req *domain.Request, headers, params map[string]string) string {
	var b strings.Builder
	b.WriteString("import requests\n\n")
	b.WriteString(fmt.Sprintf("url = '%s'\n", req.URL))

	if len(headers) > 0 {
		b.WriteString("headers = {\n")
		for k, v := range headers {
			b.WriteString(fmt.Sprintf("    '%s': '%s',\n", k, v))
		}
		b.WriteString("}\n")
	} else {
		b.WriteString("headers = {}\n")
	}

	if req.Body != "" && req.Method != "GET" && req.Method != "HEAD" {
		b.WriteString(fmt.Sprintf("payload = %s\n", req.Body))
		b.WriteString(fmt.Sprintf("response = requests.%s(url, headers=headers, json=payload)\n", strings.ToLower(req.Method)))
	} else {
		b.WriteString(fmt.Sprintf("response = requests.%s(url, headers=headers)\n", strings.ToLower(req.Method)))
	}
	b.WriteString("print(response.status_code)\nprint(response.text)")
	return b.String()
}

func generateGoSnippet(req *domain.Request, headers, params map[string]string) string {
	var b strings.Builder
	b.WriteString("package main\n\nimport (\n\t\"fmt\"\n\t\"io\"\n\t\"net/http\"\n")
	if req.Body != "" && req.Method != "GET" && req.Method != "HEAD" {
		b.WriteString("\t\"strings\"\n")
	}
	b.WriteString(")\n\nfunc main() {\n")

	if req.Body != "" && req.Method != "GET" && req.Method != "HEAD" {
		b.WriteString(fmt.Sprintf("\tbody := strings.NewReader(%q)\n", req.Body))
		b.WriteString(fmt.Sprintf("\treq, err := http.NewRequest(%q, %q, body)\n", req.Method, req.URL))
	} else {
		b.WriteString(fmt.Sprintf("\treq, err := http.NewRequest(%q, %q, nil)\n", req.Method, req.URL))
	}
	b.WriteString("\tif err != nil {\n\t\tpanic(err)\n\t}\n")

	for k, v := range headers {
		b.WriteString(fmt.Sprintf("\treq.Header.Set(%q, %q)\n", k, v))
	}

	b.WriteString("\tres, err := http.DefaultClient.Do(req)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\tdefer res.Body.Close()\n")
	b.WriteString("\tdata, _ := io.ReadAll(res.Body)\n\tfmt.Println(res.Status)\n\tfmt.Println(string(data))\n}")
	return b.String()
}
