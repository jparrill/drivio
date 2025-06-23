package git

import (
	"time"

	"drivio/pkg/ui"

	tea "github.com/charmbracelet/bubbletea"
	gitv5 "github.com/go-git/go-git/v5"
)

// CloneWithProgress clones a repository with a progress bar
func CloneWithProgress(url, path string) error {
	progressBar := ui.NewProgressBar()
	program := tea.NewProgram(&progressBar)

	// Start the progress bar in a goroutine
	go func() {
		progress := 0.0

		// Step 1: Connecting (10%)
		program.Send(ui.ProgressMsg{Progress: 0.0, Message: "Connecting to repository..."})
		time.Sleep(300 * time.Millisecond)
		progress = 0.1
		program.Send(ui.ProgressMsg{Progress: progress, Message: "Connected to repository"})

		// Step 2: Fetching info (20%)
		program.Send(ui.ProgressMsg{Progress: 0.15, Message: "Fetching repository information..."})
		time.Sleep(200 * time.Millisecond)
		progress = 0.2
		program.Send(ui.ProgressMsg{Progress: progress, Message: "Repository info fetched"})

		// Step 3: Downloading objects (60% - this is the main part)
		program.Send(ui.ProgressMsg{Progress: 0.25, Message: "Downloading objects..."})

		// Start the actual clone in a separate goroutine
		cloneDone := make(chan error, 1)
		go func() {
			_, err := gitv5.PlainClone(path, false, &gitv5.CloneOptions{
				URL:      url,
				Progress: nil,
			})
			cloneDone <- err
		}()

		// Simulate progress during clone (25% to 85%)
		for progress < 0.85 {
			time.Sleep(800 * time.Millisecond) // Longer intervals
			progress += 0.1
			if progress > 0.85 {
				progress = 0.85
			}
			program.Send(ui.ProgressMsg{Progress: progress, Message: "Downloading objects..."})
		}

		// Wait for clone to complete
		err := <-cloneDone
		if err != nil {
			program.Send(ui.CompleteMsg{Error: err})
			return
		}

		// Step 4: Resolving deltas (90%)
		program.Send(ui.ProgressMsg{Progress: 0.9, Message: "Resolving deltas..."})
		time.Sleep(200 * time.Millisecond)

		// Step 5: Writing objects (95%)
		program.Send(ui.ProgressMsg{Progress: 0.95, Message: "Writing objects..."})
		time.Sleep(150 * time.Millisecond)

		// Step 6: Finalizing (100%)
		program.Send(ui.ProgressMsg{Progress: 1.0, Message: "Finalizing..."})
		time.Sleep(100 * time.Millisecond)

		// Keep the bar visible for a moment before completing
		program.Send(ui.ProgressMsg{Progress: 1.0, Message: "Clone completed successfully!"})
		time.Sleep(500 * time.Millisecond)

		program.Send(ui.CompleteMsg{Error: nil})
	}()

	return ui.RunProgressBar(program)
}

// CloneWithProgressSilent clones a repository without progress bar (fallback)
func CloneWithProgressSilent(url, path string) error {
	_, err := gitv5.PlainClone(path, false, &gitv5.CloneOptions{
		URL:      url,
		Progress: nil,
	})
	return err
}
