package export

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ihavespoons/action-control/internal/github"
	"github.com/ihavespoons/action-control/internal/policy"

	"gopkg.in/yaml.v3"
)

// ActionExporter handles exporting policy files based on discovered actions
type ActionExporter struct {
	// Configuration options
	IncludeVersions bool   // Whether to include version tags in exported actions
	OutputPath      string // Where to write the policy file
	IncludeCustom   bool   // Whether to include custom rules for each repository
}

// NewExporter creates a new ActionExporter with default configuration
func NewExporter() *ActionExporter {
	return &ActionExporter{
		IncludeVersions: false,
		OutputPath:      "policy.yaml",
		IncludeCustom:   false,
	}
}

// GeneratePolicyFromActions creates a policy configuration from discovered actions
func (e *ActionExporter) GeneratePolicyFromActions(actionsMap map[string][]github.Action) (*policy.PolicyConfig, error) {
	// Create a new policy config
	policyConfig := &policy.PolicyConfig{
		AllowedActions: []string{},
		ExcludedRepos:  []string{},
		CustomRules:    make(map[string]policy.Policy),
	}

	// Track unique actions
	uniqueActions := make(map[string]bool)

	// Process each repository's actions
	for repo, actions := range actionsMap {
		// If we're including custom rules, create a policy for this repo
		if e.IncludeCustom {
			repoPolicy := policy.Policy{
				AllowedActions: []string{},
			}

			for _, action := range actions {
				actionName := normalizeActionName(action.Uses, e.IncludeVersions)
				repoPolicy.AllowedActions = append(repoPolicy.AllowedActions, actionName)
				uniqueActions[actionName] = true
			}

			// Sort for consistency
			sort.Strings(repoPolicy.AllowedActions)
			policyConfig.CustomRules[repo] = repoPolicy
		} else {
			// Just add to global actions list
			for _, action := range actions {
				actionName := normalizeActionName(action.Uses, e.IncludeVersions)
				uniqueActions[actionName] = true
			}
		}
	}

	// Convert unique actions map to sorted slice
	for action := range uniqueActions {
		policyConfig.AllowedActions = append(policyConfig.AllowedActions, action)
	}
	sort.Strings(policyConfig.AllowedActions)

	return policyConfig, nil
}

// ExportPolicyFile writes the policy configuration to a YAML file
func (e *ActionExporter) ExportPolicyFile(config *policy.PolicyConfig) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(e.OutputPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Marshal the config to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal policy config: %w", err)
	}

	// Add a header comment to the YAML file
	header := `# GitHub Action Control Policy
# Generated automatically by action-control
# 
# This file defines which GitHub Actions are allowed in your repositories.
# 
# allowed_actions: Actions allowed in all repositories
# excluded_repos: Repositories excluded from policy enforcement
# custom_rules: Repository-specific action rules

`
	fileContent := header + string(data)

	// Write to file using os.WriteFile instead of deprecated ioutil.WriteFile
	if err := os.WriteFile(e.OutputPath, []byte(fileContent), 0644); err != nil {
		return fmt.Errorf("failed to write policy file: %w", err)
	}

	return nil
}

// normalizeActionName standardizes action names for policy
func normalizeActionName(action string, includeVersion bool) string {
	if !includeVersion && strings.Contains(action, "@") {
		// Remove version info if not including versions
		return action[:strings.LastIndex(action, "@")]
	}
	return action
}
