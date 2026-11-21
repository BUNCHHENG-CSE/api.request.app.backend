package application

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ExtractionRule specifies how to extract a variable from an HTTP response
type ExtractionRule struct {
	Name   string `json:"name"`             // Target variable name
	Source string `json:"source,omitempty"` // "body" (default) or "header"
	Path   string `json:"path"`             // JSON path (e.g. "data.token") or header key
}

var (
	bruSetVarRegex = regexp.MustCompile(`(?i)bru\.setVar\s*\(\s*["']([^"']+)["']\s*,\s*res\.body\.([a-zA-Z0-9_\-\.\[\]]+)\s*\)`)
	pmSetVarRegex  = regexp.MustCompile(`(?i)pm\.(?:environment|collectionVariables|globals)\.set\s*\(\s*["']([^"']+)["']\s*,\s*pm\.response\.json\(\)\.([a-zA-Z0-9_\-\.\[\]]+)\s*\)`)
	atExtractRegex = regexp.MustCompile(`(?i)@extract\s+([a-zA-Z0-9_\-]+)\s*=\s*(body|header)\.([a-zA-Z0-9_\-\.\[\]]+)`)
)

// ExtractVariables inspects scripts and settings for extraction rules and extracts values from the response
func ExtractVariables(scriptsRaw, settingsRaw string, body string, headers map[string]string) map[string]string {
	extracted := make(map[string]string)

	var parsedBodyJSON any
	if strings.TrimSpace(body) != "" {
		var obj any
		if err := json.Unmarshal([]byte(body), &obj); err == nil {
			parsedBodyJSON = obj
		}
	}

	rules := parseExtractionRules(scriptsRaw)
	rules = append(rules, parseExtractionRules(settingsRaw)...)

	// Apply structured rules
	for _, rule := range rules {
		name := strings.TrimSpace(rule.Name)
		if name == "" {
			continue
		}
		source := strings.ToLower(strings.TrimSpace(rule.Source))
		if source == "" || source == "body" {
			if parsedBodyJSON != nil {
				if val, ok := extractJSONPath(parsedBodyJSON, rule.Path); ok {
					extracted[name] = fmt.Sprintf("%v", val)
				}
			}
		} else if source == "header" {
			for hKey, hVal := range headers {
				if strings.EqualFold(hKey, rule.Path) {
					extracted[name] = hVal
					break
				}
			}
		}
	}

	// Also parse script patterns (Bruno & Postman & @extract)
	parseScriptDirectives(scriptsRaw, parsedBodyJSON, headers, extracted)
	parseScriptDirectives(settingsRaw, parsedBodyJSON, headers, extracted)

	return extracted
}

func parseExtractionRules(raw string) []ExtractionRule {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" || raw == "[]" {
		return nil
	}

	var rules []ExtractionRule

	// Format 1: Direct array of ExtractionRule
	if err := json.Unmarshal([]byte(raw), &rules); err == nil && len(rules) > 0 {
		return rules
	}

	// Format 2: Wrapper object with "extract" field
	var wrapper struct {
		Extract []ExtractionRule `json:"extract"`
	}
	if err := json.Unmarshal([]byte(raw), &wrapper); err == nil && len(wrapper.Extract) > 0 {
		return wrapper.Extract
	}

	// Format 3: Simple map[string]string where key=varName and value=bodyPath or source.path
	var simpleMap map[string]string
	if err := json.Unmarshal([]byte(raw), &simpleMap); err == nil && len(simpleMap) > 0 {
		for k, v := range simpleMap {
			v = strings.TrimSpace(v)
			if strings.HasPrefix(strings.ToLower(v), "header.") {
				rules = append(rules, ExtractionRule{
					Name:   k,
					Source: "header",
					Path:   strings.TrimPrefix(v, "header."),
				})
			} else {
				rules = append(rules, ExtractionRule{
					Name:   k,
					Source: "body",
					Path:   strings.TrimPrefix(v, "body."),
				})
			}
		}
		return rules
	}

	return nil
}

func parseScriptDirectives(script string, parsedBody any, headers map[string]string, extracted map[string]string) {
	if script == "" {
		return
	}

	// bru.setVar("token", res.body.token)
	bruMatches := bruSetVarRegex.FindAllStringSubmatch(script, -1)
	for _, m := range bruMatches {
		if len(m) >= 3 {
			varName := m[1]
			path := m[2]
			if parsedBody != nil {
				if val, ok := extractJSONPath(parsedBody, path); ok {
					extracted[varName] = fmt.Sprintf("%v", val)
				}
			}
		}
	}

	// pm.environment.set("token", pm.response.json().token)
	pmMatches := pmSetVarRegex.FindAllStringSubmatch(script, -1)
	for _, m := range pmMatches {
		if len(m) >= 3 {
			varName := m[1]
			path := m[2]
			if parsedBody != nil {
				if val, ok := extractJSONPath(parsedBody, path); ok {
					extracted[varName] = fmt.Sprintf("%v", val)
				}
			}
		}
	}

	// @extract token = body.data.token
	atMatches := atExtractRegex.FindAllStringSubmatch(script, -1)
	for _, m := range atMatches {
		if len(m) >= 4 {
			varName := m[1]
			src := strings.ToLower(m[2])
			path := m[3]
			if src == "body" && parsedBody != nil {
				if val, ok := extractJSONPath(parsedBody, path); ok {
					extracted[varName] = fmt.Sprintf("%v", val)
				}
			} else if src == "header" {
				for hKey, hVal := range headers {
					if strings.EqualFold(hKey, path) {
						extracted[varName] = hVal
						break
					}
				}
			}
		}
	}
}
