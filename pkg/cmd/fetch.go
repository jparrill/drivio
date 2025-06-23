package cmd

import (
	"context"
	"fmt"
	"os"

	"drivio/pkg/config"
	"drivio/pkg/gitlab"

	"github.com/spf13/cobra"
)

var (
	// Configuration flags
	gitlabURL      string
	gitlabToken    string
	repositoryPath string
	branch         string
	filePath       string
	outputFile     string
	validateOnly   bool
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch a YAML file from a GitLab repository",
	Long: `Fetch a YAML file from a GitLab repository with configurable parameters.

Examples:
  drivio fetch --repo gitlab-org/gitlab-foss --file db/database_connections/ci.yaml
  drivio fetch --repo jparrill/my-config --file config/production.yaml --token YOUR_TOKEN
  drivio fetch --branch develop --output config.yaml
  drivio fetch --validate-only`,
	RunE: runFetch,
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	// Add flags
	fetchCmd.Flags().StringVar(&gitlabURL, "url", "", "GitLab URL (default: https://gitlab.com)")
	fetchCmd.Flags().StringVar(&gitlabToken, "token", "", "GitLab access token (optional for public repositories)")
	fetchCmd.Flags().StringVar(&repositoryPath, "repo", "", "Repository path (e.g., owner/repo)")
	fetchCmd.Flags().StringVar(&branch, "branch", "", "Branch name")
	fetchCmd.Flags().StringVar(&filePath, "file", "", "Path to the file in the repository")
	fetchCmd.Flags().StringVar(&outputFile, "output", "", "Output file path (default: stdout)")
	fetchCmd.Flags().BoolVar(&validateOnly, "validate-only", false, "Only validate connection and repository access")

	// Remove the required flag for token since it's optional for public repos
}

func runFetch(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg := config.LoadConfig()

	// Override with flags if provided
	if gitlabURL != "" {
		cfg.GitLabURL = gitlabURL
	}
	if gitlabToken != "" {
		cfg.GitLabToken = gitlabToken
	}
	if repositoryPath != "" {
		cfg.RepositoryPath = repositoryPath
	}
	if branch != "" {
		cfg.Branch = branch
	}
	if filePath != "" {
		cfg.FilePath = filePath
	}

	// Validate configuration
	if err := cfg.ValidateConfig(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Check if token is required
	if cfg.RequiresToken() {
		return fmt.Errorf("GitLab token is required for this repository. Set GITLAB_TOKEN environment variable or use --token flag")
	}

	// Create GitLab client
	client, err := gitlab.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create GitLab client: %w", err)
	}

	// Create context
	ctx := context.Background()

	// Validate connection
	if cfg.IsPublicRepository() && cfg.GitLabToken == "" {
		fmt.Printf("🔍 Accessing public repository: %s\n", cfg.RepositoryPath)
	} else {
		fmt.Printf("🔍 Validating GitLab connection...\n")
		if err := client.ValidateConnection(ctx); err != nil {
			return fmt.Errorf("connection validation failed: %w", err)
		}
		fmt.Printf("✅ GitLab connection validated successfully\n")
	}

	// Get repository info
	fmt.Printf("📁 Checking repository: %s\n", cfg.RepositoryPath)
	project, err := client.GetRepositoryInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get repository info: %w", err)
	}
	fmt.Printf("✅ Repository found: %s\n", project.Name)

	if validateOnly {
		fmt.Printf("✅ Validation completed successfully\n")
		return nil
	}

	// Fetch the file
	fmt.Printf("📄 Fetching file: %s from branch: %s\n", cfg.FilePath, cfg.Branch)
	content, err := client.GetFile(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch file: %w", err)
	}
	fmt.Printf("✅ File fetched successfully (%d bytes)\n", len(content))

	// Output the content
	if outputFile != "" {
		// Write to file
		if err := os.WriteFile(outputFile, content, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("💾 File saved to: %s\n", outputFile)
	} else {
		// Write to stdout
		fmt.Println(string(content))
	}

	return nil
}
