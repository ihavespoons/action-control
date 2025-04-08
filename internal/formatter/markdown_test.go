package formatter

import (
	"strings"
	"testing"
)

func TestFormatMarkdown(t *testing.T) {
	// Create test data
	data := map[string][]Action{
		"org/repo1": {
			{Name: "Checkout", Uses: "actions/checkout@v3"},
			{Name: "Setup Node", Uses: "actions/setup-node@v2"},
		},
		"org/repo2": {
			{Name: "Checkout", Uses: "actions/checkout@v3"},
			{Name: "Custom Action", Uses: "custom/action@v1"},
		},
	}

	result := FormatMarkdown(data)

	// Check for expected content
	expectedHeaders := []string{
		"# GitHub Actions Usage Report",
		"## Actions by Repository",
		"### org/repo1",
		"### org/repo2",
		"## Most Used Actions",
	}

	for _, header := range expectedHeaders {
		if !strings.Contains(result, header) {
			t.Errorf("Expected report to contain %q, but it doesn't", header)
		}
	}

	// Check for action references
	expectedActions := []string{
		"actions/checkout@v3",
		"actions/setup-node@v2",
		"custom/action@v1",
	}

	for _, action := range expectedActions {
		if !strings.Contains(result, action) {
			t.Errorf("Expected report to contain action %q, but it doesn't", action)
		}
	}

	// Check for usage count
	if !strings.Contains(result, "| `actions/checkout@v3` | 2 |") {
		t.Error("Expected checkout action to show count of 2")
	}
}

func TestFormatPolicyViolations(t *testing.T) {
	// Test with violations
	t.Run("with violations", func(t *testing.T) {
		violations := map[string][]string{
			"org/repo1": {"unsafe/action@v1", "another/bad-action@v2"},
			"org/repo2": {"third/violation@v3"},
		}

		result := FormatPolicyViolations(violations)

		expectedPhrases := []string{
			"# Policy Violation Report",
			"## ❌ Policy Violations",
			"### org/repo1",
			"### org/repo2",
			"unsafe/action@v1",
			"third/violation@v3",
			"Found 2 repositories with policy violations",
		}

		for _, phrase := range expectedPhrases {
			if !strings.Contains(result, phrase) {
				t.Errorf("Expected report to contain %q, but it doesn't", phrase)
			}
		}
	})

	// Test without violations
	t.Run("without violations", func(t *testing.T) {
		violations := map[string][]string{}

		result := FormatPolicyViolations(violations)

		expected := "✅ All repositories comply with the action policy."
		if !strings.Contains(result, expected) {
			t.Errorf("Expected report to contain %q, but it doesn't", expected)
		}
	})
}
