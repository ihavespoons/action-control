package formatter

import (
	"fmt"
	"sort"
	"strings"
)

// Update the FormatPolicyViolations function to mention the policy mode
func FormatPolicyViolations(violations map[string][]string, policyMode string) string {
	if len(violations) == 0 {
		return "✅ All repositories comply with the action policy."
	}

	var sb strings.Builder
	sb.WriteString("# Policy Violation Report\n\n")

	if policyMode == "deny" {
		sb.WriteString("## ❌ Denied Actions Found\n\n")
	} else {
		sb.WriteString("## ❌ Policy Violations\n\n")
	}

	// Sort repositories for consistent output
	repos := make([]string, 0, len(violations))
	for repo := range violations {
		repos = append(repos, repo)
	}
	sort.Strings(repos)

	for _, repo := range repos {
		sb.WriteString(fmt.Sprintf("### %s\n\n", repo))

		if policyMode == "deny" {
			sb.WriteString("The following denied actions were found:\n\n")
		} else {
			sb.WriteString("The following actions are not allowed by policy:\n\n")
		}

		for _, action := range violations[repo] {
			sb.WriteString(fmt.Sprintf("- `%s`\n", action))
		}
		sb.WriteString("\n")
	}

	if policyMode == "deny" {
		sb.WriteString(fmt.Sprintf("\nFound %d repositories using denied actions.\n", len(violations)))
	} else {
		sb.WriteString(fmt.Sprintf("\nFound %d repositories with policy violations.\n", len(violations)))
	}

	return sb.String()
}
