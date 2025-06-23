package git

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CommitType represents the type of a commit
type CommitType string

const (
	CommitTypeFeature  CommitType = "feat"
	CommitTypeFix      CommitType = "fix"
	CommitTypeDocs     CommitType = "docs"
	CommitTypeStyle    CommitType = "style"
	CommitTypeRefactor CommitType = "refactor"
	CommitTypeTest     CommitType = "test"
	CommitTypeChore    CommitType = "chore"
	CommitTypeBreaking CommitType = "breaking"
	CommitTypeUnknown  CommitType = "unknown"
)

// CommitInfo represents information about a commit
type CommitInfo struct {
	Hash     string
	Author   string
	Email    string
	Date     time.Time
	Message  string
	Type     CommitType
	Scope    string
	Subject  string
	Body     string
	Breaking bool
	Footer   string
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
	Total    int
	Features int
	Fixes    int
	Docs     int
	Style    int
	Refactor int
	Test     int
	Chore    int
	Breaking int
	Unknown  int
}

// Analyzer represents a Git commit analyzer
type Analyzer struct {
	repo *git.Repository
}

// NewAnalyzer creates a new Git commit analyzer
func NewAnalyzer(repoPath string) (*Analyzer, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	return &Analyzer{
		repo: repo,
	}, nil
}

// GenerateReleaseNotes generates release notes between two references
func (a *Analyzer) GenerateReleaseNotes(fromRef, toRef string) (*ReleaseNotes, error) {
	// Get the commits
	fromHash, err := a.resolveReference(fromRef)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve from reference '%s': %w", fromRef, err)
	}

	toHash, err := a.resolveReference(toRef)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve to reference '%s': %w", toRef, err)
	}

	// Get commits between the two references
	commits, err := a.getCommitsBetween(fromHash, toHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits between references: %w", err)
	}

	// Analyze commits
	analyzedCommits := make([]CommitInfo, 0, len(commits))
	stats := CommitStatistics{}

	for _, commit := range commits {
		commitInfo := a.analyzeCommit(commit)
		// Filtrar los commits cuyo subject empiece por 'Merge pull request'
		if strings.HasPrefix(commitInfo.Subject, "Merge pull request") {
			continue
		}
		analyzedCommits = append(analyzedCommits, commitInfo)

		// Update statistics
		stats.Total++
		switch commitInfo.Type {
		case CommitTypeFeature:
			stats.Features++
		case CommitTypeFix:
			stats.Fixes++
		case CommitTypeDocs:
			stats.Docs++
		case CommitTypeStyle:
			stats.Style++
		case CommitTypeRefactor:
			stats.Refactor++
		case CommitTypeTest:
			stats.Test++
		case CommitTypeChore:
			stats.Chore++
		case CommitTypeBreaking:
			stats.Breaking++
		default:
			stats.Unknown++
		}
	}

	return &ReleaseNotes{
		FromRef:     fromRef,
		ToRef:       toRef,
		FromHash:    fromHash.String(),
		ToHash:      toHash.String(),
		GeneratedAt: time.Now(),
		Commits:     analyzedCommits,
		Statistics:  stats,
	}, nil
}

// resolveReference resolves a reference to a commit hash
func (a *Analyzer) resolveReference(ref string) (plumbing.Hash, error) {
	// Try to resolve as a reference first
	reference, err := a.repo.Reference(plumbing.ReferenceName(ref), true)
	if err == nil {
		return reference.Hash(), nil
	}

	// Try to resolve as a tag
	tag, err := a.repo.Tag(ref)
	if err == nil {
		return tag.Hash(), nil
	}

	// Try to resolve as a commit hash
	if len(ref) >= 7 {
		hash := plumbing.NewHash(ref)
		if hash != plumbing.ZeroHash {
			return hash, nil
		}
	}

	return plumbing.ZeroHash, fmt.Errorf("could not resolve reference: %s", ref)
}

// getCommitsBetween gets all commits between two hashes
func (a *Analyzer) getCommitsBetween(from, to plumbing.Hash) ([]*object.Commit, error) {
	var commits []*object.Commit

	// Create an iterator for commits from 'to' to 'from'
	commitIter, err := a.repo.Log(&git.LogOptions{
		From: to,
	})
	if err != nil {
		return nil, err
	}
	defer commitIter.Close()

	// Collect commits until we reach the 'from' commit
	err = commitIter.ForEach(func(commit *object.Commit) error {
		// Stop when we reach the 'from' commit
		if commit.Hash == from {
			return fmt.Errorf("found from commit")
		}

		// Add the commit to our list
		commits = append(commits, commit)
		return nil
	})

	// Check if we stopped because we found the 'from' commit
	if err != nil && err.Error() != "found from commit" {
		return nil, err
	}

	// Reverse the commits to get them in chronological order
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}

	return commits, nil
}

// analyzeCommit analyzes a single commit and extracts information
func (a *Analyzer) analyzeCommit(commit *object.Commit) CommitInfo {
	info := CommitInfo{
		Hash:    commit.Hash.String(),
		Author:  commit.Author.Name,
		Email:   commit.Author.Email,
		Date:    commit.Author.When,
		Message: commit.Message,
	}

	// Parse conventional commit format
	info.Type, info.Scope, info.Subject, info.Body, info.Footer, info.Breaking = a.parseConventionalCommit(commit.Message)

	return info
}

// parseConventionalCommit parses a conventional commit message
func (a *Analyzer) parseConventionalCommit(message string) (CommitType, string, string, string, string, bool) {
	lines := strings.Split(message, "\n")
	if len(lines) == 0 {
		return CommitTypeUnknown, "", "", "", "", false
	}

	// Parse the first line (header)
	header := strings.TrimSpace(lines[0])

	// Conventional commit regex: type(scope): description
	re := regexp.MustCompile(`^(\w+)(?:\(([^)]+)\))?:\s*(.+)$`)
	matches := re.FindStringSubmatch(header)

	if len(matches) < 4 {
		return CommitTypeUnknown, "", header, "", "", false
	}

	commitType := CommitType(strings.ToLower(matches[1]))
	scope := matches[2]
	subject := matches[3]

	// Check for breaking changes
	breaking := false
	if strings.Contains(strings.ToLower(header), "breaking change") {
		breaking = true
		commitType = CommitTypeBreaking
	}

	// Parse body and footer
	var body, footer strings.Builder
	inFooter := false

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if line == "" {
			continue
		}

		// Check if this is a footer line
		if strings.Contains(line, ":") && !inFooter {
			// Check if previous line was empty (footer separator)
			if i > 1 && strings.TrimSpace(lines[i-1]) == "" {
				inFooter = true
			}
		}

		if inFooter {
			if footer.Len() > 0 {
				footer.WriteString("\n")
			}
			footer.WriteString(line)
		} else {
			if body.Len() > 0 {
				body.WriteString("\n")
			}
			body.WriteString(line)
		}
	}

	// Check for breaking changes in body or footer
	if strings.Contains(strings.ToLower(body.String()), "breaking change") ||
		strings.Contains(strings.ToLower(footer.String()), "breaking change") {
		breaking = true
		if commitType != CommitTypeBreaking {
			commitType = CommitTypeBreaking
		}
	}

	return commitType, scope, subject, body.String(), footer.String(), breaking
}
