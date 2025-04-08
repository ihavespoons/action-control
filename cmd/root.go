package cmd

import (
	"github.com/ihavespoons/action-control/internal/report"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	orgName      string
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:   "github-actions-reporter",
	Short: "A CLI tool to report GitHub Actions used across repositories in an organization",
	Run: func(cmd *cobra.Command, args []string) {
		// Call the report generation function
		orgData := map[string][]string{
			"orgName": {orgName},
		}
		report.GenerateReport(orgData, outputFormat)
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&orgName, "org", "", "GitHub organization name (required)")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "markdown", "Output format (markdown or json)")

	rootCmd.MarkPersistentFlagRequired("org")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		// Handle error
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Handle error
	}
}
