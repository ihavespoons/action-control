package github

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-github/v70/github"
)

// MockServer creates a test HTTP server that returns predefined responses
func MockServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	// Create a test server with the provided handler
	server := httptest.NewServer(handler)

	// Create a GitHub client
	httpClient := &http.Client{}

	// Create a new GitHub API client
	githubClient := github.NewClient(httpClient)

	// Parse the mock server URL
	serverURL, _ := url.Parse(server.URL + "/")

	// Override the base URL to point to our mock server
	githubClient.BaseURL = serverURL
	githubClient.UploadURL = serverURL

	// Create our client wrapper around the GitHub client
	client := &Client{
		client: githubClient,
		token:  "mock-token",
	}

	return server, client
}

// EncodeContent base64 encodes content for use in mocked GitHub API responses
func EncodeContent(content string) string {
	return base64.StdEncoding.EncodeToString([]byte(content))
}

// CreateMockWorkflowContent creates realistic workflow YAML content for testing
func CreateMockWorkflowContent() string {
	return `
name: Test Workflow
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Node
        uses: actions/setup-node@v2
        with:
          node-version: '16'
      - name: Run Tests
        run: npm test
`
}

// CreateMockRepositoriesResponse creates a JSON response mimicking the GitHub API repositories endpoint
func CreateMockRepositoriesResponse(repos []Repository) string {
	// Create a properly formatted JSON array
	var responseItems []string

	for _, repo := range repos {
		repoJSON := fmt.Sprintf(`{
            "name": "%s",
            "full_name": "%s",
            "description": "%s",
            "private": %v
        }`, repo.Name, repo.FullName, repo.Description, repo.IsPrivate)
		responseItems = append(responseItems, repoJSON)
	}

	// Join all repository JSON objects with commas
	return "[" + strings.Join(responseItems, ",") + "]"
}
