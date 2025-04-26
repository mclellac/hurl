// main.go
package main

import (
	"fmt"
	"hurl/display"
	"hurl/network"
	"os"

	"hurl/config" // Import local packages
)

func main() {
	// --- Argument Parsing ---
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <URL>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s https://www.example.com\n", os.Args[0])
		os.Exit(1)
	}
	url := os.Args[1]

	// --- Configuration ---
	// Ensure config directory exists (optional, good for first run)
	err := config.EnsureConfigDir()
	if err != nil {
		// Non-fatal, just warn
		fmt.Fprintf(os.Stderr, "Warning: Could not ensure config directory: %v\n", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		// LoadConfig already prints warnings, but we could exit if needed
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v. Exiting.\n", err)
		os.Exit(1) // Or decide to proceed with defaults if LoadConfig returns them
	}

	// --- HTTP Request ---
	resp, err := network.FetchHeaders(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", err)
		os.Exit(1)
	}
	// CRITICAL: Always ensure the response body is closed to free resources,
	// even if you don't read from it.
	defer resp.Body.Close()

	// --- Display Headers ---
	// Optionally print the status line first for context
	fmt.Printf("%s%s %s%s\n",
		config.GetAnsiCode(cfg.HeaderValueColor), // Use value color for status? Or a dedicated one?
		resp.Proto,
		resp.Status,
		config.ColorReset)

	display.PrintHeaders(resp.Header, cfg)

	// --- Exit ---
	// Check status code for basic success indication (optional)
	if resp.StatusCode >= 400 {
		// Optionally exit with non-zero status for client/server errors
		// os.Exit(2) // Or some other non-zero code
	}
}
