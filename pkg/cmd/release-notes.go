package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"drivio/pkg/git"

	"github.com/spf13/cobra"
)

var (
	// Release notes flags
	repoPath            string
	fromRef             string
	toRef               string
	releaseOutput       string
	outputFormat        string
	includeTypes        []string
	excludeTypes        []string
	remoteURL           string
	releaseNotesWorkDir string
)

// releaseNotesCmd represents the release-notes command
var releaseNotesCmd = &cobra.Command{
	Use:   "release-notes",
	Short: "Generate release notes between two references",
	Long: `Generate release notes between two references (tags, commits, or branches).

The command analyzes commits between two references and generates formatted release notes
based on conventional commit messages.

Examples:
  drivio release-notes --from v1.0.0 --to v1.1.0
  drivio release-notes --from abc123 --to def456 --format json
  drivio release-notes --from main --to develop --output release-notes.md
  drivio release-notes --from v1.0.0 --to v1.1.0 --include feat,fix
  drivio release-notes --remote-url https://gitlab.com/gitlab-org/gitlab-foss.git --from v16.0.0 --to v16.1.0`,
	RunE: runReleaseNotes,
}

func init() {
	rootCmd.AddCommand(releaseNotesCmd)

	// Add flags
	releaseNotesCmd.Flags().StringVar(&repoPath, "repo", ".", "Repository path (default: current directory)")
	releaseNotesCmd.Flags().StringVar(&fromRef, "from", "", "From reference (tag, commit, or branch)")
	releaseNotesCmd.Flags().StringVar(&toRef, "to", "", "To reference (tag, commit, or branch)")
	releaseNotesCmd.Flags().StringVar(&releaseOutput, "output", "", "Output file path (default: stdout)")
	releaseNotesCmd.Flags().StringVar(&outputFormat, "format", "markdown", "Output format (markdown, json, text)")
	releaseNotesCmd.Flags().StringSliceVar(&includeTypes, "include", nil, "Include only specific commit types (e.g., feat,fix)")
	releaseNotesCmd.Flags().StringSliceVar(&excludeTypes, "exclude", nil, "Exclude specific commit types (e.g., chore,docs)")
	releaseNotesCmd.Flags().StringVar(&remoteURL, "remote-url", "", "Remote repository URL to clone and analyze (overrides --repo)")
	releaseNotesCmd.Flags().StringVar(&releaseNotesWorkDir, "work-dir", ".drivio-work", "Working directory for cloned repositories and generated files")

	// Mark required flags
	releaseNotesCmd.MarkFlagRequired("from")
	releaseNotesCmd.MarkFlagRequired("to")
}

