package git

import (
	"fmt"
	"strings"
	"text/template"
	"time"
)

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

// NewFormatter creates a new formatter
func NewFormatter(format OutputFormat) *Formatter {
	return &Formatter{
		format: format,
	}
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
	const markdownTemplate = `# Release Notes

**From:** {{.FromRef}} ({{.FromHash}})
**To:** {{.ToRef}} ({{.ToHash}})
**Generated:** {{.GeneratedAt.Format "2006-01-02 15:04:05"}}

## Summary

- **Total Commits:** {{.Statistics.Total}}
- **Features:** {{.Statistics.Features}}
- **Bug Fixes:** {{.Statistics.Fixes}}
- **Documentation:** {{.Statistics.Docs}}
- **Refactoring:** {{.Statistics.Refactor}}
- **Tests:** {{.Statistics.Test}}
- **Chores:** {{.Statistics.Chore}}
- **Breaking Changes:** {{.Statistics.Breaking}}
- **Other:** {{.Statistics.Unknown}}

{{if gt .Statistics.Breaking 0}}
## ⚠️ Breaking Changes

{{range .Commits}}
{{if eq .Type "breaking"}}
- **{{.Subject}}** ({{.Hash}})
  {{if .Body}}
  {{.Body}}
  {{end}}
{{end}}
{{end}}

{{end}}
{{if gt .Statistics.Features 0}}
## ✨ Features

{{range .Commits}}
{{if eq .Type "feat"}}
- **{{.Subject}}** ({{.Hash}})
  {{if .Body}}
  {{.Body}}
  {{end}}
{{end}}
{{end}}

{{end}}
{{if gt .Statistics.Fixes 0}}
## 🐛 Bug Fixes

{{range .Commits}}
{{if eq .Type "fix"}}
- **{{.Subject}}** ({{.Hash}})
  {{if .Body}}
  {{.Body}}
  {{end}}
{{end}}
{{end}}

{{end}}
{{if gt .Statistics.Docs 0}}
## 📚 Documentation

{{range .Commits}}
{{if eq .Type "docs"}}
- **{{.Subject}}** ({{.Hash}})
  {{if .Body}}
  {{.Body}}
  {{end}}
{{end}}
{{end}}

{{end}}
{{if gt .Statistics.Refactor 0}}
## 🔧 Refactoring

{{range .Commits}}
{{if eq .Type "refactor"}}
- **{{.Subject}}** ({{.Hash}})
  {{if .Body}}
  {{.Body}}
  {{end}}
{{end}}
{{end}}

{{end}}
{{if gt .Statistics.Test 0}}
## 🧪 Tests

{{range .Commits}}
{{if eq .Type "test"}}
- **{{.Subject}}** ({{.Hash}})
  {{if .Body}}
  {{.Body}}
  {{end}}
{{end}}
{{end}}

{{end}}
{{if gt .Statistics.Chore 0}}
## 🔨 Chores

{{range .Commits}}
{{if eq .Type "chore"}}
- **{{.Subject}}** ({{.Hash}})
  {{if .Body}}
  {{.Body}}
  {{end}}
{{end}}
{{end}}

{{end}}
{{if gt .Statistics.Unknown 0}}
## 📝 Other Changes

{{range .Commits}}
{{if eq .Type "unknown"}}
- **{{.Subject}}** ({{.Hash}})
  {{if .Body}}
  {{.Body}}
  {{end}}
{{end}}
{{end}}

{{end}}
## All Commits

{{range .Commits}}
- **{{.Type}}:** {{.Subject}} ({{.Hash}}) - {{.Author}} - {{.Date.Format "2006-01-02"}}
{{end}}
`

	tmpl, err := template.New("markdown").Parse(markdownTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse markdown template: %w", err)
	}

	var result strings.Builder
	err = tmpl.Execute(&result, notes)
	if err != nil {
		return "", fmt.Errorf("failed to execute markdown template: %w", err)
	}

	return result.String(), nil
}

