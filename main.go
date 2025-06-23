package main

import (
	"log"
	"os"

	"drivio/pkg/cmd"
)

// Version information - these will be set by the build process
var (
	Version    = "0.1.0"
	CommitHash = "unknown"
	BuildTime  = "unknown"
)

func main() {
	// Set version information in the cmd package
	cmd.Version = Version
	cmd.CommitHash = CommitHash
	cmd.BuildTime = BuildTime

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
