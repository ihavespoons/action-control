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

	// Test allow mode export
	t.Run("allow mode export", func(t *testing.T) {
		exporter := NewExporter()
		exporter.PolicyMode = "allow"

		policy, err := exporter.GeneratePolicyFromActions(actionsMap)
		if err != nil {
			t.Fatalf("GeneratePolicyFromActions returned error: %v", err)
		}

		// Check policy mode
		if policy.PolicyMode != "allow" {
			t.Errorf("Expected policy mode to be 'allow', got %q", policy.PolicyMode)
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

		// Check that denied actions list is empty
		if len(policy.DeniedActions) > 0 {
			t.Errorf("Expected denied actions to be empty, got %v", policy.DeniedActions)
		}
	})

	// Test deny mode export
	t.Run("deny mode export", func(t *testing.T) {
		exporter := NewExporter()
		exporter.PolicyMode = "deny"

		policy, err := exporter.GeneratePolicyFromActions(actionsMap)
		if err != nil {
			t.Fatalf("GeneratePolicyFromActions returned error: %v", err)
		}

		// Check policy mode
		if policy.PolicyMode != "deny" {
			t.Errorf("Expected policy mode to be 'deny', got %q", policy.PolicyMode)
		}

		// Check for expected denied actions without versions
		expectedActions := []string{
			"actions/checkout",
			"actions/setup-node",
			"custom/action",
		}

		if len(policy.DeniedActions) != len(expectedActions) {
			t.Errorf("Expected %d denied actions, got %d", len(expectedActions), len(policy.DeniedActions))
		}

		for _, expected := range expectedActions {
			found := false
			for _, actual := range policy.DeniedActions {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected policy to include denied action %q", expected)
			}
		}

		// Check that allowed actions list is empty
		if len(policy.AllowedActions) > 0 {
			t.Errorf("Expected allowed actions to be empty, got %v", policy.AllowedActions)
		}
	})

	// Test invalid mode
	t.Run("invalid mode", func(t *testing.T) {
		exporter := NewExporter()
		exporter.PolicyMode = "invalid"

		_, err := exporter.GeneratePolicyFromActions(actionsMap)
		if err == nil {
			t.Fatal("Expected error for invalid policy mode, but got nil")
		}
	})

	// Test with custom rules
	t.Run("include custom rules", func(t *testing.T) {
		exporter := NewExporter()
		exporter.IncludeCustom = true

		// Test with allow mode
		t.Run("allow mode", func(t *testing.T) {
			exporter.PolicyMode = "allow"

			policy, err := exporter.GeneratePolicyFromActions(actionsMap)
			if err != nil {
				t.Fatalf("GeneratePolicyFromActions returned error: %v", err)
			}

			// Check for custom rules
			if len(policy.CustomRules) != 2 {
				t.Errorf("Expected 2 custom rules, got %d", len(policy.CustomRules))
			}

			// Check repo1 rules - should be in allow mode
			repo1Rule, exists := policy.CustomRules["org/repo1"]
			if !exists {
				t.Error("Expected custom rule for org/repo1")
			} else {
				if repo1Rule.PolicyMode != "allow" {
					t.Errorf("Expected org/repo1 policy mode to be 'allow', got %q", repo1Rule.PolicyMode)
				}
				if len(repo1Rule.AllowedActions) != 2 {
					t.Errorf("Expected 2 allowed actions for org/repo1, got %d", len(repo1Rule.AllowedActions))
				}
			}
		})

		// Test with deny mode
		t.Run("deny mode", func(t *testing.T) {
			exporter.PolicyMode = "deny"

			policy, err := exporter.GeneratePolicyFromActions(actionsMap)
			if err != nil {
				t.Fatalf("GeneratePolicyFromActions returned error: %v", err)
			}

			// Check repo2 rules - should be in deny mode
			repo2Rule, exists := policy.CustomRules["org/repo2"]
			if !exists {
				t.Error("Expected custom rule for org/repo2")
			} else {
				if repo2Rule.PolicyMode != "deny" {
					t.Errorf("Expected org/repo2 policy mode to be 'deny', got %q", repo2Rule.PolicyMode)
				}
				if len(repo2Rule.DeniedActions) != 2 {
					t.Errorf("Expected 2 denied actions for org/repo2, got %d", len(repo2Rule.DeniedActions))
				}
			}
		})
	})
}

func TestExportPolicyFile(t *testing.T) {
	// Test exporting both types of policy files
	t.Run("export allow mode policy", func(t *testing.T) {
		// Create a temp directory for test files
		tempDir := t.TempDir()
		testFilePath := filepath.Join(tempDir, "test-policy-allow.yaml")

		// Create test data
		actionsMap := map[string][]github.Action{
			"org/repo1": {
				{Name: "Checkout", Uses: "actions/checkout@v3"},
			},
		}

		exporter := NewExporter()
		exporter.OutputPath = testFilePath
		exporter.PolicyMode = "allow"

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

		// Check for allowed actions
		if !strings.Contains(contentStr, "allowed_actions:") {
			t.Error("Expected file to contain allowed_actions section")
		}

		// Check for policy mode
		if !strings.Contains(contentStr, "policy_mode: allow") {
			t.Error("Expected file to contain 'policy_mode: allow'")
		}
	})

	t.Run("export deny mode policy", func(t *testing.T) {
		// Create a temp directory for test files
		tempDir := t.TempDir()
		testFilePath := filepath.Join(tempDir, "test-policy-deny.yaml")

		// Create test data
		actionsMap := map[string][]github.Action{
			"org/repo1": {
				{Name: "Checkout", Uses: "actions/checkout@v3"},
			},
		}

		exporter := NewExporter()
		exporter.OutputPath = testFilePath
		exporter.PolicyMode = "deny"

		policy, err := exporter.GeneratePolicyFromActions(actionsMap)
		if err != nil {
			t.Fatalf("GeneratePolicyFromActions returned error: %v", err)
		}

		// Export the policy
		if err := exporter.ExportPolicyFile(policy); err != nil {
			t.Fatalf("ExportPolicyFile returned error: %v", err)
		}

		// Read file contents
		content, err := os.ReadFile(testFilePath)
		if err != nil {
			t.Fatalf("Failed to read policy file: %v", err)
		}

		contentStr := string(content)

		// Check for denied actions
		if !strings.Contains(contentStr, "denied_actions:") {
			t.Error("Expected file to contain denied_actions section")
		}

		// Check for policy mode
		if !strings.Contains(contentStr, "policy_mode: deny") {
			t.Error("Expected file to contain 'policy_mode: deny'")
		}
	})
}
