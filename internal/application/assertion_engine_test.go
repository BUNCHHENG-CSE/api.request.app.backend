package application

import (
	"testing"
)

func TestEvaluateAssertions(t *testing.T) {
	res := &ExecutionResult{
		Status: 200,
		Time:   150,
		Headers: map[string]string{
			"Content-Type": "application/json; charset=utf-8",
			"X-Request-ID": "test-123",
		},
		Body: `{"status":"success","data":{"id":42,"username":"alice"}}`,
	}

	rules := `[
		{"target": "status", "operator": "equals", "expected": 200},
		{"target": "status", "operator": "not_equals", "expected": 404},
		{"target": "time", "operator": "less_than", "expected": 500},
		{"target": "header", "property": "Content-Type", "operator": "contains", "expected": "application/json"},
		{"target": "body", "operator": "contains", "expected": "alice"},
		{"target": "json_field", "property": "data.id", "operator": "equals", "expected": "42"},
		{"target": "json_field", "property": "data.username", "operator": "equals", "expected": "alice"},
		{"target": "json_field", "property": "data.id", "operator": "exists"},
		{"target": "json_field", "property": "nonexistent", "operator": "exists"}
	]`

	results := evaluateAssertions(rules, res)
	if len(results) != 9 {
		t.Fatalf("expected 9 results, got %d", len(results))
	}

	for i := 0; i < 8; i++ {
		if !results[i].Passed {
			t.Errorf("rule %d failed unexpectedly: %s", i, results[i].Message)
		}
	}

	// 9th rule checks nonexistent field exists -> should fail
	if results[8].Passed {
		t.Errorf("rule 8 should have failed for nonexistent field")
	}
}
