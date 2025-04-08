package report

import (
	"fmt"

	"github.com/ihavespoons/action-control/internal/formatter"
)

type ActionReport struct {
	Repository string   `json:"repository"`
	Actions    []string `json:"actions"`
}

func GenerateReport(actionsData map[string][]string, format string) (string, error) {
	var report string

	switch format {
	case "json":
		var err error
		report, err = formatter.FormatJSON(actionsData)
		if err != nil {
			return "", fmt.Errorf("failed to format JSON: %w", err)
		}
	case "markdown":
		convertedData := make(map[string][]formatter.Action)
		for key, actions := range actionsData {
			var formattedActions []formatter.Action
			for _, action := range actions {
				formattedActions = append(formattedActions, formatter.Action{Name: action})
			}
			convertedData[key] = formattedActions
		}
		report = formatter.FormatMarkdown(convertedData)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	return report, nil
}
