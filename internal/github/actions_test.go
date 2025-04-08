package github

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestGetActions(t *testing.T) {
	// Setup mock response for a repository with workflows
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Debug the received path
		t.Logf("Mock server received request to path: %s", r.URL.Path)

		// Handle request for workflow directory listing
		if r.URL.Path == "/repos/owner/repo/contents/.github/workflows" {
			fmt.Fprint(w, `[
                {
                    "name": "ci.yml",
                    "path": ".github/workflows/ci.yml",
                    "type": "file"
                }
            ]`)
			return
		}

		// Handle request for specific workflow file
		if r.URL.Path == "/repos/owner/repo/contents/.github/workflows/ci.yml" {
			responseContent := fmt.Sprintf(`{
                "name": "ci.yml",
                "path": ".github/workflows/ci.yml",
                "content": "%s"
            }`, EncodeContent(CreateMockWorkflowContent()))
			fmt.Fprint(w, responseContent)
			return
		}

		// Return 404 for unknown paths
		t.Logf("No mock handler for path: %s, returning 404", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	})

	server, client := MockServer(t, mockHandler)
	defer server.Close()

	actions, err := client.GetActions(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("GetActions returned error: %v", err)
	}

	// Verify we found the expected actions
	if len(actions) != 2 {
		t.Fatalf("Expected 2 actions, got %d", len(actions))
	}

	// Verify that we extract the expected actions
	expectedActions := []string{"actions/checkout@v3", "actions/setup-node@v2"}
	for i, action := range actions {
		if action.Uses != expectedActions[i] {
			t.Errorf("Expected action %q, got %q", expectedActions[i], action.Uses)
		}
	}
}

func TestExtractActionsFromWorkflow(t *testing.T) {
	workflowYaml := CreateMockWorkflowContent()

	actions, err := extractActionsFromWorkflow([]byte(workflowYaml), "test-workflow.yml")
	if err != nil {
		t.Fatalf("extractActionsFromWorkflow returned error: %v", err)
	}

	if len(actions) != 2 {
		t.Fatalf("Expected 2 actions, got %d", len(actions))
	}

	// Check first action
	if actions[0].Uses != "actions/checkout@v3" {
		t.Errorf("Expected first action to be 'actions/checkout@v3', got %q", actions[0].Uses)
	}

	// Check second action
	if actions[1].Uses != "actions/setup-node@v2" {
		t.Errorf("Expected second action to be 'actions/setup-node@v2', got %q", actions[1].Uses)
	}

	// Check names
	if actions[1].Name != "Setup Node" {
		t.Errorf("Expected action name to be 'Setup Node', got %q", actions[1].Name)
	}
}
