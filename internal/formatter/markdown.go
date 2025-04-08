package formatter

import (
	"fmt"
	"sort"
	"strings"
)

// Action represents a GitHub action usage in a repository
type Action struct {
	Name string
	Uses string
}

// FormatMarkdown formats the actions data as a Markdown document
func FormatMarkdown(data map[string][]Action) string {
	var builder strings.Builder

	builder.WriteString("# GitHub Actions Usage Report\n\n")

	// Sort repositories for consistent output
	repos := make([]string, 0, len(data))
	for repo := range data {
		repos = append(repos, repo)
	}
	sort.Strings(repos)

	// Track unique actions
	uniqueActions := make(map[string]int)

	// Generate report by repository
	builder.WriteString("## Actions by Repository\n\n")
	for _, repo := range repos {
		actions := data[repo]
		if len(actions) == 0 {
			continue
		}

		builder.WriteString(fmt.Sprintf("### %s\n\n", repo))
		builder.WriteString("| Action Name | Action Reference |\n")
		builder.WriteString("|------------|------------------|\n")

		for _, action := range actions {
			// Count unique actions
			uniqueActions[action.Uses]++

			name := action.Name
			if name == "" {
				name = "_Unnamed_"
			}
			builder.WriteString(fmt.Sprintf("| %s | `%s` |\n", name, action.Uses))
		}
		builder.WriteString("\n")
	}

	// Generate summary of most used actions
	builder.WriteString("## Most Used Actions\n\n")
	builder.WriteString("| Action | Usage Count |\n")
	builder.WriteString("|--------|------------|\n")

	// Convert map to slice for sorting
	type actionUsage struct {
		action string
		count  int
	}

	usages := make([]actionUsage, 0, len(uniqueActions))
	for action, count := range uniqueActions {
		usages = append(usages, actionUsage{action, count})
	}

	// Sort by count (descending)
	sort.Slice(usages, func(i, j int) bool {
		return usages[i].count > usages[j].count
	})

	// Print top actions (limit to 20)
	limit := 20
	if len(usages) < limit {
		limit = len(usages)
	}

	for i := 0; i < limit; i++ {
		builder.WriteString(fmt.Sprintf("| `%s` | %d |\n", usages[i].action, usages[i].count))
	}

	return builder.String()
}
