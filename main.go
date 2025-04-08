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
	cobra.OnInitialize(initConfig)

	var rootCmd = &cobra.Command{
		Use:   "action-control",
		Short: "A CLI tool to enforce a Github actions policy that you create",
	}

	// Report command - keeps the existing functionality
	var reportCmd = &cobra.Command{
		Use:   "report",
		Short: "Report on GitHub Actions used in repositories across your organization.",
		Run: func(cmd *cobra.Command, args []string) {
			runReport()
		},
	}

	// Enforce command - new functionality
	var enforceCmd = &cobra.Command{
		Use:   "enforce",
		Short: "Enforce policy on GitHub Actions usage",
		Run: func(cmd *cobra.Command, args []string) {
			runEnforce()
		},
	}

	// Export command - generate a policy file from discovered actions
	var exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export a policy file based on discovered GitHub Actions",
		Run: func(cmd *cobra.Command, args []string) {
			runExport()
		},
	}

	// Add flags to commands
	rootCmd.PersistentFlags().String("config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().String("org", "", "GitHub organization name")
	rootCmd.PersistentFlags().String("repo", "", "Specific repository to check (format: owner/repo)")
	rootCmd.PersistentFlags().String("output", "", "Output format (markdown or json)")

	enforceCmd.Flags().String("policy", "policy.yaml", "Path to policy configuration file")

	exportCmd.Flags().String("file", "policy.yaml", "Output file path for generated policy")
	exportCmd.Flags().Bool("include-versions", false, "Include version tags in action references")
	exportCmd.Flags().Bool("include-custom", false, "Generate custom rules for each repository")
	exportCmd.Flags().String("policy-mode", "allow", "Policy mode: allow or deny") // Add this line

	// Bind flags to viper
	viper.BindPFlag("organization", rootCmd.PersistentFlags().Lookup("org"))
	viper.BindPFlag("repository", rootCmd.PersistentFlags().Lookup("repo"))
	viper.BindPFlag("output_format", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("policy_file", enforceCmd.Flags().Lookup("policy"))
	viper.BindPFlag("export_file", exportCmd.Flags().Lookup("file"))
	viper.BindPFlag("include_versions", exportCmd.Flags().Lookup("include-versions"))
	viper.BindPFlag("include_custom", exportCmd.Flags().Lookup("include-custom"))
	viper.BindPFlag("policy_mode", exportCmd.Flags().Lookup("policy-mode")) // Add this line

	// Add commands to root
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(enforceCmd)
	rootCmd.AddCommand(exportCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func runReport() {
	token := viper.GetString("github_token")
	if token == "" {
		log.Fatal("GitHub token not provided. Set it in config.yaml or as GITHUB_TOKEN environment variable.")
	}

	// Get org and repo parameters
	org := viper.GetString("organization")
	specificRepo := viper.GetString("repository")

	// Validate inputs
	if org == "" && specificRepo == "" {
		log.Fatal("Either organization (--org) or specific repository (--repo) must be provided.")
	}

	outputFormat := viper.GetString("output_format")
	if outputFormat == "" {
		outputFormat = "markdown"
	}

	client := github.NewClient(token)
	ctx := context.Background()

	githubActionsMap := make(map[string][]github.Action)
	var err error

	if specificRepo != "" {
		// Handle single repository case
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
		// Organization-wide scan
		fmt.Printf("Scanning repositories in %s organization...\n", org)
		githubActionsMap, err = client.ActionsForOrg(ctx, org)
		if err != nil {
			log.Fatalf("Error retrieving actions: %v", err)
		}
	}

	// Convert from github.Action to formatter.Action
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
	token := viper.GetString("github_token")
	if token == "" {
		log.Fatal("GitHub token not provided. Set it in config.yaml or as GITHUB_TOKEN environment variable.")
	}

	// Get org and repo parameters
	org := viper.GetString("organization")
	specificRepo := viper.GetString("repository")

	// Validate inputs
	if org == "" && specificRepo == "" {
		log.Fatal("Either organization (--org) or specific repository (--repo) must be provided.")
	}

	policyFile := viper.GetString("policy_file")
	if policyFile == "" {
		policyFile = "policy.yaml"
	}

	// Load local policy
	localPolicy, err := policy.LoadPolicyConfig(policyFile)
	if err != nil {
		log.Fatalf("Error loading policy file: %v", err)
	}

	client := github.NewClient(token)
	ctx := context.Background()

	githubActionsMap := make(map[string][]github.Action)

	if specificRepo != "" {
		// Handle single repository case
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
		// Organization-wide scan
		fmt.Printf("Scanning repositories in %s organization and enforcing policy...\n", org)
		githubActionsMap, err = client.ActionsForOrg(ctx, org)
		if err != nil {
			log.Fatalf("Error retrieving actions: %v", err)
		}
	}

	// Track policy violations
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

		// Try to load repo-specific policy
		var repoPolicy *policy.PolicyConfig
		repoPolicy = localPolicy // Start with local policy

		// Try to get repo policy file if it exists
		repoPolicyContent, err := client.GetRepositoryContent(ctx, owner, repoName, ".github/action-control-policy.yaml")
		if err == nil && len(repoPolicyContent) > 0 {
			// Merge with local policy
			repoPolicy, err = policy.MergeRepoPolicy(localPolicy, repoPolicyContent, repoFullName)
			if err != nil {
				log.Printf("Warning: Could not parse policy file in repository %s: %v", repoFullName, err)
				// Fall back to local policy
				repoPolicy = localPolicy
			}
		}

		// Extract action strings for policy check
		actionStrings := make([]string, len(actions))
		for i, action := range actions {
			actionStrings[i] = action.Uses
		}

		// Check compliance
		repoViolations, compliant := policy.CheckActionCompliance(repoPolicy, repoFullName, actionStrings)
		if !compliant {
			violations[repoFullName] = repoViolations
		}
	}

	// Report results
	report := formatter.FormatPolicyViolations(violations, localPolicy.PolicyMode)
	fmt.Println(report)

	// Exit with error if violations found
	if len(violations) > 0 {
		os.Exit(1)
	}
}

func runExport() {
	token := viper.GetString("github_token")
	if token == "" {
		log.Fatal("GitHub token not provided. Set it in config.yaml or as GITHUB_TOKEN environment variable.")
	}

	// Get org and repo parameters
	org := viper.GetString("organization")
	specificRepo := viper.GetString("repository")

	// Validate inputs
	if org == "" && specificRepo == "" {
		log.Fatal("Either organization (--org) or specific repository (--repo) must be provided.")
	}

	// Set up exporter with configuration
	exporter := export.NewExporter()
	exporter.OutputPath = viper.GetString("export_file")
	exporter.IncludeVersions = viper.GetBool("include_versions")
	exporter.IncludeCustom = viper.GetBool("include_custom")
	exporter.PolicyMode = viper.GetString("policy_mode") // Add this line

	// Validate policy mode
	if exporter.PolicyMode != "allow" && exporter.PolicyMode != "deny" {
		log.Fatalf("Invalid policy mode: %s, must be 'allow' or 'deny'", exporter.PolicyMode)
	}

	client := github.NewClient(token)
	ctx := context.Background()

	githubActionsMap := make(map[string][]github.Action)
	var err error

	if specificRepo != "" {
		// Handle single repository case
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
		// Organization-wide scan
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

	fmt.Printf("Successfully exported %s-mode policy file to %s\n", exporter.PolicyMode, exporter.OutputPath)

	// Adjust the output message to use the correct field based on policy mode
	if exporter.PolicyMode == "allow" {
		fmt.Printf("Found %d allowed actions across %d repositories\n",
			len(policyConfig.AllowedActions), len(githubActionsMap))
	} else {
		fmt.Printf("Found %d denied actions across %d repositories\n",
			len(policyConfig.DeniedActions), len(githubActionsMap))
	}
}

func initConfig() {
	if configFile := viper.GetString("config"); configFile != "" {
		// If a specific config file is provided, use it
		viper.SetConfigFile(configFile)
	} else {
		// Default config file name and type
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		// Add config search paths in order of precedence:
		// 1. Current directory (highest priority)
		viper.AddConfigPath(".")
		// 2. User's .config directory (conventional XDG config location)
		viper.AddConfigPath("$HOME/.config/action-control")
		// 3. User's home directory (fallback)
		viper.AddConfigPath("$HOME")
	}

	// Enable environment variable overrides
	viper.AutomaticEnv()
	viper.SetEnvPrefix("ACTION_CONTROL")

	// Try to read the config file, but continue if not found
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, using defaults
			log.Println("No configuration file found, using defaults and environment variables")
		} else {
			// Config file was found but another error occurred
			log.Printf("Warning: Error reading config file: %v", err)
		}
	} else {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}
}
