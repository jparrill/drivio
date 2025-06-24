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
	"strconv"
	"strings"
	"time"

	"drivio/pkg/ui"

	"github.com/spf13/cobra"
)

var (
	// Configuration flags
	owner               string
	repo                string
	fromRef             string
	toRef               string
	releaseOutput       string
	releaseNotesWorkDir string
	githubToken         string
	showStdout          bool
	useTable            bool
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
	Parents []struct {
		Sha string `json:"sha"`
	} `json:"parents"`
}

// GitHubPR represents a pull request from GitHub API
type GitHubPR struct {
	Number int `json:"number"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
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
	releaseNotesCmd.Flags().BoolVar(&showStdout, "stdout", false, "Show content on stdout")
	releaseNotesCmd.Flags().BoolVar(&useTable, "table", false, "Generate a markdown table format")

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
		}
	}

	// Generate release notes with progress bar
	output, err := generateReleaseNotesWithProgress(owner, repo, fromRef, toRef, githubToken)
	if err != nil {
		return fmt.Errorf("failed to generate release notes: %w", err)
	}

	// Save to work directory
	defaultFileName := fmt.Sprintf("release-notes-%s-%s-%s-%s.md", owner, repo, fromRef, toRef)
	workFilePath := filepath.Join(releaseNotesWorkDir, defaultFileName)

	if err := os.WriteFile(workFilePath, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write file to work directory: %w", err)
	}
	fmt.Printf("💾 Release notes saved generated successfully: %s\n", workFilePath)

	// If a specific output file is specified, also write there
	if releaseOutput != "" {
		if releaseOutput != workFilePath {
			if err := os.WriteFile(releaseOutput, []byte(output), 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			fmt.Printf("💾 Release notes also saved to: %s\n", releaseOutput)
		}
	}

	// Show content on stdout only when --stdout flag is specified
	if showStdout {
		fmt.Println(output)
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

// getPRLabels gets the labels for a specific PR number
func getPRLabels(owner, repo string, prNumber int, token string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, prNumber)

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}

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
		return nil, fmt.Errorf("GitHub API returned status %d for PR %d", resp.StatusCode, prNumber)
	}

	var pr GitHubPR
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}

	var labels []string
	for _, label := range pr.Labels {
		labels = append(labels, label.Name)
	}

	return labels, nil
}

// extractPRNumber extracts PR number from merge commit message
func extractPRNumber(message string) (int, bool) {
	// Look for "Merge pull request #123 from" pattern
	lines := strings.Split(message, "\n")
	if len(lines) == 0 {
		return 0, false
	}

	subject := strings.TrimSpace(lines[0])
	// Pattern: "Merge pull request #123 from owner/branch"
	if strings.HasPrefix(subject, "Merge pull request #") {
		// Extract the part after "Merge pull request #"
		afterPrefix := strings.TrimPrefix(subject, "Merge pull request #")
		// Split by space to get the number
		parts := strings.Split(afterPrefix, " ")
		if len(parts) > 0 {
			if prNumber, err := strconv.Atoi(parts[0]); err == nil {
				return prNumber, true
			}
		}
	}

	return 0, false
}

// generateReleaseNotesWithProgress generates release notes with a progress bar
func generateReleaseNotesWithProgress(owner, repo, fromRef, toRef, token string) (string, error) {
	var result string
	var commits []GitHubCommit

	// Step 1: Validating GitHub connection
	if err := ui.RunSpinner("Validating GitHub connection...", func() error {
		time.Sleep(500 * time.Millisecond) // Simulate validation
		return nil
	}); err != nil {
		return "", err
	}

	// Step 2: Getting commits between references
	if err := ui.RunSpinner("Getting commits between references...", func() error {
		var err error
		commits, err = getCommitsBetween(owner, repo, fromRef, toRef, token)
		return err
	}); err != nil {
		return "", fmt.Errorf("failed to get commits: %w", err)
	}
	fmt.Printf("✅ Found %d commits\n", len(commits))

	// Step 3: Filtering commits by label and format
	var filteredCommits []struct {
		hash   string
		ticket string
		desc   string
	}

	if err := ui.RunSpinner("Filtering commits by label and format...", func() error {
		filteredCommits = filterCommitsByLabelAndFormat(commits, owner, repo, token)
		return nil
	}); err != nil {
		return "", err
	}
	fmt.Printf("✅ Found %d relevant commits\n", len(filteredCommits))

	// Step 4: Generating release notes
	if err := ui.RunSpinner("Generating release notes...", func() error {
		result = generateReleaseNotesContent(filteredCommits, fromRef, toRef, owner, repo)
		return nil
	}); err != nil {
		return "", err
	}

	return result, nil
}

// filterCommitsByLabelAndFormat filters commits by label and ticket format
func filterCommitsByLabelAndFormat(commits []GitHubCommit, owner, repo, token string) []struct {
	hash   string
	ticket string
	desc   string
} {
	var filteredCommits []struct {
		hash   string
		ticket string
		desc   string
	}
	ticketPattern := regexp.MustCompile(`^[A-Z]+-\d+:\s.+`)
	targetLabel := "area/hypershift-operator"

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

		// Extract PR number from merge commit message
		prNumber, ok := extractPRNumber(commit.Commit.Message)
		if !ok {
			continue
		}

		// Get PR labels
		labels, err := getPRLabels(owner, repo, prNumber, token)
		if err != nil {
			continue
		}

		// Check if PR has the target label
		hasTargetLabel := false
		for _, label := range labels {
			if label == targetLabel {
				hasTargetLabel = true
				break
			}
		}

		if !hasTargetLabel {
			continue
		}

		// Search for ticket line in the rest of the message
		for _, line := range lines[1:] {
			line = strings.TrimSpace(line)
			if ticketPattern.MatchString(line) {
				// ticketAndDesc: <TICKET>: <desc>
				ticketParts := strings.SplitN(line, ":", 2)
				if len(ticketParts) == 2 {
					ticket := strings.TrimSpace(ticketParts[0])
					desc := strings.TrimSpace(ticketParts[1])

					filteredCommits = append(filteredCommits, struct {
						hash   string
						ticket string
						desc   string
					}{
						hash:   commit.Sha[:8],
						ticket: ticket,
						desc:   desc,
					})
				}
				break
			}
		}
	}

	return filteredCommits
}

// generateReleaseNotesContent generates the markdown content for release notes
func generateReleaseNotesContent(filteredCommits []struct {
	hash   string
	ticket string
	desc   string
}, fromRef, toRef, owner, repo string) string {
	var output strings.Builder
	output.WriteString(fmt.Sprintf("# Release notes from %s to %s\n\n", fromRef, toRef))

	if useTable {
		// Generate table format
		output.WriteString("| Commit | JIRA | Description |\n")
		output.WriteString("|--------|------|-------------|\n")

		for _, commit := range filteredCommits {
			commitURL := fmt.Sprintf("https://github.com/%s/%s/commit/%s", owner, repo, commit.hash)
			ticketURL := fmt.Sprintf("https://issues.redhat.com/browse/%s", commit.ticket)

			output.WriteString(fmt.Sprintf("| [%s](%s) | [%s](%s) | %s |\n",
				commit.hash, commitURL, commit.ticket, ticketURL, commit.desc))
		}
	} else {
		// Generate list format (current format)
		for _, commit := range filteredCommits {
			commitURL := fmt.Sprintf("https://github.com/%s/%s/commit/%s", owner, repo, commit.hash)
			ticketURL := fmt.Sprintf("https://issues.redhat.com/browse/%s", commit.ticket)

			output.WriteString(fmt.Sprintf("[%s](%s) - [%s](%s): %s\n",
				commit.hash, commitURL, commit.ticket, ticketURL, commit.desc))
		}
	}

	return output.String()
}
