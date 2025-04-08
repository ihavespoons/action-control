package github

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestListRepositories(t *testing.T) {
	// Test case with successful response
	t.Run("successful response", func(t *testing.T) {
		mockRepos := []Repository{
			{Name: "repo1", FullName: "org/repo1", Description: "Test repo 1", IsPrivate: false},
			{Name: "repo2", FullName: "org/repo2", Description: "Test repo 2", IsPrivate: true},
		}

		mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/orgs/test-org/repos" {
				t.Errorf("Expected path to be /orgs/test-org/repos, got %s", r.URL.Path)
			}

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, CreateMockRepositoriesResponse(mockRepos))
		})

		server, client := MockServer(t, mockHandler)
		defer server.Close()

		repos, err := client.ListRepositories(context.Background(), "test-org")
		if err != nil {
			t.Fatalf("ListRepositories returned error: %v", err)
		}

		if len(repos) != len(mockRepos) {
			t.Fatalf("Expected %d repos, got %d", len(mockRepos), len(repos))
		}

		for i, repo := range repos {
			expected := mockRepos[i]
			if repo.Name != expected.Name || repo.FullName != expected.FullName {
				t.Errorf("Repo #%d: expected %+v, got %+v", i, expected, repo)
			}
		}
	})

	// Test case with error response
	t.Run("error response", func(t *testing.T) {
		mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message": "Not Found"}`)
		})

		server, client := MockServer(t, mockHandler)
		defer server.Close()

		_, err := client.ListRepositories(context.Background(), "non-existent-org")
		if err == nil {
			t.Fatal("Expected error but got nil")
		}
	})
}