// formatJSON formats release notes as JSON
func (f *Formatter) formatJSON(notes *ReleaseNotes) (string, error) {
	// Simple JSON formatting - in a real implementation, you might want to use encoding/json
	json := fmt.Sprintf(`{
  "from_ref": "%s",
  "to_ref": "%s",
  "from_hash": "%s",
  "to_hash": "%s",
  "generated_at": "%s",
  "statistics": {
    "total": %d,
    "features": %d,
    "fixes": %d,
    "docs": %d,
    "style": %d,
    "refactor": %d,
    "test": %d,
    "chore": %d,
    "breaking": %d,
    "unknown": %d
  },
  "commits": [`,
		notes.FromRef, notes.ToRef, notes.FromHash, notes.ToHash,
		notes.GeneratedAt.Format(time.RFC3339),
		notes.Statistics.Total, notes.Statistics.Features, notes.Statistics.Fixes,
		notes.Statistics.Docs, notes.Statistics.Style, notes.Statistics.Refactor,
		notes.Statistics.Test, notes.Statistics.Chore, notes.Statistics.Breaking,
		notes.Statistics.Unknown)

	for i, commit := range notes.Commits {
		if i > 0 {
			json += ","
		}
		json += fmt.Sprintf(`
    {
      "hash": "%s",
      "author": "%s",
      "email": "%s",
      "date": "%s",
      "type": "%s",
      "scope": "%s",
      "subject": "%s",
      "body": "%s",
      "breaking": %t,
      "footer": "%s"
    }`,
			commit.Hash, commit.Author, commit.Email, commit.Date.Format(time.RFC3339),
			commit.Type, commit.Scope, commit.Subject, strings.ReplaceAll(commit.Body, `"`, `\"`),
			commit.Breaking, strings.ReplaceAll(commit.Footer, `"`, `\"`))
	}

	json += `
  ]
}`

	return json, nil
}

// formatText formats release notes as plain text
func (f *Formatter) formatText(notes *ReleaseNotes) (string, error) {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Release Notes\n"))
	result.WriteString(fmt.Sprintf("=============\n\n"))
	result.WriteString(fmt.Sprintf("From: %s (%s)\n", notes.FromRef, notes.FromHash))
	result.WriteString(fmt.Sprintf("To: %s (%s)\n", notes.ToRef, notes.ToHash))
	result.WriteString(fmt.Sprintf("Generated: %s\n\n", notes.GeneratedAt.Format("2006-01-02 15:04:05")))

	result.WriteString(fmt.Sprintf("Summary:\n"))
	result.WriteString(fmt.Sprintf("- Total Commits: %d\n", notes.Statistics.Total))
	result.WriteString(fmt.Sprintf("- Features: %d\n", notes.Statistics.Features))
	result.WriteString(fmt.Sprintf("- Bug Fixes: %d\n", notes.Statistics.Fixes))
	result.WriteString(fmt.Sprintf("- Documentation: %d\n", notes.Statistics.Docs))
	result.WriteString(fmt.Sprintf("- Refactoring: %d\n", notes.Statistics.Refactor))
	result.WriteString(fmt.Sprintf("- Tests: %d\n", notes.Statistics.Test))
	result.WriteString(fmt.Sprintf("- Chores: %d\n", notes.Statistics.Chore))
	result.WriteString(fmt.Sprintf("- Breaking Changes: %d\n", notes.Statistics.Breaking))
	result.WriteString(fmt.Sprintf("- Other: %d\n\n", notes.Statistics.Unknown))

	// Group commits by type
	commitsByType := make(map[CommitType][]CommitInfo)
	for _, commit := range notes.Commits {
		commitsByType[commit.Type] = append(commitsByType[commit.Type], commit)
	}

	// Output commits by type
	typeOrder := []CommitType{CommitTypeBreaking, CommitTypeFeature, CommitTypeFix, CommitTypeDocs, CommitTypeRefactor, CommitTypeTest, CommitTypeChore, CommitTypeUnknown}

	for _, commitType := range typeOrder {
		if commits, exists := commitsByType[commitType]; exists && len(commits) > 0 {
			result.WriteString(fmt.Sprintf("%s:\n", strings.ToUpper(string(commitType))))
			for _, commit := range commits {
				result.WriteString(fmt.Sprintf("- %s (%s) - %s\n", commit.Subject, commit.Hash, commit.Author))
				if commit.Body != "" {
					result.WriteString(fmt.Sprintf("  %s\n", commit.Body))
				}
			}
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}
