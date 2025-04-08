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

	// Updated required inputs - removed organization input
	requiredInputs := []string{"github_token", "output_format", "policy_content"}
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

	// Check that github_token and policy_content are required
	requiredParams := []string{"github_token", "policy_content"}
	for _, param := range requiredParams {
		if inputConfig, exists := inputs[param].(map[string]interface{}); exists {
			required, ok := inputConfig["required"].(bool)
			if !ok || !required {
				t.Errorf("%s input should be marked as required", param)
			}
		}
	}

	// Check that policy_content has the correct description
	if policyContent, exists := inputs["policy_content"].(map[string]interface{}); exists {
		description, ok := policyContent["description"].(string)
		if !ok || !strings.Contains(description, "ignoring local policy files") {
			t.Error("policy_content description should mention it ignores local policy files")
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

		// Check that repository flag is included with github.repository
		hasRepoFlag := false
		for i, arg := range args {
			if arg == "--repo" && i+1 < len(args) {
				if args[i+1] == "${{ github.repository }}" {
					hasRepoFlag = true
					break
				}
			}
		}
		if !hasRepoFlag {
			t.Error("Args should include '--repo' flag with github.repository variable")
		}
	}

	// Check that environment variables are set correctly
	env, ok := actionConfig["env"].(map[string]interface{})
	if !ok {
		t.Error("Action should have environment variables defined")
	} else {
		// Check for required environment variables
		requiredEnvVars := []string{
			"ACTION_CONTROL_GITHUB_TOKEN",
			"ACTION_CONTROL_POLICY_CONTENT",
		}

		for _, envVar := requiredEnvVars {
			if _, exists := env[envVar]; !exists {
				t.Errorf("%s environment variable should be defined", envVar)
			}
		}
	}
}

// TestDockerfile validates that the Dockerfile exists and has correct configuration
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
	if !strings.Contains(content, "TARGETARCH") {
		t.Error("Dockerfile should use TARGETARCH argument")
	}
}

// Test the entrypoint.sh script
func TestEntrypointScript(t *testing.T) {
	// Check entrypoint script existence
	entrypointPath := "../entrypoint.sh"
	if _, err := os.Stat(entrypointPath); os.IsNotExist(err) {
		t.Fatalf("entrypoint.sh file not found at %s", entrypointPath)
	}

	// Read entrypoint script
	data, err := os.ReadFile(entrypointPath)
	if err != nil {
		t.Fatalf("Failed to read entrypoint.sh: %v", err)
	}

	content := string(data)

	// Essential elements that should be in the script
	essentialPhrases := []string{
		"ACTION_CONTROL_POLICY_CONTENT",
		"IGNORE_FLAG",
		"--ignore-local-policy",
		"exec /app/action-control",
	}

	for _, phrase := range essentialPhrases {
		if !strings.Contains(content, phrase) {
			t.Errorf("entrypoint.sh should contain %q", phrase)
		}
	}

	// Make sure script handles the policy content environment variable properly
	if !strings.Contains(content, "if [ -n \"$ACTION_CONTROL_POLICY_CONTENT\" ]") {
		t.Error("entrypoint.sh should check for ACTION_CONTROL_POLICY_CONTENT")
	}
}

// TestEnvironmentVariableHandling verifies that the action uses environment variables correctly
func TestEnvironmentVariableHandling(t *testing.T) {
	// Find the action.yml file
	actionYamlPath := "../action.yml"
	data, err := os.ReadFile(actionYamlPath)
	if err != nil {
		t.Fatalf("Failed to read action.yml: %v", err)
	}

	var actionConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &actionConfig); err != nil {
		t.Fatalf("Failed to parse action.yml: %v", err)
	}

	// Check that policy_content input is mapped to the environment variable
	env, ok := actionConfig["env"].(map[string]interface{})
	if !ok {
		t.Fatal("Action env section is missing or not properly formatted")
	}

	policyContentEnv, exists := env["ACTION_CONTROL_POLICY_CONTENT"].(string)
	if !exists {
		t.Fatal("ACTION_CONTROL_POLICY_CONTENT environment variable should be defined")
	}

	// Check that it references the policy_content input
	if !strings.Contains(policyContentEnv, "inputs.policy_content") {
		t.Errorf("ACTION_CONTROL_POLICY_CONTENT should reference inputs.policy_content, got: %s", policyContentEnv)
	}

	// Check that entrypoint script exists and has the correct behavior
	scriptPath := "../entrypoint.sh"
	scriptData, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("Failed to read entrypoint.sh: %v", err)
	}

	scriptContent := string(scriptData)

	// Make sure the script exits if no policy content is provided
	if !strings.Contains(scriptContent, "exit 1") {
		t.Error("entrypoint.sh should exit with error if ACTION_CONTROL_POLICY_CONTENT is not provided")
	}
}
