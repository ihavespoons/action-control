package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ihavespoons/action-control/internal/export"
	"github.com/ihavespoons/action-control/internal/formatter"
	"github.com/ihavespoons/action-control/internal/github"
	"github.com/ihavespoons/action-control/internal/policy"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	// Initialize configuration before command execution
	cobra.OnInitialize(initConfig)

	// Define root command
	var rootCmd = &cobra.Command{
		Use:   "action-control",
		Short: "A CLI tool to enforce a Github actions policy that you create",
	}

	// Define subcommands
	var reportCmd = &cobra.Command{
		Use:   "report",
		Short: "Report on GitHub Actions used in repositories across your organization",
		Run: func(cmd *cobra.Command, args []string) {
			runReport()
		},
	}

	var enforceCmd = &cobra.Command{
		Use:   "enforce",
		Short: "Enforce policy on GitHub Actions usage",
		Run: func(cmd *cobra.Command, args []string) {
			runEnforce()
		},
	}

	var exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export a policy file based on discovered GitHub Actions",
		Run: func(cmd *cobra.Command, args []string) {
			runExport()
		},
	}

	// Configure global flags available to all commands
	rootCmd.PersistentFlags().String("config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().String("org", "", "GitHub organization name")
	rootCmd.PersistentFlags().String("repo", "", "Specific repository to check (format: owner/repo)")
	rootCmd.PersistentFlags().String("output", "", "Output format (markdown or json)")

	// Configure command-specific flags
	enforceCmd.Flags().String("policy", "policy.yaml", "Path to policy configuration file")
	enforceCmd.Flags().Bool("ignore-local-policy", false, "Ignore local policy files and only use provided policy")
	enforceCmd.Flags().MarkHidden("ignore-local-policy") // Hidden flag for internal use

	exportCmd.Flags().String("file", "policy.yaml", "Output file path for generated policy")
	exportCmd.Flags().Bool("include-versions", false, "Include version tags in action references")
	exportCmd.Flags().Bool("include-custom", false, "Generate custom rules for each repository")
	exportCmd.Flags().String("policy-mode", "allow", "Policy mode: allow or deny")

	// Bind flags to viper to enable config file and environment variable usage
	viper.BindPFlag("organization", rootCmd.PersistentFlags().Lookup("org"))
	viper.BindPFlag("repository", rootCmd.PersistentFlags().Lookup("repo"))
	viper.BindPFlag("output_format", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("policy_file", enforceCmd.Flags().Lookup("policy"))
	viper.BindPFlag("ignore_local_policy", enforceCmd.Flags().Lookup("ignore-local-policy"))
	viper.BindPFlag("export_file", exportCmd.Flags().Lookup("file"))
	viper.BindPFlag("include_versions", exportCmd.Flags().Lookup("include-versions"))
	viper.BindPFlag("include_custom", exportCmd.Flags().Lookup("include-custom"))
	viper.BindPFlag("policy_mode", exportCmd.Flags().Lookup("policy-mode"))

	// Add subcommands to root command
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(enforceCmd)
	rootCmd.AddCommand(exportCmd)

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func runReport() {
	// Validate GitHub token
	token := viper.GetString("github_token")
	if token == "" {
		log.Fatal("GitHub token not provided. Set it in config.yaml or as GITHUB_TOKEN environment variable.")
	}

	// Get target organization or repository
	org := viper.GetString("organization")
	specificRepo := viper.GetString("repository")

	// At least one target must be specified
	if org == "" && specificRepo == "" {
		log.Fatal("Either organization (--org) or specific repository (--repo) must be provided.")
	}

	// Set default output format if not specified
	outputFormat := viper.GetString("output_format")
	if outputFormat == "" {
		outputFormat = "markdown"
	}

	// Initialize GitHub API client
	client := github.NewClient(token)
	ctx := context.Background()

	// Map to store discovered actions by repository
	githubActionsMap := make(map[string][]github.Action)
	var err error

	// Fetch actions from GitHub
	if specificRepo != "" {
		// Scan a single repository
		parts := strings.Split(specificRepo, "/")
		if len(parts) != 2 {
			log.Fatalf("Invalid repository format. Use 'owner/repo' format.")
		}
		owner, repo := parts[0], parts[1]

		fmt.Printf("Scanning repository %s...\n", specificRepo)
		actions, err := client.GetActions(ctx, owner, repo)
		if err != nil {
			log.Fatalf("Error retrieving actions from repository %s: %v", specificRepo, err)
		}
		if len(actions) > 0 {
			githubActionsMap[specificRepo] = actions
		}
	} else {
		// Scan an entire organization
		fmt.Printf("Scanning repositories in %s organization...\n", org)
		githubActionsMap, err = client.ActionsForOrg(ctx, org)
		if err != nil {
			log.Fatalf("Error retrieving actions: %v", err)
		}
	}

	// Convert GitHub actions to formatter-compatible structure
	actionsMap := make(map[string][]formatter.Action)
	for repo, actions := range githubActionsMap {
		formatterActions := make([]formatter.Action, len(actions))
		for i, action := range actions {
			formatterActions[i] = formatter.Action{
				Name: action.Name,
				Uses: action.Uses,
			}
		}
		actionsMap[repo] = formatterActions
	}

	// Format and output the results
	var result string
	switch outputFormat {
	case "json":
		jsonData, err := formatter.FormatJSON(actionsMap)
		if err != nil {
			log.Fatalf("Error formatting JSON: %v", err)
		}
		result = jsonData
	case "markdown":
		result = formatter.FormatMarkdown(actionsMap)
	default:
		log.Fatalf("Unsupported output format: %s", outputFormat)
	}

	fmt.Println(result)
}

func runEnforce() {
	// Validate GitHub token
	token := viper.GetString("github_token")
	if token == "" {
		log.Fatal("GitHub token not provided. Set it in config.yaml or as GITHUB_TOKEN environment variable.")
	}

	// Get target organization or repository
	org := viper.GetString("organization")
	specificRepo := viper.GetString("repository")

	// At least one target must be specified
	if org == "" && specificRepo == "" {
		log.Fatal("Either organization (--org) or specific repository (--repo) must be provided.")
	}

	// Determine policy source: environment variable or file
	policyContent := os.Getenv("ACTION_CONTROL_POLICY_CONTENT")
	ignoreLocalPolicy := viper.GetBool("ignore_local_policy")

	var localPolicy *policy.PolicyConfig
	var err error

	// Handle policy from environment variable with highest priority when flag is set
	if policyContent != "" && ignoreLocalPolicy {
		log.Println("Using policy from environment variable")

		// Create a temporary file for the policy content
		tmpFile, err := os.CreateTemp("", "policy-*.yaml")
		if err != nil {
			log.Fatalf("Error creating temporary policy file: %v", err)
		}
		defer os.Remove(tmpFile.Name()) // Clean up after we're done

		// Write content to temporary file
		if _, err := tmpFile.WriteString(policyContent); err != nil {
			tmpFile.Close()
			log.Fatalf("Error writing to temporary policy file: %v", err)
		}
		tmpFile.Close()

		// Load policy configuration from temporary file
		localPolicy, err = policy.LoadPolicyConfig(tmpFile.Name())
		if err != nil {
			log.Fatalf("Error loading policy from environment variable: %v", err)
		}
	} else {
		// Use policy from file
		policyFile := viper.GetString("policy_file")
		if policyFile == "" {
			policyFile = "policy.yaml"
		}

		// Load policy configuration from file
		localPolicy, err = policy.LoadPolicyConfig(policyFile)
		if err != nil {
			log.Fatalf("Error loading policy file: %v", err)
		}
	}

	// Initialize GitHub API client
	client := github.NewClient(token)
	ctx := context.Background()

	// Map to store discovered actions by repository
	githubActionsMap := make(map[string][]github.Action)

	// Fetch actions from GitHub
	if specificRepo != "" {
		// Scan a single repository
		parts := strings.Split(specificRepo, "/")
		if len(parts) != 2 {
			log.Fatalf("Invalid repository format. Use 'owner/repo' format.")
		}
		owner, repo := parts[0], parts[1]

		fmt.Printf("Scanning repository %s and enforcing policy...\n", specificRepo)
		actions, err := client.GetActions(ctx, owner, repo)
		if err != nil {
			log.Fatalf("Error retrieving actions from repository %s: %v", specificRepo, err)
		}
		if len(actions) > 0 {
			githubActionsMap[specificRepo] = actions
		}
	} else {
		// Scan an entire organization
		fmt.Printf("Scanning repositories in %s organization and enforcing policy...\n", org)
		githubActionsMap, err = client.ActionsForOrg(ctx, org)
		if err != nil {
			log.Fatalf("Error retrieving actions: %v", err)
		}
	}

	// Track policy violations found
	violations := make(map[string][]string)

	// Check each repository against policy
	for repoFullName, actions := range githubActionsMap {
		// Extract owner and repo name
		parts := strings.Split(repoFullName, "/")
		if len(parts) != 2 {
			continue
		}
		owner := parts[0]
		repoName := parts[1]

		// Use local policy as base
		var repoPolicy *policy.PolicyConfig
		repoPolicy = localPolicy

		// Check for repository-specific policy if not ignoring local policies
		if !ignoreLocalPolicy {
			repoPolicyContent, err := client.GetRepositoryContent(ctx, owner, repoName, ".github/action-control-policy.yaml")
			if err == nil && len(repoPolicyContent) > 0 {
				// Merge repository policy with local policy
				repoPolicy, err = policy.MergeRepoPolicy(localPolicy, repoPolicyContent, repoFullName)
				if err != nil {
					log.Printf("Warning: Could not parse policy file in repository %s: %v", repoFullName, err)
					// Fall back to local policy on error
					repoPolicy = localPolicy
				}
			}
		}

		// Extract action strings for policy check
		actionStrings := make([]string, len(actions))
		for i, action := range actions {
			actionStrings[i] = action.Uses
		}

		// Check actions against policy
		repoViolations, compliant := policy.CheckActionCompliance(repoPolicy, repoFullName, actionStrings)
		if !compliant {
			violations[repoFullName] = repoViolations
		}
	}

	// Generate and print report
	report := formatter.FormatPolicyViolations(violations, localPolicy.PolicyMode)
	fmt.Println(report)

	// Exit with error code if violations found
	if len(violations) > 0 {
		os.Exit(1)
	}
}

func runExport() {
	// Validate GitHub token
	token := viper.GetString("github_token")
	if token == "" {
		log.Fatal("GitHub token not provided. Set it in config.yaml or as GITHUB_TOKEN environment variable.")
	}

	// Get target organization or repository
	org := viper.GetString("organization")
	specificRepo := viper.GetString("repository")

	// At least one target must be specified
	if org == "" && specificRepo == "" {
		log.Fatal("Either organization (--org) or specific repository (--repo) must be provided.")
	}

	// Configure exporter with user preferences
	exporter := export.NewExporter()
	exporter.OutputPath = viper.GetString("export_file")
	exporter.IncludeVersions = viper.GetBool("include_versions")
	exporter.IncludeCustom = viper.GetBool("include_custom")
	exporter.PolicyMode = viper.GetString("policy_mode")

	// Validate policy mode
	if exporter.PolicyMode != "allow" && exporter.PolicyMode != "deny" {
		log.Fatalf("Invalid policy mode: %s, must be 'allow' or 'deny'", exporter.PolicyMode)
	}

	// Initialize GitHub API client
	client := github.NewClient(token)
	ctx := context.Background()

	// Map to store discovered actions by repository
	githubActionsMap := make(map[string][]github.Action)
	var err error

	// Fetch actions from GitHub
	if specificRepo != "" {
		// Export from a single repository
		parts := strings.Split(specificRepo, "/")
		if len(parts) != 2 {
			log.Fatalf("Invalid repository format. Use 'owner/repo' format.")
		}
		owner, repo := parts[0], parts[1]

		fmt.Printf("Scanning repository %s for actions...\n", specificRepo)
		actions, err := client.GetActions(ctx, owner, repo)
		if err != nil {
			log.Fatalf("Error retrieving actions from repository %s: %v", specificRepo, err)
		}
		if len(actions) > 0 {
			githubActionsMap[specificRepo] = actions
		}
	} else {
		// Export from an entire organization
		fmt.Printf("Scanning repositories in %s organization for actions...\n", org)
		githubActionsMap, err = client.ActionsForOrg(ctx, org)
		if err != nil {
			log.Fatalf("Error retrieving actions: %v", err)
		}
	}

	// Generate policy from discovered actions
	policyConfig, err := exporter.GeneratePolicyFromActions(githubActionsMap)
	if err != nil {
		log.Fatalf("Error generating policy: %v", err)
	}

	// Export the policy to a file
	if err := exporter.ExportPolicyFile(policyConfig); err != nil {
		log.Fatalf("Error writing policy file: %v", err)
	}

	// Print success message
	fmt.Printf("Successfully exported %s-mode policy file to %s\n", exporter.PolicyMode, exporter.OutputPath)

	// Show summary based on policy mode
	if exporter.PolicyMode == "allow" {
		fmt.Printf("Found %d allowed actions across %d repositories\n",
			len(policyConfig.AllowedActions), len(githubActionsMap))
	} else {
		fmt.Printf("Found %d denied actions across %d repositories\n",
			len(policyConfig.DeniedActions), len(githubActionsMap))
	}
}

// initConfig reads configuration from file and environment variables
func initConfig() {
	if configFile := viper.GetString("config"); configFile != "" {
		// Use specific config file if provided
		viper.SetConfigFile(configFile)
	} else {
		// Set default config name and type
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		// Add search paths in order of precedence
		viper.AddConfigPath(".")                            // Current directory (highest priority)
		viper.AddConfigPath("$HOME/.config/action-control") // XDG config location
		viper.AddConfigPath("$HOME")                        // User's home directory (fallback)
	}

	// Enable environment variable overrides
	viper.AutomaticEnv()
	viper.SetEnvPrefix("ACTION_CONTROL")

	// Try to read config file, but continue if not found
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found is okay, just use defaults
			log.Println("No configuration file found, using defaults and environment variables")
		} else {
			// Other config errors should be reported
			log.Printf("Warning: Error reading config file: %v", err)
		}
	} else {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}
}
