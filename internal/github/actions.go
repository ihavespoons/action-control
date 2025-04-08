package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/go-github/v70/github"
	"gopkg.in/yaml.v3"
)

// Action represents a GitHub action reference from a workflow file
type Action struct {
	Name string
	Uses string
}

// GetActions retrieves all actions used in workflow files for a repository
func (c *Client) GetActions(ctx context.Context, owner, repo string) ([]Action, error) {
	// First, get all workflow files in .github/workflows directory
	opts := &github.RepositoryContentGetOptions{}
	_, dirContent, _, err := c.client.Repositories.GetContents(
		ctx,
		owner,
		repo,
		".github/workflows",
		opts,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get workflow directory: %w", err)
	}

	var allActions []Action

	// Process each workflow file
	for _, file := range dirContent {
		if !strings.HasSuffix(*file.Name, ".yml") && !strings.HasSuffix(*file.Name, ".yaml") {
			continue
		}

		fileContent, _, _, err := c.client.Repositories.GetContents(
			ctx,
			owner,
			repo,
			*file.Path,
			opts,
		)

		if err != nil {
			continue // Skip files we can't access
		}

		if fileContent == nil || fileContent.Content == nil {
			continue
		}

		content, err := base64.StdEncoding.DecodeString(*fileContent.Content)
		if err != nil {
			continue
		}

		actions, err := extractActionsFromWorkflow(content, *file.Name)
		if err != nil {
			continue
		}

		allActions = append(allActions, actions...)
	}

	return allActions, nil
}

// extractActionsFromWorkflow parses a workflow file and extracts action references
func extractActionsFromWorkflow(content []byte, filename string) ([]Action, error) {
	var workflow map[string]interface{}
	if err := yaml.Unmarshal(content, &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow file %s: %w", filename, err)
	}

	actions := []Action{}

	// Extract the workflow name
	workflowName, _ := workflow["name"].(string)

	// Process jobs section if it exists
	if jobs, ok := workflow["jobs"].(map[string]interface{}); ok {
		for jobName, jobConfig := range jobs {
			if jobMap, ok := jobConfig.(map[string]interface{}); ok {
				// Check for a job-level 'uses' field (e.g., for reusable workflows)
				if uses, ok := jobMap["uses"].(string); ok {
					actions = append(actions, Action{
						Name: fmt.Sprintf("%s (job: %s)", workflowName, jobName),
						Uses: uses,
					})
				}

				// Process steps if they exist
				if steps, ok := jobMap["steps"].([]interface{}); ok {
					for _, step := range steps {
						if stepMap, ok := step.(map[string]interface{}); ok {
							if uses, ok := stepMap["uses"].(string); ok {
								name := ""
								if n, ok := stepMap["name"].(string); ok {
									name = n
								}
								actions = append(actions, Action{
									Name: name,
									Uses: uses,
								})
							}
						}
					}
				}
			}
		}
	}

	return actions, nil
}
