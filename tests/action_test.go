package tests

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestActionYaml validates that the GitHub Action YAML file is correctly formatted
func TestActionYaml(t *testing.T) {
	// Find the action.yml file
	actionYamlPath := "../action.yml"
	if _, err := os.Stat(actionYamlPath); os.IsNotExist(err) {
		t.Fatalf("action.yml file not found at %s", actionYamlPath)
	}

	// Read the YAML file
	data, err := os.ReadFile(actionYamlPath)
	if err != nil {
		t.Fatalf("Failed to read action.yml: %v", err)
	}

	// Parse the YAML
	var actionConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &actionConfig); err != nil {
		t.Fatalf("Failed to parse action.yml: %v", err)
	}

	// Check required fields
	requiredFields := []string{"name", "description", "inputs", "runs"}
	for _, field := range requiredFields {
		if _, exists := actionConfig[field]; !exists {
			t.Errorf("Required field %q missing from action.yml", field)
		}
	}

	// Check inputs
	inputs, ok := actionConfig["inputs"].(map[string]interface{})
	if !ok {
		t.Fatal("Inputs section is not properly formatted")
	}

	// Check required inputs - github_token is required, command is not in your actual file
	requiredInputs := []string{"github_token", "organization", "repository", "output_format", "policy_file"}
	for _, input := range requiredInputs {
		inputConfig, exists := inputs[input].(map[string]interface{})
		if !exists {
			t.Errorf("Input %q missing from action.yml", input)
			continue
		}

		if _, exists := inputConfig["description"]; !exists {
			t.Errorf("Input %q is missing a description", input)
		}
	}

	// Check that github_token is required
	if githubToken, exists := inputs["github_token"].(map[string]interface{}); exists {
		required, ok := githubToken["required"].(bool)
		if !ok || !required {
			t.Error("github_token input should be marked as required")
		}
	}

	// Check runs section
	runs, ok := actionConfig["runs"].(map[string]interface{})
	if !ok {
		t.Fatal("Runs section is not properly formatted")
	}

	if using, ok := runs["using"].(string); !ok || using != "docker" {
		t.Errorf("Expected runs.using to be 'docker', got %v", using)
	}

	if _, exists := runs["image"]; !exists {
		t.Error("Runs section is missing 'image' field")
	}

	// Check that args includes the enforce command
	args, ok := runs["args"].([]interface{})
	if !ok {
		t.Error("Runs section is missing 'args' field or it's not an array")
	} else {
		if len(args) == 0 || args[0] != "enforce" {
			t.Error("The first arg should be 'enforce' command")
		}

		// Check that organization flag is included
		hasOrgFlag := false
		for i, arg := range args {
			if arg == "--org" && i+1 < len(args) {
				hasOrgFlag = true
				break
			}
		}
		if !hasOrgFlag {
			t.Error("Args should include '--org' flag")
		}
	}

	// Check that environment variables are set
	env, ok := actionConfig["env"].(map[string]interface{})
	if !ok {
		t.Error("Action should have environment variables defined")
	} else {
		if _, exists := env["ACTION_CONTROL_GITHUB_TOKEN"]; !exists {
			t.Error("ACTION_CONTROL_GITHUB_TOKEN environment variable should be defined")
		}
	}
}

// TestDockerfile validates that the Dockerfile exists
func TestDockerfile(t *testing.T) {
	dockerfilePath := "../Dockerfile"
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		t.Fatalf("Dockerfile not found at %s", dockerfilePath)
	}

	// Read Dockerfile
	data, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("Failed to read Dockerfile: %v", err)
	}

	content := string(data)

	// Check for essential components based on current Dockerfile
	essentialPhrases := []string{
		"FROM alpine:",
		"WORKDIR /app",
		"COPY",
		"RUN apk --no-cache add ca-certificates",
		"ENTRYPOINT",
	}

	for _, phrase := range essentialPhrases {
		if !strings.Contains(content, phrase) {
			t.Errorf("Dockerfile should contain %q", phrase)
		}
	}

	// Check for entrypoint script handling
	if !strings.Contains(content, "entrypoint.sh") {
		t.Error("Dockerfile should reference entrypoint.sh")
	}

	// Check for build architecture handling
	if !strings.Contains(content, "BUILDARCH") {
		t.Error("Dockerfile should use BUILDARCH argument")
	}
}

// Add test case to verify policy file handling in GitHub Action
func TestActionProcessesPolicyFile(t *testing.T) {
	// Find the action.yml file
	actionYamlPath := "../action.yml"
	if _, err := os.Stat(actionYamlPath); os.IsNotExist(err) {
		t.Fatalf("action.yml file not found at %s", actionYamlPath)
	}

	// Read the YAML file
	data, err := os.ReadFile(actionYamlPath)
	if err != nil {
		t.Fatalf("Failed to read action.yml: %v", err)
	}

	// Parse the YAML
	var actionConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &actionConfig); err != nil {
		t.Fatalf("Failed to parse action.yml: %v", err)
	}

	// Check the runs section for proper policy file handling
	runs, ok := actionConfig["runs"].(map[string]interface{})
	if !ok {
		t.Fatal("Runs section is not properly formatted")
	}

	args, ok := runs["args"].([]interface{})
	if !ok {
		t.Fatal("Args section is not properly formatted")
	}

	// Check for policy flag
	hasPolicyFlag := false
	for i, arg := range args {
		if arg == "--policy" && i+1 < len(args) {
			hasPolicyFlag = true

			// Verify that the policy file is correctly configured
			policyRef, ok := args[i+1].(string)
			if !ok || !strings.Contains(policyRef, "policy_file") {
				t.Errorf("Policy file reference is not properly set up, got: %v", args[i+1])
			}
			break
		}
	}

	if !hasPolicyFlag {
		t.Error("GitHub Action args should include --policy flag")
	}
}
