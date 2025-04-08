package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ihavespoons/action-control/internal/github"
)

func TestGeneratePolicyFromActions(t *testing.T) {
	// Create test data
	actionsMap := map[string][]github.Action{
		"org/repo1": {
			{Name: "Checkout", Uses: "actions/checkout@v3"},
			{Name: "Setup Node", Uses: "actions/setup-node@v2"},
		},
		"org/repo2": {
			{Name: "Checkout", Uses: "actions/checkout@v3"},
			{Name: "Custom Action", Uses: "custom/action@v1"},
		},
	}

	// Test standard export (no versions)
	t.Run("standard export", func(t *testing.T) {
		exporter := NewExporter()

		policy, err := exporter.GeneratePolicyFromActions(actionsMap)
		if err != nil {
			t.Fatalf("GeneratePolicyFromActions returned error: %v", err)
		}

		// Check for expected allowed actions without versions
		expectedActions := []string{
			"actions/checkout",
			"actions/setup-node",
			"custom/action",
		}

		if len(policy.AllowedActions) != len(expectedActions) {
			t.Errorf("Expected %d allowed actions, got %d", len(expectedActions), len(policy.AllowedActions))
		}

		for _, expected := range expectedActions {
			found := false
			for _, actual := range policy.AllowedActions {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected policy to include action %q", expected)
			}
		}

		// Check that no custom rules are generated
		if len(policy.CustomRules) > 0 {
			t.Errorf("Expected no custom rules, got %d", len(policy.CustomRules))
		}
	})

	// Test with versions included
	t.Run("include versions", func(t *testing.T) {
		exporter := NewExporter()
		exporter.IncludeVersions = true

		policy, err := exporter.GeneratePolicyFromActions(actionsMap)
		if err != nil {
			t.Fatalf("GeneratePolicyFromActions returned error: %v", err)
		}

		// Check that versions are included
		expectedActions := []string{
			"actions/checkout@v3",
			"actions/setup-node@v2",
			"custom/action@v1",
		}

		for _, expected := range expectedActions {
			found := false
			for _, actual := range policy.AllowedActions {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected policy to include action %q", expected)
			}
		}
	})

	// Test with custom rules
	t.Run("include custom rules", func(t *testing.T) {
		exporter := NewExporter()
		exporter.IncludeCustom = true

		policy, err := exporter.GeneratePolicyFromActions(actionsMap)
		if err != nil {
			t.Fatalf("GeneratePolicyFromActions returned error: %v", err)
		}

		// Check for custom rules
		if len(policy.CustomRules) != 2 {
			t.Errorf("Expected 2 custom rules, got %d", len(policy.CustomRules))
		}

		// Check repo1 rules
		repo1Rule, exists := policy.CustomRules["org/repo1"]
		if !exists {
			t.Error("Expected custom rule for org/repo1")
		} else {
			if len(repo1Rule.AllowedActions) != 2 {
				t.Errorf("Expected 2 allowed actions for org/repo1, got %d", len(repo1Rule.AllowedActions))
			}
		}

		// Check repo2 rules
		repo2Rule, exists := policy.CustomRules["org/repo2"]
		if !exists {
			t.Error("Expected custom rule for org/repo2")
		} else {
			if len(repo2Rule.AllowedActions) != 2 {
				t.Errorf("Expected 2 allowed actions for org/repo2, got %d", len(repo2Rule.AllowedActions))
			}
		}
	})
}

func TestExportPolicyFile(t *testing.T) {
	// Create a temp directory for test files
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "test-policy.yaml")

	// Create test data
	actionsMap := map[string][]github.Action{
		"org/repo1": {
			{Name: "Checkout", Uses: "actions/checkout@v3"},
			{Name: "Setup Node", Uses: "actions/setup-node@v2"},
		},
	}

	exporter := NewExporter()
	exporter.OutputPath = testFilePath

	policy, err := exporter.GeneratePolicyFromActions(actionsMap)
	if err != nil {
		t.Fatalf("GeneratePolicyFromActions returned error: %v", err)
	}

	// Export the policy
	if err := exporter.ExportPolicyFile(policy); err != nil {
		t.Fatalf("ExportPolicyFile returned error: %v", err)
	}

	// Check that file exists
	if _, err := os.Stat(testFilePath); err != nil {
		t.Fatalf("Expected policy file to exist at %s, but it doesn't: %v", testFilePath, err)
	}

	// Read file contents
	content, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read policy file: %v", err)
	}

	// Check file content
	contentStr := string(content)

	// Check header comment
	if !strings.Contains(contentStr, "# GitHub Action Control Policy") {
		t.Error("Expected file to contain header comment")
	}

	// Check for allowed actions
	if !strings.Contains(contentStr, "allowed_actions:") {
		t.Error("Expected file to contain allowed_actions section")
	}

	// Check for specific actions
	if !strings.Contains(contentStr, "- actions/checkout") {
		t.Error("Expected file to contain 'actions/checkout'")
	}

	if !strings.Contains(contentStr, "- actions/setup-node") {
		t.Error("Expected file to contain 'actions/setup-node'")
	}
}
