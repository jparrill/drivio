package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	// Release notes flags
	owner               string
	repo                string
	fromRef             string
	toRef               string
	releaseOutput       string
	releaseNotesWorkDir string
	githubToken         string
)

// GitHubCommit represents a commit from GitHub API
type GitHubCommit struct {
	Sha    string `json:"sha"`
	Commit struct {
		Author struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"author"`
		Message string `json:"message"`
	} `json:"commit"`
}

// releaseNotesCmd represents the release-notes command
var releaseNotesCmd = &cobra.Command{
	Use:   "release-notes",
	Short: "Generate release notes using GitHub API",
	Long: `Generate release notes between two references using GitHub API directly.

This command uses GitHub's API to fetch commits between two references and generates
release notes without cloning the repository.

Examples:
  drivio release-notes --owner openshift --repo hypershift --from v0.1.59 --to v0.1.63
  drivio release-notes --owner myorg --repo myrepo --from v1.0.0 --to v1.1.0 --output release-notes.md`,
	RunE: runReleaseNotes,
}

func init() {
	rootCmd.AddCommand(releaseNotesCmd)

	// Add flags
	releaseNotesCmd.Flags().StringVar(&owner, "owner", "", "GitHub repository owner/organization")
	releaseNotesCmd.Flags().StringVar(&repo, "repo", "", "GitHub repository name")
	releaseNotesCmd.Flags().StringVar(&fromRef, "from", "", "From reference (tag, commit, or branch)")
	releaseNotesCmd.Flags().StringVar(&toRef, "to", "", "To reference (tag, commit, or branch)")
	releaseNotesCmd.Flags().StringVar(&releaseOutput, "output", "", "Output file path (default: stdout)")
	releaseNotesCmd.Flags().StringVar(&releaseNotesWorkDir, "work-dir", ".drivio-work", "Working directory for generated files")
	releaseNotesCmd.Flags().StringVar(&githubToken, "github-token", "", "GitHub token for authentication (optional)")

	// Mark required flags
	releaseNotesCmd.MarkFlagRequired("owner")
	releaseNotesCmd.MarkFlagRequired("repo")
	releaseNotesCmd.MarkFlagRequired("from")
	releaseNotesCmd.MarkFlagRequired("to")
}

// loadEnvrc loads environment variables from .envrc file
func loadEnvrc() error {
	envrcPath := ".envrc"
	if _, err := os.Stat(envrcPath); os.IsNotExist(err) {
		return nil // .envrc doesn't exist, that's ok
	}

	file, err := os.Open(envrcPath)
	if err != nil {
		return fmt.Errorf("failed to open .envrc: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			value = strings.Trim(value, `"'`)
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

func runReleaseNotes(cmd *cobra.Command, args []string) error {
	// Load environment variables from .envrc
	if err := loadEnvrc(); err != nil {
		fmt.Printf("⚠️  Warning: failed to load .envrc: %v\n", err)
	}

	// Create work directory if it doesn't exist
	if err := os.MkdirAll(releaseNotesWorkDir, 0755); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

	// Load GitHub token from environment if not provided via flag
	if githubToken == "" {
		githubToken = os.Getenv("GITHUB_TOKEN")
		if githubToken == "" {
			fmt.Println("⚠️  No GitHub token provided. Using unauthenticated requests (may hit rate limits)")
		} else {
			fmt.Println("🔑 Using GitHub token from environment")
		}
	} else {
		fmt.Println("🔑 Using GitHub token from command line flag")
	}

	fmt.Printf("📊 Generating release notes for %s/%s from %s to %s\n", owner, repo, fromRef, toRef)

	// Get commits using GitHub API
	commits, err := getCommitsBetween(owner, repo, fromRef, toRef, githubToken)
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	fmt.Printf("🔍 Found %d commits between references\n", len(commits))

	// Only include merge commits whose body contains a line with the ticket format
	var filteredCommits []string
	ticketPattern := regexp.MustCompile(`^[A-Z]+-\d+:\s.+`)
	for _, commit := range commits {
		lines := strings.Split(commit.Commit.Message, "\n")
		if len(lines) == 0 {
			continue
		}
		subject := strings.TrimSpace(lines[0])

		// Only process merge commits
		if !strings.HasPrefix(subject, "Merge pull request") {
			continue
		}

		// Search for ticket line in the rest of the message
		for _, line := range lines[1:] {
			line = strings.TrimSpace(line)
			if ticketPattern.MatchString(line) {
				filteredCommits = append(filteredCommits, commit.Sha[:8]+" "+line)
				break
			}
		}
	}

	fmt.Printf("✅ Found %d merge commits with ticket format\n", len(filteredCommits))

	// Generate release notes: only the hash and ticket line, one per line
	var output strings.Builder
	for _, ticketLine := range filteredCommits {
		output.WriteString(ticketLine + "\n")
	}

	// Save to work directory
	defaultFileName := fmt.Sprintf("release-notes-%s-%s-%s.md", owner, repo, fromRef)
	workFilePath := filepath.Join(releaseNotesWorkDir, defaultFileName)

	if err := os.WriteFile(workFilePath, []byte(output.String()), 0644); err != nil {
		return fmt.Errorf("failed to write file to work directory: %w", err)
	}
	fmt.Printf("💾 Release notes saved to work directory: %s\n", workFilePath)

	if releaseOutput != "" {
		// If a specific output file is specified, also write there
		if releaseOutput != workFilePath {
			if err := os.WriteFile(releaseOutput, []byte(output.String()), 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			fmt.Printf("💾 Release notes also saved to: %s\n", releaseOutput)
		}
		// Show content on stdout when --output is specified
		fmt.Println(output.String())
	}

	return nil
}

// getCommitsBetween gets all commits between two references using GitHub API
func getCommitsBetween(owner, repo, fromRef, toRef, token string) ([]GitHubCommit, error) {
	// Use GitHub compare API to get commits between two references
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/compare/%s...%s", owner, repo, fromRef, toRef)

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "drivio-release-notes")
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var compareResult struct {
		Commits []GitHubCommit `json:"commits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&compareResult); err != nil {
		return nil, err
	}

	return compareResult.Commits, nil
}
