package git

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CommitInfo represents information about a commit
type CommitInfo struct {
	Hash    string
	Author  string
	Email   string
	Date    time.Time
	Subject string
}

// ReleaseNotes represents the generated release notes
type ReleaseNotes struct {
	FromRef     string
	ToRef       string
	FromHash    string
	ToHash      string
	GeneratedAt time.Time
	Commits     []CommitInfo
	Statistics  CommitStatistics
}

// CommitStatistics represents statistics about commits
type CommitStatistics struct {
	Total int
}

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

// OutputFormat represents the output format for release notes
type OutputFormat string

const (
	FormatMarkdown OutputFormat = "markdown"
	FormatJSON     OutputFormat = "json"
	FormatText     OutputFormat = "text"
)

// Formatter represents a release notes formatter
type Formatter struct {
	format OutputFormat
}

// NewFormatter creates a new formatter with the specified format
func NewFormatter(format OutputFormat) *Formatter {
	return &Formatter{format: format}
}

// Format formats release notes according to the specified format
func (f *Formatter) Format(notes *ReleaseNotes) (string, error) {
	switch f.format {
	case FormatMarkdown:
		return f.formatMarkdown(notes)
	case FormatJSON:
		return f.formatJSON(notes)
	case FormatText:
		return f.formatText(notes)
	default:
		return "", fmt.Errorf("unsupported format: %s", f.format)
	}
}

// formatMarkdown formats release notes as Markdown
func (f *Formatter) formatMarkdown(notes *ReleaseNotes) (string, error) {
	var sb strings.Builder

	// Formato exacto del script notes.go de HyperShift
	for _, commit := range notes.Commits {
		sb.WriteString(fmt.Sprintf("%s %s\n", commit.Hash[:8], commit.Subject))
	}

	return sb.String(), nil
}

// formatJSON formats release notes as JSON
func (f *Formatter) formatJSON(notes *ReleaseNotes) (string, error) {
	data, err := json.MarshalIndent(notes, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// formatText formats release notes as plain text
func (f *Formatter) formatText(notes *ReleaseNotes) (string, error) {
	var sb strings.Builder

	// Formato exacto del script notes.go de HyperShift
	for _, commit := range notes.Commits {
		sb.WriteString(fmt.Sprintf("%s %s\n", commit.Hash[:8], commit.Subject))
	}

	return sb.String(), nil
}

// Analyzer represents a GitHub commit analyzer
type Analyzer struct {
	client  *http.Client
	baseURL string
}

// NewAnalyzer creates a new GitHub commit analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: "https://api.github.com",
	}
}

// GenerateReleaseNotes generates release notes between two references using GitHub API
func (a *Analyzer) GenerateReleaseNotes(owner, repo, fromRef, toRef string) (*ReleaseNotes, error) {
	// Get commits between the two references using GitHub API
	commits, err := a.getCommitsBetween(owner, repo, fromRef, toRef)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits between references: %w", err)
	}

	analyzedCommits := make([]CommitInfo, 0)
	stats := CommitStatistics{}

	fmt.Printf("🔍 Found %d total commits between references\n", len(commits))

	// El API de GitHub los devuelve en orden del más antiguo al más reciente, pero lo aseguramos
	for _, commit := range commits {
		// Excluir merges
		lines := strings.Split(commit.Commit.Message, "\n")
		if len(lines) == 0 {
			continue
		}
		subject := strings.TrimSpace(lines[0])
		if strings.HasPrefix(subject, "Merge pull request") {
			continue
		}

		commitInfo := CommitInfo{
			Hash:    commit.Sha,
			Author:  commit.Commit.Author.Name,
			Email:   commit.Commit.Author.Email,
			Date:    commit.Commit.Author.Date,
			Subject: strings.TrimSpace(commit.Commit.Message), // Mensaje completo
		}
		analyzedCommits = append(analyzedCommits, commitInfo)
	}

	stats.Total = len(analyzedCommits)

	return &ReleaseNotes{
		FromRef:     fromRef,
		ToRef:       toRef,
		FromHash:    "",
		ToHash:      "",
		GeneratedAt: time.Now(),
		Commits:     analyzedCommits,
		Statistics:  stats,
	}, nil
}

// getCommitsBetween gets all commits between two references using GitHub API
func (a *Analyzer) getCommitsBetween(owner, repo, fromRef, toRef string) ([]GitHubCommit, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/compare/%s...%s", a.baseURL, owner, repo, fromRef, toRef)

	fmt.Printf("🔗 Calling GitHub API: %s\n", url)

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add headers for better rate limiting
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "drivio-release-notes")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("📡 GitHub API response status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		// Read the response body to get more details about the error
		body, _ := json.Marshal(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var compareResult struct {
		Commits []GitHubCommit `json:"commits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&compareResult); err != nil {
		return nil, err
	}

	return compareResult.Commits, nil
}
