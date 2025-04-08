package policy

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// PolicyConfig defines the structure for the policy configuration file
type PolicyConfig struct {
	AllowedActions []string          `yaml:"allowed_actions,omitempty"`
	DeniedActions  []string          `yaml:"denied_actions,omitempty"`
	ExcludedRepos  []string          `yaml:"excluded_repos,omitempty"`
	CustomRules    map[string]Policy `yaml:"custom_rules,omitempty"`
	PolicyMode     string            `yaml:"policy_mode,omitempty"` // "allow" or "deny"
}

// Policy defines repository-specific policy
type Policy struct {
	AllowedActions []string `yaml:"allowed_actions,omitempty"`
	DeniedActions  []string `yaml:"denied_actions,omitempty"`
	PolicyMode     string   `yaml:"policy_mode,omitempty"` // "allow" or "deny"
}

// LoadPolicyConfig loads policy configuration from the specified file
func LoadPolicyConfig(configPath string) (*PolicyConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy config: %w", err)
	}

	var config PolicyConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse policy config: %w", err)
	}

	// Set default policy mode if not specified
	if config.PolicyMode == "" {
		if len(config.AllowedActions) > 0 {
			config.PolicyMode = "allow"
		} else if len(config.DeniedActions) > 0 {
			config.PolicyMode = "deny"
		} else {
			config.PolicyMode = "allow" // Default to allow mode if neither is specified
		}
	}

	return &config, nil
}

// MergeRepoPolicy merges repository-specific policy with global policy
func MergeRepoPolicy(globalPolicy *PolicyConfig, repoPolicyContent []byte, repoName string) (*PolicyConfig, error) {
	// Create a deep copy of the global policy
	mergedPolicy := &PolicyConfig{
		AllowedActions: make([]string, len(globalPolicy.AllowedActions)),
		DeniedActions:  make([]string, len(globalPolicy.DeniedActions)),
		ExcludedRepos:  make([]string, len(globalPolicy.ExcludedRepos)),
		CustomRules:    make(map[string]Policy),
		PolicyMode:     globalPolicy.PolicyMode,
	}

	// Copy slices and map
	copy(mergedPolicy.AllowedActions, globalPolicy.AllowedActions)
	copy(mergedPolicy.DeniedActions, globalPolicy.DeniedActions)
	copy(mergedPolicy.ExcludedRepos, globalPolicy.ExcludedRepos)
	for k, v := range globalPolicy.CustomRules {
		mergedPolicy.CustomRules[k] = v
	}

	// Parse repo policy
	var repoPolicy PolicyConfig
	if err := yaml.Unmarshal(repoPolicyContent, &repoPolicy); err != nil {
		return nil, fmt.Errorf("failed to parse repository policy: %w", err)
	}

	// Apply repo-specific overrides if provided
	customRule, exists := repoPolicy.CustomRules[repoName]
	if exists {
		mergedPolicy.CustomRules[repoName] = customRule
	} else if len(repoPolicy.AllowedActions) > 0 || len(repoPolicy.DeniedActions) > 0 {
		// If repo doesn't have a specific custom rule but has global actions,
		// create a custom rule for it
		policy := Policy{
			AllowedActions: repoPolicy.AllowedActions,
			DeniedActions:  repoPolicy.DeniedActions,
		}

		// Set policy mode for the repo if specified, otherwise inherit
		if repoPolicy.PolicyMode != "" {
			policy.PolicyMode = repoPolicy.PolicyMode
		} else {
			policy.PolicyMode = determineRepoMode(policy, globalPolicy.PolicyMode)
		}

		mergedPolicy.CustomRules[repoName] = policy
	}

	return mergedPolicy, nil
}

// determineRepoMode figures out the appropriate policy mode for a repository
func determineRepoMode(policy Policy, defaultMode string) string {
	if policy.PolicyMode != "" {
		return policy.PolicyMode
	}

	if len(policy.AllowedActions) > 0 {
		return "allow"
	} else if len(policy.DeniedActions) > 0 {
		return "deny"
	}

	return defaultMode
}

// CheckActionCompliance verifies that all actions comply with the policy
func CheckActionCompliance(policy *PolicyConfig, repoName string, actions []string) ([]string, bool) {
	// Check if repository is excluded from policy
	for _, excludedRepo := range policy.ExcludedRepos {
		if excludedRepo == repoName {
			return nil, true // Repository is excluded, so it's compliant
		}
	}

	// Determine which policy to apply (global or custom)
	var allowedActions, deniedActions []string
	var policyMode string

	if customPolicy, exists := policy.CustomRules[repoName]; exists {
		// Use custom policy for this repository
		allowedActions = customPolicy.AllowedActions
		deniedActions = customPolicy.DeniedActions
		policyMode = customPolicy.PolicyMode

		// If custom policy mode is not specified, inherit from global
		if policyMode == "" {
			policyMode = policy.PolicyMode
		}

		// If custom policy doesn't specify actions for its mode, inherit from global
		if policyMode == "allow" && len(allowedActions) == 0 {
			allowedActions = policy.AllowedActions
		} else if policyMode == "deny" && len(deniedActions) == 0 {
			deniedActions = policy.DeniedActions
		}
	} else {
		// Use global policy
		allowedActions = policy.AllowedActions
		deniedActions = policy.DeniedActions
		policyMode = policy.PolicyMode
	}

	// Default to allow mode if not specified
	if policyMode == "" {
		if len(allowedActions) > 0 {
			policyMode = "allow"
		} else if len(deniedActions) > 0 {
			policyMode = "deny"
		} else {
			policyMode = "allow" // Default fallback
		}
	}

	// Check actions against policy
	var violations []string

	// Normalize actions by removing version info for policy checking
	for _, actionWithVersion := range actions {
		action := normalizeAction(actionWithVersion)

		if policyMode == "allow" {
			// In allow mode, action must be in the allowed list
			if !contains(allowedActions, action) && !contains(allowedActions, actionWithVersion) {
				violations = append(violations, actionWithVersion)
			}
		} else if policyMode == "deny" {
			// In deny mode, action must NOT be in the denied list
			if contains(deniedActions, action) || contains(deniedActions, actionWithVersion) {
				violations = append(violations, actionWithVersion)
			}
		}
	}

	return violations, len(violations) == 0
}

// normalizeAction removes version info from action string
func normalizeAction(action string) string {
	for i := 0; i < len(action); i++ {
		if action[i] == '@' {
			return action[:i]
		}
	}
	return action
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