func runReleaseNotes(cmd *cobra.Command, args []string) error {
	// Create work directory if it doesn't exist
	if err := os.MkdirAll(releaseNotesWorkDir, 0755); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}

	var repoDir string
	var cleanup func()

	if remoteURL != "" {
		// Create a unique subdirectory for this repository
		timestamp := time.Now().Format("20060102-150405")
		repoName := extractRepoName(remoteURL)
		cloneDir := filepath.Join(releaseNotesWorkDir, fmt.Sprintf("%s-%s", repoName, timestamp))

		if err := os.MkdirAll(cloneDir, 0755); err != nil {
			return fmt.Errorf("failed to create clone directory: %w", err)
		}

		fmt.Printf("🌐 Cloning remote repository: %s\n", remoteURL)

		// Use the new progress bar for cloning
		err := git.CloneWithProgress(remoteURL, cloneDir)
		if err != nil {
			os.RemoveAll(cloneDir)
			return fmt.Errorf("failed to clone remote repository: %w", err)
		}

		fmt.Printf("🔍 Analyzing repository: %s\n", cloneDir)

		repoDir = cloneDir
		cleanup = func() {
			fmt.Printf("🧹 Cleaning up cloned repository: %s\n", cloneDir)
			os.RemoveAll(cloneDir)
		}
		defer cleanup()
	} else {
		repoDir = repoPath
	}

	if repoDir == "." {
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		repoDir = currentDir
	}

	if _, err := os.Stat(filepath.Join(repoDir, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository: %s", repoDir)
	}

	format := git.OutputFormat(outputFormat)
	switch format {
	case git.FormatMarkdown, git.FormatJSON, git.FormatText:
		// Valid format
	default:
		return fmt.Errorf("unsupported output format: %s. Supported formats: markdown, json, text", outputFormat)
	}

	fmt.Printf("📊 Generating release notes from %s to %s\n", fromRef, toRef)

	analyzer, err := git.NewAnalyzer(repoDir)
	if err != nil {
		return fmt.Errorf("failed to create analyzer: %w", err)
	}

	notes, err := analyzer.GenerateReleaseNotes(fromRef, toRef)
	if err != nil {
		return fmt.Errorf("failed to generate release notes: %w", err)
	}

	if len(includeTypes) > 0 || len(excludeTypes) > 0 {
		notes.Commits = filterCommits(notes.Commits, includeTypes, excludeTypes)
		notes.Statistics = calculateStatistics(notes.Commits)
	}

	fmt.Printf("✅ Found %d commits\n", notes.Statistics.Total)

	formatter := git.NewFormatter(format)
	formatted, err := formatter.Format(notes)
	if err != nil {
		return fmt.Errorf("failed to format release notes: %w", err)
	}

	// Always save to work directory with a default name
	defaultFileName := fmt.Sprintf("release-notes-%s.%s", time.Now().Format("20060102-150405"), getFileExtension(outputFormat))
	workFilePath := filepath.Join(releaseNotesWorkDir, defaultFileName)
	if err := os.WriteFile(workFilePath, []byte(formatted), 0644); err != nil {
		return fmt.Errorf("failed to write file to work directory: %w", err)
	}
	fmt.Printf("💾 Release notes saved to work directory: %s\n", workFilePath)

	if releaseOutput != "" {
		// If a specific output file is specified, also write there and show content
		if releaseOutput != workFilePath {
			if err := os.WriteFile(releaseOutput, []byte(formatted), 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			fmt.Printf("💾 Release notes also saved to: %s\n", releaseOutput)
		}
		// Show content on stdout when --output is specified
		fmt.Println(formatted)
	}
	// If no --output is specified, don't show content on stdout

	return nil
}

// extractRepoName extracts the repository name from a Git URL
func extractRepoName(url string) string {
	// Remove .git extension if present
	if len(url) > 4 && url[len(url)-4:] == ".git" {
		url = url[:len(url)-4]
	}

	// Extract the last part of the path
	parts := filepath.SplitList(url)
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		// Handle URLs with slashes
		if idx := filepath.Base(lastPart); idx != "" {
			return idx
		}
		return lastPart
	}

	return "unknown-repo"
}

// getFileExtension returns the appropriate file extension for the output format
func getFileExtension(format string) string {
	switch format {
	case "markdown":
		return "md"
	case "json":
		return "json"
	case "text":
		return "txt"
	default:
		return "md"
	}
}

// filterCommits filters commits based on include and exclude types
func filterCommits(commits []git.CommitInfo, includeTypes, excludeTypes []string) []git.CommitInfo {
	if len(includeTypes) == 0 && len(excludeTypes) == 0 {
		return commits
	}

	var filtered []git.CommitInfo

	for _, commit := range commits {
		// Check if commit type should be included
		if len(includeTypes) > 0 {
			included := false
			for _, includeType := range includeTypes {
				if string(commit.Type) == includeType {
					included = true
					break
				}
			}
			if !included {
				continue
			}
		}

		// Check if commit type should be excluded
		if len(excludeTypes) > 0 {
			excluded := false
			for _, excludeType := range excludeTypes {
				if string(commit.Type) == excludeType {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}

		filtered = append(filtered, commit)
	}

	return filtered
}

// calculateStatistics recalculates statistics for filtered commits
func calculateStatistics(commits []git.CommitInfo) git.CommitStatistics {
	stats := git.CommitStatistics{}

	for _, commit := range commits {
		stats.Total++
		switch commit.Type {
		case git.CommitTypeFeature:
			stats.Features++
		case git.CommitTypeFix:
			stats.Fixes++
		case git.CommitTypeDocs:
			stats.Docs++
		case git.CommitTypeStyle:
			stats.Style++
		case git.CommitTypeRefactor:
			stats.Refactor++
		case git.CommitTypeTest:
			stats.Test++
		case git.CommitTypeChore:
			stats.Chore++
		case git.CommitTypeBreaking:
			stats.Breaking++
		default:
			stats.Unknown++
		}
	}

	return stats
}
