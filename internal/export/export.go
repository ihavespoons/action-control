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
	PolicyMode      string // Which policy mode to use ("allow" or "deny")
}

// NewExporter creates a new ActionExporter with default configuration
func NewExporter() *ActionExporter {
	return &ActionExporter{
		IncludeVersions: false,
		OutputPath:      "policy.yaml",
		IncludeCustom:   false,
		PolicyMode:      "allow", // Default to allow list
	}
}

// GeneratePolicyFromActions creates a policy configuration from discovered actions
func (e *ActionExporter) GeneratePolicyFromActions(actionsMap map[string][]github.Action) (*policy.PolicyConfig, error) {
	// Create a new policy config
	policyConfig := &policy.PolicyConfig{
		PolicyMode:    e.PolicyMode,
		ExcludedRepos: []string{},
		CustomRules:   make(map[string]policy.Policy),
	}

	// Initialize the appropriate list based on policy mode
	if e.PolicyMode == "allow" {
		policyConfig.AllowedActions = []string{}
	} else if e.PolicyMode == "deny" {
		policyConfig.DeniedActions = []string{}
	} else {
		return nil, fmt.Errorf("invalid policy mode: %s, must be 'allow' or 'deny'", e.PolicyMode)
	}

	// Track unique actions
	uniqueActions := make(map[string]bool)

	// Process each repository's actions
	for repo, actions := range actionsMap {
		// If we're including custom rules, create a policy for this repo
		if e.IncludeCustom {
			repoPolicy := policy.Policy{
				PolicyMode: e.PolicyMode,
			}

			// Initialize the appropriate list based on policy mode
			if e.PolicyMode == "allow" {
				repoPolicy.AllowedActions = []string{}
			} else if e.PolicyMode == "deny" {
				repoPolicy.DeniedActions = []string{}
			}

			for _, action := range actions {
				actionName := normalizeActionName(action.Uses, e.IncludeVersions)

				// Add to repository policy
				if e.PolicyMode == "allow" {
					repoPolicy.AllowedActions = append(repoPolicy.AllowedActions, actionName)
				} else if e.PolicyMode == "deny" {
					repoPolicy.DeniedActions = append(repoPolicy.DeniedActions, actionName)
				}

				uniqueActions[actionName] = true
			}

			// Sort for consistency
			if e.PolicyMode == "allow" {
				sort.Strings(repoPolicy.AllowedActions)
			} else if e.PolicyMode == "deny" {
				sort.Strings(repoPolicy.DeniedActions)
			}

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
	uniqueActionsList := make([]string, 0, len(uniqueActions))
	for action := range uniqueActions {
		uniqueActionsList = append(uniqueActionsList, action)
	}
	sort.Strings(uniqueActionsList)

	// Add to appropriate list in policy config
	if e.PolicyMode == "allow" {
		policyConfig.AllowedActions = uniqueActionsList
	} else if e.PolicyMode == "deny" {
		policyConfig.DeniedActions = uniqueActionsList
	}

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
# This file defines policy for GitHub Actions in your repositories.
# 
# Policy can work in two modes:
#   - allow: Only listed actions are allowed (default)
#   - deny: All actions are allowed except listed ones
# 
# allowed_actions: Actions explicitly allowed (used in allow mode)
# denied_actions: Actions explicitly denied (used in deny mode)
# policy_mode: Which mode to use ("allow" or "deny")
# excluded_repos: Repositories excluded from policy enforcement
# custom_rules: Repository-specific action rules

`
	fileContent := header + string(data)

	// Write to file using os.WriteFile
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
