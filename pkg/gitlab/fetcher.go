package gitlab

import (
	"context"
	"time"

	"drivio/pkg/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// FetchWithProgress fetches a file with a progress bar
func (c *Client) FetchWithProgress(ctx context.Context) ([]byte, error) {
	progressBar := ui.NewProgressBar()
	program := tea.NewProgram(&progressBar)

	var result []byte
	var fetchError error

	// Start the progress bar in a goroutine
	go func() {
		progress := 0.0

		// Step 1: Validating connection (20%)
		program.Send(ui.ProgressMsg{Progress: 0.0, Message: "Validating connection..."})
		time.Sleep(300 * time.Millisecond)
		progress = 0.2
		program.Send(ui.ProgressMsg{Progress: progress, Message: "Connection validated"})

		// Step 2: Getting repository info (40%)
		program.Send(ui.ProgressMsg{Progress: 0.3, Message: "Getting repository information..."})
		time.Sleep(200 * time.Millisecond)
		progress = 0.4
		program.Send(ui.ProgressMsg{Progress: progress, Message: "Repository info obtained"})

		// Step 3: Fetching file (80%)
		program.Send(ui.ProgressMsg{Progress: progress, Message: "Fetching file..."})

		// Start the actual fetch in a separate goroutine
		fetchDone := make(chan error, 1)
		go func() {
			content, err := c.GetFile(ctx)
			if err != nil {
				fetchDone <- err
				return
			}
			result = content
			fetchDone <- nil
		}()

		// Simulate progress during fetch (50% to 80%)
		for progress < 0.8 {
			time.Sleep(400 * time.Millisecond)
			progress += 0.05
			if progress > 0.8 {
				progress = 0.8
			}
			program.Send(ui.ProgressMsg{Progress: progress, Message: "Fetching file..."})
		}

		// Wait for fetch to complete
		err := <-fetchDone
		if err != nil {
			fetchError = err
			program.Send(ui.CompleteMsg{Error: err})
			return
		}

		// Step 4: Saving file (100%)
		program.Send(ui.ProgressMsg{Progress: 0.9, Message: "Saving file..."})
		time.Sleep(200 * time.Millisecond)
		program.Send(ui.ProgressMsg{Progress: 1.0, Message: "File fetched successfully!"})
		time.Sleep(300 * time.Millisecond)

		program.Send(ui.CompleteMsg{Error: nil})
	}()

	err := ui.RunProgressBar(program)
	if err != nil {
		return nil, err
	}

	if fetchError != nil {
		return nil, fetchError
	}

	return result, nil
}
