package formatter

import (
	"encoding/json"
	"fmt"
)

// FormatJSON takes a report data structure and returns a JSON string representation of it.
func FormatJSON(data interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling to JSON: %w", err)
	}
	return string(jsonData), nil
}
