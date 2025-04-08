package policy

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadPolicyConfig(t *testing.T) {
	// Test allow mode policy
	t.Run("allow mode policy", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "policy-allow.yaml")

		// Write test policy file
		policyContent := `
policy_mode: allow
allowed_actions:
  - actions/checkout
  - actions/setup-node
excluded_repos:
  - org/test-repo
custom_rules:
  org/special-repo:
    allowed_actions:
      - actions/checkout
      - custom/action
`

		if err := os.WriteFile(testFile, []byte(policyContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Load the policy
		policy, err := LoadPolicyConfig(testFile)
		if err != nil {
			t.Fatalf("LoadPolicyConfig returned error: %v", err)
		}

		// Verify policy contents
		if policy.PolicyMode != "allow" {
			t.Errorf("Expected policy mode to be 'allow', got %q", policy.PolicyMode)
		}

		expectedAllowed := []string{"actions/checkout", "actions/setup-node"}
		if !reflect.DeepEqual(policy.AllowedActions, expectedAllowed) {
			t.Errorf("Expected allowed actions %v, got %v", expectedAllowed, policy.AllowedActions)
		}
	})

	// Test deny mode policy
	t.Run("deny mode policy", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "policy-deny.yaml")

		// Write test policy file
		policyContent := `
policy_mode: deny
denied_actions:
  - unsafe/action
  - deprecated/action
excluded_repos:
  - org/test-repo
custom_rules:
  org/special-repo:
    policy_mode: deny
    denied_actions:
      - special/unsafe-action
`

		if err := os.WriteFile(testFile, []byte(policyContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Load the policy
		policy, err := LoadPolicyConfig(testFile)
		if err != nil {
			t.Fatalf("LoadPolicyConfig returned error: %v", err)
		}

		// Verify policy contents
		if policy.PolicyMode != "deny" {
			t.Errorf("Expected policy mode to be 'deny', got %q", policy.PolicyMode)
		}

		expectedDenied := []string{"unsafe/action", "deprecated/action"}
		if !reflect.DeepEqual(policy.DeniedActions, expectedDenied) {
			t.Errorf("Expected denied actions %v, got %v", expectedDenied, policy.DeniedActions)
		}

		// Check custom rules
		customRule, exists := policy.CustomRules["org/special-repo"]
		if !exists {
			t.Error("Expected custom rule for org/special-repo, but didn't find it")
		} else {
			if customRule.PolicyMode != "deny" {
				t.Errorf("Expected custom rule policy mode to be 'deny', got %q", customRule.PolicyMode)
			}

			expectedCustomDenied := []string{"special/unsafe-action"}
			if !reflect.DeepEqual(customRule.DeniedActions, expectedCustomDenied) {
				t.Errorf("Expected custom denied actions %v, got %v", expectedCustomDenied, customRule.DeniedActions)
			}
		}
	})

	// Test automatic mode detection when not specified
	t.Run("automatic mode detection", func(t *testing.T) {
		tempDir := t.TempDir()

		// Test with allowed_actions present
		t.Run("with allowed_actions", func(t *testing.T) {
			testFile := filepath.Join(tempDir, "policy-auto-allow.yaml")
			policyContent := `
allowed_actions:
  - actions/checkout
`
			if err := os.WriteFile(testFile, []byte(policyContent), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			policy, err := LoadPolicyConfig(testFile)
			if err != nil {
				t.Fatalf("LoadPolicyConfig returned error: %v", err)
			}

			if policy.PolicyMode != "allow" {
				t.Errorf("Expected policy mode to be automatically set to 'allow', got %q", policy.PolicyMode)
			}
		})

		// Test with denied_actions present
		t.Run("with denied_actions", func(t *testing.T) {
			testFile := filepath.Join(tempDir, "policy-auto-deny.yaml")
			policyContent := `
denied_actions:
  - unsafe/action
`
			if err := os.WriteFile(testFile, []byte(policyContent), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			policy, err := LoadPolicyConfig(testFile)
			if err != nil {
				t.Fatalf("LoadPolicyConfig returned error: %v", err)
			}

			if policy.PolicyMode != "deny" {
				t.Errorf("Expected policy mode to be automatically set to 'deny', got %q", policy.PolicyMode)
			}
		})
	})
}

func TestCheckActionCompliance(t *testing.T) {
	// Test with allow mode policy
	t.Run("allow mode policy", func(t *testing.T) {
		policy := &PolicyConfig{
			PolicyMode:     "allow",
			AllowedActions: []string{"actions/checkout", "actions/setup-node"},
			ExcludedRepos:  []string{"org/excluded-repo"},
			CustomRules: map[string]Policy{
				"org/custom-repo": {
					PolicyMode:     "allow",
					AllowedActions: []string{"actions/checkout", "custom/special-action"},
				},
			},
		}

		testCases := []struct {
			name           string
			repo           string
			actions        []string
			wantViolations []string
			wantCompliant  bool
		}{
			{
				name:          "compliant standard repo",
				repo:          "org/standard-repo",
				actions:       []string{"actions/checkout@v3", "actions/setup-node@v2"},
				wantCompliant: true,
			},
			{
				name:           "non-compliant standard repo",
				repo:           "org/standard-repo",
				actions:        []string{"actions/checkout@v3", "unauthorized/action@v1"},
				wantViolations: []string{"unauthorized/action@v1"},
				wantCompliant:  false,
			},
			{
				name:          "excluded repo",
				repo:          "org/excluded-repo",
				actions:       []string{"any/action@v1", "another/action@v2"},
				wantCompliant: true,
			},
			{
				name:          "custom repo compliant",
				repo:          "org/custom-repo",
				actions:       []string{"actions/checkout@v2", "custom/special-action@v1"},
				wantCompliant: true,
			},
			{
				name:           "custom repo non-compliant",
				repo:           "org/custom-repo",
				actions:        []string{"actions/setup-node@v2"},
				wantViolations: []string{"actions/setup-node@v2"},
				wantCompliant:  false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				violations, compliant := CheckActionCompliance(policy, tc.repo, tc.actions)

				if compliant != tc.wantCompliant {
					t.Errorf("Expected compliant=%v, got %v", tc.wantCompliant, compliant)
				}

				if !reflect.DeepEqual(violations, tc.wantViolations) {
					t.Errorf("Expected violations %v, got %v", tc.wantViolations, violations)
				}
			})
		}
	})

	// Test with deny mode policy
	t.Run("deny mode policy", func(t *testing.T) {
		policy := &PolicyConfig{
			PolicyMode:    "deny",
			DeniedActions: []string{"unsafe/action", "deprecated/action"},
			ExcludedRepos: []string{"org/excluded-repo"},
			CustomRules: map[string]Policy{
				"org/custom-repo": {
					PolicyMode:    "deny",
					DeniedActions: []string{"custom/unsafe-action", "custom/deprecated-action"},
				},
			},
		}

		testCases := []struct {
			name           string
			repo           string
			actions        []string
			wantViolations []string
			wantCompliant  bool
		}{
			{
				name:          "compliant standard repo",
				repo:          "org/standard-repo",
				actions:       []string{"actions/checkout@v3", "actions/setup-node@v2"},
				wantCompliant: true,
			},
			{
				name:           "non-compliant standard repo",
				repo:           "org/standard-repo",
				actions:        []string{"actions/checkout@v3", "unsafe/action@v1"},
				wantViolations: []string{"unsafe/action@v1"},
				wantCompliant:  false,
			},
			{
				name:          "excluded repo",
				repo:          "org/excluded-repo",
				actions:       []string{"unsafe/action@v1", "deprecated/action@v2"},
				wantCompliant: true, // excluded repos are always compliant
			},
			{
				name:          "custom repo compliant",
				repo:          "org/custom-repo",
				actions:       []string{"actions/checkout@v2", "actions/setup-node@v2"},
				wantCompliant: true,
			},
			{
				name:           "custom repo non-compliant",
				repo:           "org/custom-repo",
				actions:        []string{"custom/unsafe-action@v1"},
				wantViolations: []string{"custom/unsafe-action@v1"},
				wantCompliant:  false,
			},
			{
				name:           "mixed actions with version info",
				repo:           "org/standard-repo",
				actions:        []string{"actions/checkout@v3", "deprecated/action"},
				wantViolations: []string{"deprecated/action"},
				wantCompliant:  false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				violations, compliant := CheckActionCompliance(policy, tc.repo, tc.actions)

				if compliant != tc.wantCompliant {
					t.Errorf("Expected compliant=%v, got %v", tc.wantCompliant, compliant)
				}

				if !reflect.DeepEqual(violations, tc.wantViolations) {
					t.Errorf("Expected violations %v, got %v", tc.wantViolations, violations)
				}
			})
		}
	})
}

