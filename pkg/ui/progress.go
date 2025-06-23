package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressBar represents a progress bar model
type ProgressBar struct {
	width     int
	progress  float64
	message   string
	startTime time.Time
	completed bool
	error     error
	frame     int
}

// ProgressMsg represents a progress update message
type ProgressMsg struct {
	Progress float64
	Message  string
}

// CompleteMsg represents a completion message
type CompleteMsg struct {
	Error error
}

var animatedFrames = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

// Init initializes the progress bar
func (p ProgressBar) Init() tea.Cmd {
	return tickAnim()
}

func tickAnim() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return t
	})
}

// Update handles progress bar updates
func (p ProgressBar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return p, tea.Quit
		}
	case ProgressMsg:
		p.progress = msg.Progress
		p.message = msg.Message
	case CompleteMsg:
		p.completed = true
		p.error = msg.Error
		return p, tea.Quit
	case tea.WindowSizeMsg:
		p.width = msg.Width
	case time.Time:
		p.frame = (p.frame + 1) % len(animatedFrames)
		if !p.completed {
			return p, tickAnim()
		}
	}
	return p, nil
}

// View renders the progress bar
func (p ProgressBar) View() string {
	if p.completed {
		if p.error != nil {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ff6b6b")).
				Render(fmt.Sprintf("❌ Error: %s", p.error.Error()))
		}
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#51cf66")).
			Render("✅ Completed successfully")
	}

	barWidth := 30 // Fixed width for consistency
	filled := int(float64(barWidth) * p.progress)
	empty := barWidth - filled

	// Animated spinner
	spinner := animatedFrames[p.frame]

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	barStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#74c0fc")).
		Bold(true)

	progressText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#868e96")).
		Render(fmt.Sprintf("%3.0f%%", p.progress*100))

	// Simple one-line progress bar with spacing
	return fmt.Sprintf("\n\n%s %s %s", spinner, barStyle.Render(bar), progressText)
}

// NewProgressBar creates a new progress bar
func NewProgressBar() ProgressBar {
	return ProgressBar{
		width:     80,
		progress:  0,
		message:   "Initializing...",
		startTime: time.Now(),
		completed: false,
		frame:     0,
	}
}

// RunProgressBar runs the progress bar with the given program
func RunProgressBar(program *tea.Program) error {
	_, err := program.Run()
	return err
}
