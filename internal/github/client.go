package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v70/github"
	"golang.org/x/oauth2"
)

// Client provides access to GitHub API
type Client struct {
	client *github.Client
	token  string
}

// NewClient creates a new GitHub client with the provided token
func NewClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client: github.NewClient(tc),
		token:  token,
	}
}

// GetRepositoryContent retrieves file content from a repository
func (c *Client) GetRepositoryContent(ctx context.Context, owner, repo, path string) ([]byte, error) {
	fileContent, _, resp, err := c.client.Repositories.GetContents(
		ctx,
		owner,
		repo,
		path,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get file content for %s: %w", path, err)
	}

	// Check if the file exists
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Return empty content if file doesn't exist
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if fileContent == nil || fileContent.Content == nil {
		return nil, fmt.Errorf("empty file content")
	}

	content, err := base64.StdEncoding.DecodeString(*fileContent.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to decode content: %w", err)
	}

	return content, nil
}

// ActionsForOrg retrieves all actions used across an organization's repositories
func (c *Client) ActionsForOrg(ctx context.Context, org string) (map[string][]Action, error) {
	repos, err := c.ListRepositories(ctx, org)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]Action)

	for _, repo := range repos {
		parts := strings.Split(repo.FullName, "/")
		if len(parts) != 2 {
			continue
		}

		owner := parts[0]
		repoName := parts[1]

		actions, err := c.GetActions(ctx, owner, repoName)
		if err != nil {
			// Log error but continue with other repositories
			continue
		}

		result[repo.FullName] = actions
	}

	return result, nil
}
