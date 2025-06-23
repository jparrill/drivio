package main

import (
	"log"
	"os"

	"drivio/pkg/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
