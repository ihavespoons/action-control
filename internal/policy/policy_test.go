package policy

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadPolicyConfig(t *testing.T) {
	// Create temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "policy.yaml")

	// Write test policy file
	policyContent := `
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
	expectedAllowed := []string{"actions/checkout", "actions/setup-node"}
	if !reflect.DeepEqual(policy.AllowedActions, expectedAllowed) {
		t.Errorf("Expected allowed actions %v, got %v", expectedAllowed, policy.AllowedActions)
	}

	expectedExcluded := []string{"org/test-repo"}
	if !reflect.DeepEqual(policy.ExcludedRepos, expectedExcluded) {
		t.Errorf("Expected excluded repos %v, got %v", expectedExcluded, policy.ExcludedRepos)
	}

	// Check custom rules
	customRule, exists := policy.CustomRules["org/special-repo"]
	if !exists {
		t.Error("Expected custom rule for org/special-repo, but didn't find it")
	} else {
		expectedCustomActions := []string{"actions/checkout", "custom/action"}
		if !reflect.DeepEqual(customRule.AllowedActions, expectedCustomActions) {
			t.Errorf("Expected custom actions %v, got %v", expectedCustomActions, customRule.AllowedActions)
		}
	}
}

func TestCheckActionCompliance(t *testing.T) {
	policy := &PolicyConfig{
		AllowedActions: []string{"actions/checkout", "actions/setup-node"},
		ExcludedRepos:  []string{"org/excluded-repo"},
		CustomRules: map[string]Policy{
			"org/custom-repo": {
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
}

func TestMergeRepoPolicy(t *testing.T) {
	localPolicy := &PolicyConfig{
		AllowedActions: []string{"actions/checkout", "actions/setup-node"},
		ExcludedRepos:  []string{"org/excluded-repo"},
		CustomRules: map[string]Policy{
			"org/custom-repo": {
				AllowedActions: []string{"actions/checkout", "custom/action"},
			},
		},
	}

	repoConfig := `
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
}
