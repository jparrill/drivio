package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	cleanWorkDir string
	cleanForce   bool
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up the work directory",
	Long: `Clean up the work directory by removing all downloaded files and cloned repositories.

This command removes all files and directories in the work directory to free up disk space.

Examples:
  drivio clean
  drivio clean --work-dir /tmp/drivio-work
  drivio clean --force`,
	RunE: runClean,
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	// Add flags
	cleanCmd.Flags().StringVar(&cleanWorkDir, "work-dir", ".drivio-work", "Working directory to clean")
	cleanCmd.Flags().BoolVar(&cleanForce, "force", false, "Force cleanup without confirmation")
}

func runClean(cmd *cobra.Command, args []string) error {
	// Check if work directory exists
	if _, err := os.Stat(cleanWorkDir); os.IsNotExist(err) {
		fmt.Printf("📁 Work directory does not exist: %s\n", cleanWorkDir)
		return nil
	}

	// Get directory contents for confirmation
	entries, err := os.ReadDir(cleanWorkDir)
	if err != nil {
		return fmt.Errorf("failed to read work directory: %w", err)
	}

	if len(entries) == 0 {
		fmt.Printf("📁 Work directory is already empty: %s\n", cleanWorkDir)
		return nil
	}

	// Show what will be deleted
	fmt.Printf("📁 Work directory: %s\n", cleanWorkDir)
	fmt.Printf("🗑️  Found %d items to clean:\n", len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			fmt.Printf("  - %s (error getting info)\n", entry.Name())
		} else {
			if entry.IsDir() {
				fmt.Printf("  - 📁 %s (directory)\n", entry.Name())
			} else {
				fmt.Printf("  - 📄 %s (%d bytes)\n", entry.Name(), info.Size())
			}
		}
	}

	// Ask for confirmation unless --force is used
	if !cleanForce {
		fmt.Print("\n❓ Are you sure you want to delete all these files? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("❌ Cleanup cancelled")
			return nil
		}
	}

	// Remove all contents
	for _, entry := range entries {
		entryPath := filepath.Join(cleanWorkDir, entry.Name())
		if err := os.RemoveAll(entryPath); err != nil {
			fmt.Printf("⚠️  Warning: failed to remove %s: %v\n", entryPath, err)
		} else {
			fmt.Printf("✅ Removed: %s\n", entry.Name())
		}
	}

	fmt.Printf("🧹 Cleanup completed for: %s\n", cleanWorkDir)
	return nil
}
