package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLIExecution tests that the CLI can be executed with various commands
// Note: These tests require the binary to be built first
func TestCLIExecution(t *testing.T) {
	// Skip integration tests if SKIP_INTEGRATION environment variable is set
	if os.Getenv("SKIP_INTEGRATION") != "" {
		t.Skip("Skipping integration tests")
	}

	// Find binary location - assumes you've built it with "go build"
	binPath, err := findBinary()
	if err != nil {
		t.Fatalf("Failed to find binary: %v", err)
	}

	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a simple policy file
	policyPath := filepath.Join(tempDir, "test-policy.yaml")
	policyContent := `
allowed_actions:
  - actions/checkout
  - actions/setup-node
`
	if err := os.WriteFile(policyPath, []byte(policyContent), 0644); err != nil {
		t.Fatalf("Failed to create test policy file: %v", err)
	}

	// Test the help command
	t.Run("help command", func(t *testing.T) {
		cmd := exec.Command(binPath, "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Command failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "A CLI tool to report and enforce GitHub Actions usage policies") {
			t.Errorf("Expected help output to contain tool description")
		}
	})

	// Test report command help
	t.Run("report command help", func(t *testing.T) {
		cmd := exec.Command(binPath, "report", "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Command failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Report on GitHub Actions") {
			t.Errorf("Expected help output to contain report description")
		}
	})

	// Test enforce command help
	t.Run("enforce command help", func(t *testing.T) {
		cmd := exec.Command(binPath, "enforce", "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Command failed: %v\nOutput: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Enforce policy") {
			t.Errorf("Expected help output to contain enforce description")
		}
	})

	// Note: We don't test API calls here, as that would require a GitHub token
	// and would make network calls, which isn't ideal for automated testing.
	// The unit tests with mocks cover this functionality.
}

// findBinary attempts to locate the action-control binary
func findBinary() (string, error) {
	// Try common locations
	locations := []string{
		"../action-control",
		"../bin/action-control",
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			absPath, err := filepath.Abs(loc)
			if err != nil {
				return "", err
			}
			return absPath, nil
		}
	}

	// If not found in standard locations, try building it
	cmd := exec.Command("go", "build", "-o", "action-control")
	cmd.Dir = ".."
	if err := cmd.Run(); err == nil {
		return filepath.Abs("../action-control")
	}

	return "", fmt.Errorf("could not find or build action-control binary")
}
