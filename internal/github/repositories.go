package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v70/github"
)

// Repository represents a GitHub repository with basic information
type Repository struct {
	Name        string
	FullName    string
	Description string
	IsPrivate   bool
}

// ListRepositories retrieves all repositories for an organization
func (c *Client) ListRepositories(ctx context.Context, org string) ([]Repository, error) {
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 100, // Adjust as needed
		},
	}

	var allRepos []Repository
	for {
		repos, resp, err := c.client.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}

		for _, repo := range repos {
			allRepos = append(allRepos, Repository{
				Name:        repo.GetName(),
				FullName:    repo.GetFullName(),
				Description: repo.GetDescription(),
				IsPrivate:   repo.GetPrivate(),
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}
