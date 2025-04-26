package main

import (
	// Use pflag instead of the standard flag package
	flag "github.com/spf13/pflag"
	"fmt"
	"os"
	"strings"

	"github.com/mclellac/hurl/config"
	"github.com/mclellac/hurl/display"
	"github.com/mclellac/hurl/flagvar"
	"github.com/mclellac/hurl/network"
)

func main() {
	// Define flags using pflag
	var customHeaders flagvar.HeaderFlags

	// Use pflag's "P" variants to define both long and short flags together
	methodPtr := flag.StringP("request", "X", "GET", "HTTP request method")
	flag.VarP(&customHeaders, "header", "H", "Add custom request header (e.g., \"Key: Value\")")
	insecurePtr := flag.BoolP("insecure", "k", false, "Allow insecure server connections")
	locationPtr := flag.BoolP("location", "L", false, "Follow redirects (HTTP 3xx)")
	headPtr := flag.BoolP("head", "I", false, "Perform HTTP HEAD request (overrides -X)")
	verbosePtr := flag.BoolP("verbose", "v", false, "Make the operation more talkative")

	// Flags without short versions remain the same
	akamaiPragmaPtr := flag.Bool("akamai-pragma", false, "Send Akamai Pragma debug headers")

	// pflag handles --help/-h automatically and correctly formats Usage
	flag.Usage = func() {
		// Custom usage message format
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <URL>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s -I https://www.example.com\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s -L http://httpbin.org/redirect/1\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults() // pflag's PrintDefaults formats correctly
	}

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage() // Print the usage message on error
		os.Exit(1)
	}
	url := flag.Arg(0)

	method := strings.ToUpper(*methodPtr)
	if *headPtr {
		method = "HEAD"
	}
	followRedirects := *locationPtr

	err := config.EnsureConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not ensure config directory: %v\n", err)
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v. Exiting.\n", err)
		os.Exit(1)
	}

	reqOptions := network.RequestOptions{
		Method:          method,
		URL:             url,
		CustomHeaders:   customHeaders.Get(),
		InsecureSkipTLS: *insecurePtr,
		FollowRedirects: followRedirects,
		AddAkamaiPragma: *akamaiPragmaPtr,
		Verbose:         *verbosePtr,
		Config:          cfg,
	}

	resp, err := network.Fetch(reqOptions)

	if resp != nil {
		defer resp.Body.Close()
	}

	// Check error from Fetch *after* attempting Close() via defer
	if err != nil {
		if !reqOptions.Verbose {
			fmt.Fprintf(os.Stderr, "%sError executing request: %v%s\n", config.ColorRed, err, config.ColorReset)
		}
		os.Exit(1)
	}

	if !reqOptions.Verbose {
		fmt.Printf("%s%s %s%s\n",
			config.GetAnsiCode(cfg.HeaderValueColor),
			resp.Proto,
			resp.Status,
			config.ColorReset)

		display.PrintHeaders(os.Stdout, resp.Header, cfg)
	}

	if resp.StatusCode >= 400 {
		// os.Exit(2) // Optional: exit non-zero for >= 400 status codes
	}
}