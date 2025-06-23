package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "drivio",
	Short: "Drivio - CLI tool for production environment updates",
	Long: `Drivio is a command-line interface tool designed to help
manage and update production environments efficiently and safely.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Here you can define your flags and configuration settings
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.drivio.yaml)")
}
