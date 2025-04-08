package tests

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Test constants
const (
	actionYamlPath = "../action.yml"
	dockerfilePath = "../Dockerfile"
	entrypointPath = "../entrypoint.sh"
)

// TestActionControl runs all tests for the GitHub Action
func TestActionControl(t *testing.T) {
	t.Run("ActionYamlStructure", testActionYamlStructure)
	t.Run("ActionInputs", testActionInputs)
	t.Run("ActionRuns", testActionRuns)
	t.Run("DockerfileConfig", testDockerfileConfig)
	t.Run("EntrypointScript", testEntrypointScript)
	t.Run("CommandLineHandling", testCommandLineHandling)
}

// Helper function to read and parse the action.yml file
func readActionYaml(t *testing.T) map[string]interface{} {
	t.Helper()

	if _, err := os.Stat(actionYamlPath); os.IsNotExist(err) {
		t.Fatalf("action.yml file not found at %s", actionYamlPath)
	}

	data, err := os.ReadFile(actionYamlPath)
	if err != nil {
		t.Fatalf("Failed to read action.yml: %v", err)
	}

	var actionConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &actionConfig); err != nil {
		t.Fatalf("Failed to parse action.yml: %v", err)
	}

	return actionConfig
}

// testActionYamlStructure verifies the basic structure of the action.yml file
func testActionYamlStructure(t *testing.T) {
	actionConfig := readActionYaml(t)

	// Check required fields
	requiredFields := []string{"name", "description", "inputs", "runs"}
	for _, field := range requiredFields {
		if _, exists := actionConfig[field]; !exists {
			t.Errorf("Required field %q missing from action.yml", field)
		}
	}

	// Check branding
	branding, ok := actionConfig["branding"].(map[string]interface{})
	if !ok {
		t.Error("Branding section is missing")
	} else {
		icon, iconOk := branding["icon"].(string)
		color, colorOk := branding["color"].(string)

		if !iconOk || icon == "" {
			t.Error("Branding icon is missing or empty")
		}

		if !colorOk || color == "" {
			t.Error("Branding color is missing or empty")
		}
	}
}

// testActionInputs verifies the inputs in the action.yml file
func testActionInputs(t *testing.T) {
	actionConfig := readActionYaml(t)

	inputs, ok := actionConfig["inputs"].(map[string]interface{})
	if !ok {
		t.Fatal("Inputs section is not properly formatted")
		return
	}

	// Expected inputs
	expectedInputs := map[string]struct {
		required    bool
		hasDefault  bool
		description string
	}{
		"github_token": {
			required:    false, // Has default
			hasDefault:  true,
			description: "token",
		},
		"output_format": {
			required:    false, // Has default
			hasDefault:  true,
			description: "format",
		},
		"policy_content": {
			required:    true,
			hasDefault:  false,
			description: "ignoring local policy files",
		},
	}

	// Check each expected input
	for inputName, expected := range expectedInputs {
		inputConfig, exists := inputs[inputName].(map[string]interface{})
		if !exists {
			t.Errorf("Input %q missing from action.yml", inputName)
			continue
		}

		// Check description
		description, descExists := inputConfig["description"].(string)
		if !descExists {
			t.Errorf("Input %q is missing a description", inputName)
		} else if !strings.Contains(strings.ToLower(description), expected.description) {
			t.Errorf("Input %q description should mention %q", inputName, expected.description)
		}

		// Check required flag
		required, requiredExists := inputConfig["required"].(bool)
		if expected.required && (!requiredExists || !required) {
			t.Errorf("Input %q should be marked as required", inputName)
		}

		// Check default value if expected
		if expected.hasDefault && inputConfig["default"] == nil {
			t.Errorf("Input %q should have a default value", inputName)
		}
	}
}

