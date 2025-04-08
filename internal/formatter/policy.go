package formatter

import (
	"fmt"
	"strings"
)

// FormatPolicyViolations formats policy violations as a readable report
func FormatPolicyViolations(violations map[string][]string) string {
	var builder strings.Builder

	builder.WriteString("# Policy Violation Report\n\n")

	if len(violations) == 0 {
		builder.WriteString("✅ All repositories comply with the action policy.\n")
		return builder.String()
	}

	builder.WriteString("## ❌ Policy Violations\n\n")
	builder.WriteString("The following repositories contain actions that violate the policy:\n\n")

	for repo, actions := range violations {
		builder.WriteString(fmt.Sprintf("### %s\n\n", repo))
		builder.WriteString("Disallowed actions:\n\n")
		for _, action := range actions {
			builder.WriteString(fmt.Sprintf("- `%s`\n", action))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("## Summary\n\n")
	builder.WriteString(fmt.Sprintf("Found %d repositories with policy violations.\n", len(violations)))

	return builder.String()
}