func TestMergeRepoPolicy(t *testing.T) {
	// Test merging with allow-mode policy
	t.Run("merging allow-mode policy", func(t *testing.T) {
		localPolicy := &PolicyConfig{
			PolicyMode:     "allow",
			AllowedActions: []string{"actions/checkout", "actions/setup-node"},
			ExcludedRepos:  []string{"org/excluded-repo"},
		}

		repoConfig := `
policy_mode: allow
allowed_actions:
  - actions/special-action
custom_rules:
  org/test-repo:
    allowed_actions:
      - repo/specific-action
`

		merged, err := MergeRepoPolicy(localPolicy, []byte(repoConfig), "org/test-repo")
		if err != nil {
			t.Fatalf("MergeRepoPolicy returned error: %v", err)
		}

		// Check that original global actions are preserved
		for _, action := range localPolicy.AllowedActions {
			found := false
			for _, mergedAction := range merged.AllowedActions {
				if action == mergedAction {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected merged policy to contain action %q, but it doesn't", action)
			}
		}

		// Check that repository custom rule was applied
		repoRule, exists := merged.CustomRules["org/test-repo"]
		if !exists {
			t.Error("Expected merged policy to have custom rule for org/test-repo")
		} else {
			expectedAction := "repo/specific-action"
			found := false
			for _, action := range repoRule.AllowedActions {
				if action == expectedAction {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected custom rule to contain action %q, but it doesn't", expectedAction)
			}
		}
	})

	// Test merging with deny-mode policy
	t.Run("merging deny-mode policy", func(t *testing.T) {
		localPolicy := &PolicyConfig{
			PolicyMode:    "deny",
			DeniedActions: []string{"unsafe/action", "deprecated/action"},
			ExcludedRepos: []string{"org/excluded-repo"},
		}

		repoConfig := `
policy_mode: deny
denied_actions:
  - repo/custom-denied-action
custom_rules:
  org/test-repo:
    denied_actions:
      - repo/specific-denied-action
`

		merged, err := MergeRepoPolicy(localPolicy, []byte(repoConfig), "org/test-repo")
		if err != nil {
			t.Fatalf("MergeRepoPolicy returned error: %v", err)
		}

		// Check policy mode is preserved
		if merged.PolicyMode != "deny" {
			t.Errorf("Expected merged policy mode to be 'deny', got %q", merged.PolicyMode)
		}

		// Check that original global denied actions are preserved
		for _, action := range localPolicy.DeniedActions {
			found := false
			for _, mergedAction := range merged.DeniedActions {
				if action == mergedAction {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected merged policy to contain denied action %q, but it doesn't", action)
			}
		}

		// Check that repository custom rule was applied
		repoRule, exists := merged.CustomRules["org/test-repo"]
		if !exists {
			t.Error("Expected merged policy to have custom rule for org/test-repo")
		} else {
			expectedAction := "repo/specific-denied-action"
			found := false
			for _, action := range repoRule.DeniedActions {
				if action == expectedAction {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected custom rule to contain denied action %q, but it doesn't", expectedAction)
			}
		}
	})

	// Test merging with mixed mode policies
	t.Run("merging mixed-mode policies", func(t *testing.T) {
		localPolicy := &PolicyConfig{
			PolicyMode:     "allow",
			AllowedActions: []string{"actions/checkout", "actions/setup-node"},
		}

		repoConfig := `
policy_mode: deny
denied_actions:
  - repo/custom-denied-action
custom_rules:
  org/test-repo:
    policy_mode: deny
    denied_actions:
      - repo/specific-denied-action
`

		merged, err := MergeRepoPolicy(localPolicy, []byte(repoConfig), "org/test-repo")
		if err != nil {
			t.Fatalf("MergeRepoPolicy returned error: %v", err)
		}

		// Global policy mode should remain unchanged
		if merged.PolicyMode != "allow" {
			t.Errorf("Expected global policy mode to remain 'allow', got %q", merged.PolicyMode)
		}

		// Check that repository custom rule has correct mode
		repoRule, exists := merged.CustomRules["org/test-repo"]
		if !exists {
			t.Error("Expected merged policy to have custom rule for org/test-repo")
		} else {
			if repoRule.PolicyMode != "deny" {
				t.Errorf("Expected repo policy mode to be 'deny', got %q", repoRule.PolicyMode)
			}
		}
	})
}
