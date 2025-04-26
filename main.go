package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mclellac/hurl/config"
	"github.com/mclellac/hurl/display"
	"github.com/mclellac/hurl/network"
)

func main() {
	akamaiPragmaPtr := flag.Bool("akamai-pragma", false, "Send Akamai Pragma debug headers")

	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--akamai-pragma] <URL>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s https://www.example.com\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s --akamai-pragma https://www.akamai.com\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	url := flag.Arg(0)

	err := config.EnsureConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not ensure config directory: %v\n", err)
	}

	cfg, err := config.LoadConfig()
	if err != err {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v. Exiting.\n", err)
		os.Exit(1)
	}

	resp, err := network.FetchHeaders(url, *akamaiPragmaPtr)
	// err stored from FetchHeaders is checked later

	if resp != nil {
		defer resp.Body.Close()
	}

	// Check error from FetchHeaders *after* attempting Close() via defer
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s%s %s%s\n",
		config.GetAnsiCode(cfg.HeaderValueColor),
		resp.Proto,
		resp.Status,
		config.ColorReset)

	display.PrintHeaders(resp.Header, cfg)

	if resp.StatusCode >= 400 {
		// os.Exit(2) // Optional: exit non-zero for >= 400 status codes
	}
}

