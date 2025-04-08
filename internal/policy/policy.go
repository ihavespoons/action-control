package policy

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// PolicyConfig represents the configuration for GitHub action policies
type PolicyConfig struct {
	AllowedActions []string          `yaml:"allowed_actions"`
	ExcludedRepos  []string          `yaml:"excluded_repos"`
	CustomRules    map[string]Policy `yaml:"custom_rules"` // Repository-specific rules
}

// Policy represents a specific policy for a repository
type Policy struct {
	AllowedActions []string `yaml:"allowed_actions"`
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

	return &config, nil
}

// MergeRepoPolicy merges the local policy with repo-specific policy if available
func MergeRepoPolicy(localPolicy *PolicyConfig, repoConfigData []byte, repoName string) (*PolicyConfig, error) {
	// Start with a copy of the local policy
	mergedPolicy := &PolicyConfig{
		AllowedActions: make([]string, len(localPolicy.AllowedActions)),
		ExcludedRepos:  make([]string, len(localPolicy.ExcludedRepos)),
		CustomRules:    make(map[string]Policy),
	}

	copy(mergedPolicy.AllowedActions, localPolicy.AllowedActions)
	copy(mergedPolicy.ExcludedRepos, localPolicy.ExcludedRepos)
	for k, v := range localPolicy.CustomRules {
		mergedPolicy.CustomRules[k] = v
	}

	// If we have repo-specific config data, parse and merge it
	if len(repoConfigData) > 0 {
		var repoPolicy PolicyConfig
		if err := yaml.Unmarshal(repoConfigData, &repoPolicy); err != nil {
			return nil, fmt.Errorf("failed to parse repo policy: %w", err)
		}

		// If this repo has specific rules in the repo policy, add them to the merged policy
		if repoRule, exists := repoPolicy.CustomRules[repoName]; exists {
			mergedPolicy.CustomRules[repoName] = repoRule
		}
	}

	return mergedPolicy, nil
}

// CheckActionCompliance checks if all the actions in a repository comply with policy
func CheckActionCompliance(policy *PolicyConfig, repoName string, actions []string) ([]string, bool) {
	// Check if repo is excluded from policy enforcement
	for _, excludedRepo := range policy.ExcludedRepos {
		if strings.EqualFold(repoName, excludedRepo) {
			return nil, true
		}
	}

	// Determine which list of allowed actions to use
	allowedActions := policy.AllowedActions

	// If there's a custom policy for this repo, use that instead
	if customPolicy, exists := policy.CustomRules[repoName]; exists {
		allowedActions = customPolicy.AllowedActions
	}

	// Convert allowed actions to a map for faster lookups
	allowedMap := make(map[string]bool)
	for _, action := range allowedActions {
		allowedMap[normalizeActionName(action)] = true
	}

	// Check each action against the allowed list
	var violations []string
	for _, action := range actions {
		normalized := normalizeActionName(action)
		if !allowedMap[normalized] {
			violations = append(violations, action)
		}
	}

	return violations, len(violations) == 0
}

// normalizeActionName standardizes action names for comparison
func normalizeActionName(action string) string {
	// Remove version/tag if present
	if idx := strings.LastIndex(action, "@"); idx != -1 {
		action = action[:idx]
	}
	return strings.ToLower(action)
}

// GenerateEmptyConfig creates an empty policy configuration
func GenerateEmptyConfig() *PolicyConfig {
	return &PolicyConfig{
		AllowedActions: []string{},
		ExcludedRepos:  []string{},
		CustomRules:    make(map[string]Policy),
	}
}