// testActionRuns verifies the runs section in the action.yml file
func testActionRuns(t *testing.T) {
	actionConfig := readActionYaml(t)

	runs, ok := actionConfig["runs"].(map[string]interface{})
	if !ok {
		t.Fatal("Runs section is not properly formatted")
		return
	}

	// Check for Docker configuration
	if using, ok := runs["using"].(string); !ok || using != "docker" {
		t.Errorf("Expected runs.using to be 'docker', got %v", using)
	}

	if image, ok := runs["image"].(string); !ok || image == "" {
		t.Error("Runs section is missing a valid 'image' field")
	}

	// Check command arguments
	args, ok := runs["args"].([]interface{})
	if !ok {
		t.Error("Runs section is missing 'args' field or it's not an array")
		return
	}

	// Check that first argument is the 'enforce' command
	if len(args) == 0 || args[0] != "enforce" {
		t.Error("The first argument should be the 'enforce' command")
	}

	// Check for required arguments and values
	requiredArgs := map[string]string{
		"--repo":           "${{ github.repository }}",
		"--output":         "${{ inputs.output_format }}",
		"--github-token":   "${{ github.token }}",
		"--policy-content": "${{ inputs.policy_content }}",
	}

	for flag, expectedValue := range requiredArgs {
		found := false
		for i, arg := range args {
			if arg == flag && i+1 < len(args) {
				if args[i+1] == expectedValue {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("Args should include '%s' flag with value '%s'", flag, expectedValue)
		}
	}
}

// testDockerfileConfig verifies the Dockerfile configuration
func testDockerfileConfig(t *testing.T) {
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		t.Fatalf("Dockerfile not found at %s", dockerfilePath)
	}

	data, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("Failed to read Dockerfile: %v", err)
	}

	content := string(data)

	// Check for essential components
	essentialPhrases := []string{
		"FROM alpine:",
		"WORKDIR /app",
		"COPY ./action-control-linux-${TARGETARCH}",
		"RUN apk --no-cache add ca-certificates",
		"ENTRYPOINT [",
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

	// Check for architecture-specific binary
	if !strings.Contains(content, "TARGETARCH") {
		t.Error("Dockerfile should use TARGETARCH for architecture-specific binaries")
	}
}

// testEntrypointScript verifies the entrypoint.sh script
func testEntrypointScript(t *testing.T) {
	if _, err := os.Stat(entrypointPath); os.IsNotExist(err) {
		t.Fatalf("entrypoint.sh file not found at %s", entrypointPath)
	}

	data, err := os.ReadFile(entrypointPath)
	if err != nil {
		t.Fatalf("Failed to read entrypoint.sh: %v", err)
	}

	content := string(data)

	// Check for shebang and file permissions
	if !strings.HasPrefix(content, "#!/bin/sh") {
		t.Error("entrypoint.sh should start with #!/bin/sh")
	}

	// Check for critical functionality
	essentialFunctionality := []string{
		"CMD=",                     // Command argument extraction
		"POLICY_CONTENT=",          // Policy content extraction
		"TEMP_POLICY_FILE=",        // Temporary file creation
		"--ignore-local-policy",    // Local policy override
		"exec /app/action-control", // Main binary execution
		"exit 1",                   // Error handling
	}

	for _, phrase := range essentialFunctionality {
		if !strings.Contains(content, phrase) {
			t.Errorf("entrypoint.sh should contain %q", phrase)
		}
	}
}

// testCommandLineHandling verifies that arguments are properly handled
func testCommandLineHandling(t *testing.T) {
	// Read entrypoint script
	scriptData, err := os.ReadFile(entrypointPath)
	if err != nil {
		t.Fatalf("Failed to read entrypoint.sh: %v", err)
	}

	scriptContent := string(scriptData)

	// Check that the script extracts command-line arguments properly
	argumentExtractions := []string{
		"CMD=\"$1\"",            // Extract command (enforce, report, etc)
		"REPO=\"$3\"",           // Extract repository
		"OUTPUT_FORMAT=\"$5\"",  // Extract output format
		"GITHUB_TOKEN=\"$7\"",   // Extract GitHub token
		"POLICY_CONTENT=\"$9\"", // Extract policy content
	}

	for _, extraction := range argumentExtractions {
		if !strings.Contains(scriptContent, extraction) {
			t.Errorf("entrypoint.sh should contain %q", extraction)
		}
	}

	// Check that command line arguments are passed to the executable
	execCommand := "exec /app/action-control"
	if !strings.Contains(scriptContent, execCommand) {
		t.Errorf("entrypoint.sh should execute the main binary with %q", execCommand)
	}

	// Check that the script validates required inputs
	if !strings.Contains(scriptContent, "if [ -n \"$GITHUB_TOKEN\" ]") {
		t.Error("entrypoint.sh should validate GitHub token")
	}

	if !strings.Contains(scriptContent, "if [ -n \"$POLICY_CONTENT\" ]") {
		t.Error("entrypoint.sh should validate policy content")
	}
}
