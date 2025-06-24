package ui

import (
	"fmt"
	"time"
)

// SpinnerFrames contains the frames for simple animation
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// RunSpinner runs a simple spinner with a message until the task completes
func RunSpinner(message string, task func() error) error {
	done := make(chan error, 1)
	go func() {
		done <- task()
	}()
	frame := 0
	for {
		select {
		case err := <-done:
			if err != nil {
				fmt.Printf("\r❌ %s\n", message)
				return err
			}
			fmt.Printf("\r✅ %s\n", message)
			return nil
		default:
			fmt.Printf("\r%s %s", SpinnerFrames[frame], message)
			frame = (frame + 1) % len(SpinnerFrames)
			time.Sleep(100 * time.Millisecond)
		}
	}
}
