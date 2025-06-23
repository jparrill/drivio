package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version information
	Version    = "0.1.0"
	CommitHash = "unknown"
	BuildTime  = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "drivio",
	Short: "Drivio - CLI tool for production environment updates",
	Long: `Drivio is a command-line interface tool designed to help
manage and update production environments efficiently and safely.

Features:
  • Fetch configuration files from GitLab repositories
  • Validate connections and repository access
  • Support for multiple environments and branches
  • Secure token-based authentication

Examples:
  drivio fetch --token YOUR_TOKEN --repo owner/repo --file config.yaml
  drivio fetch --validate-only
  drivio --version`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", Version, CommitHash, BuildTime),
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Here you can define your flags and configuration settings
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.drivio.yaml)")
}
