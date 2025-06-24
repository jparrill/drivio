package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"drivio/pkg/config"
	"drivio/pkg/gitlab"
	"drivio/pkg/ui"

	"github.com/spf13/cobra"
	gitlabAPI "gitlab.com/gitlab-org/api/client-go"
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
	fetchWorkDir   string
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
	fetchCmd.Flags().StringVar(&fetchWorkDir, "work-dir", ".drivio-work", "Working directory for downloaded files")

	// Remove the required flag for token since it's optional for public repos
}

func runFetch(cmd *cobra.Command, args []string) error {
	// Create work directory if it doesn't exist
	if err := os.MkdirAll(fetchWorkDir, 0755); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

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

	ctx := context.Background()

	// Step 1: Validate connection
	if err := ui.RunSpinner("Validating GitLab connection...", func() error {
		if cfg.IsPublicRepository() && cfg.GitLabToken == "" {
			return nil // No validation needed for public repo
		}
		return client.ValidateConnection(ctx)
	}); err != nil {
		return fmt.Errorf("connection validation failed: %w", err)
	}

	// Step 2: Get repository info
	var project *gitlabAPI.Project
	if err := ui.RunSpinner("Getting repository info...", func() error {
		var err error
		project, err = client.GetRepositoryInfo(ctx)
		return err
	}); err != nil {
		return fmt.Errorf("failed to get repository info: %w", err)
	}
	fmt.Printf("✅ Repository found: %s\n", project.Name)

	if validateOnly {
		fmt.Printf("✅ Validation completed successfully\n")
		return nil
	}

	// Step 3: Fetch the file
	var content []byte
	if err := ui.RunSpinner("Fetching file...", func() error {
		var err error
		content, err = client.GetFile(ctx)
		return err
	}); err != nil {
		return fmt.Errorf("failed to fetch file: %w", err)
	}
	fmt.Printf("✅ File fetched successfully (%d bytes)\n", len(content))

	// Step 4: Save to work directory
	defaultFileName := "fetched_file.yaml"
	workFilePath := filepath.Join(fetchWorkDir, defaultFileName)
	if err := ui.RunSpinner("Saving file...", func() error {
		return os.WriteFile(workFilePath, content, 0644)
	}); err != nil {
		return fmt.Errorf("failed to write file to work directory: %w", err)
	}
	fmt.Printf("💾 File saved successfully: %s\n", workFilePath)

	if outputFile != "" {
		// If a specific output file is specified, also write there and show content
		if outputFile != workFilePath {
			if err := os.WriteFile(outputFile, content, 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			fmt.Printf("💾 File also saved to: %s\n", outputFile)
		}
		// Show content on stdout when --output is specified
		fmt.Println(string(content))
	}
	// If no --output is specified, don't show content on stdout

	return nil
}
