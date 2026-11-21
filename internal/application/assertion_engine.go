package application

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type AssertionRule struct {
	Target   string `json:"target"`             // "status", "time", "header", "body", "json_field"
	Property string `json:"property,omitempty"` // header name or JSON path (e.g. "data.id")
	Operator string `json:"operator"`           // "equals", "not_equals", "less_than", "greater_than", "contains", "exists"
	Expected any    `json:"expected,omitempty"`
}

type AssertionResult struct {
	Target   string `json:"target"`
	Property string `json:"property,omitempty"`
	Operator string `json:"operator"`
	Expected any    `json:"expected,omitempty"`
	Actual   any    `json:"actual,omitempty"`
	Passed   bool   `json:"passed"`
	Message  string `json:"message"`
}

func evaluateAssertions(rulesRaw string, res *ExecutionResult) []AssertionResult {
	rulesRaw = strings.TrimSpace(rulesRaw)
	if rulesRaw == "" || rulesRaw == "[]" {
		return nil
	}

	var rules []AssertionRule
	if err := json.Unmarshal([]byte(rulesRaw), &rules); err != nil {
		return nil
	}

	results := make([]AssertionResult, 0, len(rules))

	var parsedJSONBody map[string]any
	var parsedArrayBody []any
	_ = json.Unmarshal([]byte(res.Body), &parsedJSONBody)
	_ = json.Unmarshal([]byte(res.Body), &parsedArrayBody)

	for _, rule := range rules {
		target := strings.ToLower(strings.TrimSpace(rule.Target))
		operator := strings.ToLower(strings.TrimSpace(rule.Operator))
		result := AssertionResult{
			Target:   rule.Target,
			Property: rule.Property,
			Operator: rule.Operator,
			Expected: rule.Expected,
		}

		switch target {
		case "status", "status_code":
			result.Actual = res.Status
			result.Passed, result.Message = compareNumeric(int64(res.Status), rule.Expected, operator, "Status code")

		case "time", "duration", "latency":
			result.Actual = res.Time
			result.Passed, result.Message = compareNumeric(res.Time, rule.Expected, operator, "Response time")

		case "header":
			prop := strings.TrimSpace(rule.Property)
			var headerVal string
			for k, v := range res.Headers {
				if strings.EqualFold(k, prop) {
					headerVal = v
					break
				}
			}
			result.Actual = headerVal
			result.Passed, result.Message = compareString(headerVal, rule.Expected, operator, fmt.Sprintf("Header '%s'", prop))

		case "body":
			result.Actual = res.Body
			result.Passed, result.Message = compareString(res.Body, rule.Expected, operator, "Body")

		case "json", "json_field":
			var val any
			var exists bool
			if parsedJSONBody != nil {
				val, exists = extractJSONPath(parsedJSONBody, rule.Property)
			} else if parsedArrayBody != nil {
				val, exists = extractJSONPath(parsedArrayBody, rule.Property)
			}
			if !exists && parsedArrayBody != nil && rule.Property == "length" {
				val = len(parsedArrayBody)
				exists = true
			}
			result.Actual = val
			if operator == "exists" {
				result.Passed = exists
				if exists {
					result.Message = fmt.Sprintf("JSON field '%s' exists", rule.Property)
				} else {
					result.Message = fmt.Sprintf("JSON field '%s' does not exist", rule.Property)
				}
			} else if !exists {
				result.Passed = false
				result.Message = fmt.Sprintf("JSON field '%s' does not exist", rule.Property)
			} else {
				expectedStr := fmt.Sprintf("%v", rule.Expected)
				actualStr := fmt.Sprintf("%v", val)
				if operator == "equals" || operator == "eq" {
					result.Passed = actualStr == expectedStr
					result.Message = fmt.Sprintf("JSON field '%s' equals %s", rule.Property, expectedStr)
				} else if operator == "contains" {
					result.Passed = strings.Contains(actualStr, expectedStr)
					result.Message = fmt.Sprintf("JSON field '%s' contains %s", rule.Property, expectedStr)
				} else {
					result.Passed = actualStr == expectedStr
					result.Message = fmt.Sprintf("JSON field '%s' comparison: %s", rule.Property, actualStr)
				}
			}

		default:
			result.Passed = false
			result.Message = fmt.Sprintf("Unsupported assertion target: %s", rule.Target)
		}

		results = append(results, result)
	}

	return results
}

func compareNumeric(actual int64, expected any, operator, label string) (bool, string) {
	expNum, err := toInt64(expected)
	if err != nil {
		return false, fmt.Sprintf("%s expected a numeric value, got: %v", label, expected)
	}

	switch operator {
	case "equals", "eq", "==":
		passed := actual == expNum
		return passed, fmt.Sprintf("%s %d equals %d", label, actual, expNum)
	case "not_equals", "ne", "!=":
		passed := actual != expNum
		return passed, fmt.Sprintf("%s %d not equals %d", label, actual, expNum)
	case "less_than", "lt", "<":
		passed := actual < expNum
		return passed, fmt.Sprintf("%s %d is less than %d", label, actual, expNum)
	case "greater_than", "gt", ">":
		passed := actual > expNum
		return passed, fmt.Sprintf("%s %d is greater than %d", label, actual, expNum)
	case "less_than_or_equal", "lte", "<=":
		passed := actual <= expNum
		return passed, fmt.Sprintf("%s %d is <= %d", label, actual, expNum)
	case "greater_than_or_equal", "gte", ">=":
		passed := actual >= expNum
		return passed, fmt.Sprintf("%s %d is >= %d", label, actual, expNum)
	default:
		return false, fmt.Sprintf("Unsupported numeric operator '%s'", operator)
	}
}

func compareString(actual string, expected any, operator, label string) (bool, string) {
	expStr := fmt.Sprintf("%v", expected)
	switch operator {
	case "equals", "eq", "==":
		passed := actual == expStr
		return passed, fmt.Sprintf("%s equals '%s'", label, expStr)
	case "not_equals", "ne", "!=":
		passed := actual != expStr
		return passed, fmt.Sprintf("%s does not equal '%s'", label, expStr)
	case "contains":
		passed := strings.Contains(actual, expStr)
		return passed, fmt.Sprintf("%s contains '%s'", label, expStr)
	case "not_contains":
		passed := !strings.Contains(actual, expStr)
		return passed, fmt.Sprintf("%s does not contain '%s'", label, expStr)
	case "exists":
		passed := len(actual) > 0
		return passed, fmt.Sprintf("%s exists", label)
	default:
		return false, fmt.Sprintf("Unsupported string operator '%s'", operator)
	}
}

func toInt64(val any) (int64, error) {
	switch v := val.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %v to integer", val)
	}
}

func extractJSONPath(data any, path string) (any, bool) {
	if data == nil || path == "" {
		return nil, false
	}
	path = strings.ReplaceAll(path, "[", ".")
	path = strings.ReplaceAll(path, "]", "")
	parts := strings.Split(path, ".")
	var current any = data

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		switch c := current.(type) {
		case map[string]any:
			val, exists := c[part]
			if !exists {
				return nil, false
			}
			current = val
		case []any:
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(c) {
				return nil, false
			}
			current = c[idx]
		default:
			return nil, false
		}
	}
	return current, true
}
