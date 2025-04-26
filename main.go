package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mclellac/hurl/config"
	"github.com/mclellac/hurl/display"
	"github.com/mclellac/hurl/flagvar"
	"github.com/mclellac/hurl/network"
)

func main() {
	var customHeaders flagvar.HeaderFlags
	// Flags definition
	methodPtr := flag.String("X", "GET", "HTTP request method")
	flag.StringVar(methodPtr, "request", "GET", "HTTP request method") // Alias

	flag.Var(&customHeaders, "H", "Add custom request header (e.g., \"Key: Value\")")
	flag.Var(&customHeaders, "header", "Add custom request header (e.g., \"Key: Value\")") // Alias

	insecurePtr := flag.Bool("k", false, "Allow insecure server connections")
	flag.BoolVar(insecurePtr, "insecure", false, "Allow insecure server connections") // Alias

	locationPtr := flag.Bool("L", false, "Follow redirects (HTTP 3xx)") // NEW: -L flag
	flag.BoolVar(locationPtr, "location", false, "Follow redirects (HTTP 3xx)") // Alias
	// Removed --no-redirect flag

	headPtr := flag.Bool("I", false, "Perform HTTP HEAD request (overrides -X)") // NEW: -I flag
	flag.BoolVar(headPtr, "head", false, "Perform HTTP HEAD request (overrides -X)") // Alias

	akamaiPragmaPtr := flag.Bool("akamai-pragma", false, "Send Akamai Pragma debug headers")
	verbosePtr := flag.Bool("v", false, "Make the operation more talkative")
	flag.BoolVar(verbosePtr, "verbose", false, "Make the operation more talkative")

	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <URL>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s -I https://www.example.com\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s -L http://httpbin.org/redirect/1\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	url := flag.Arg(0)

	// Determine method: Use -X unless -I is specified
	method := strings.ToUpper(*methodPtr)
	if *headPtr {
		method = "HEAD" // -I overrides -X
	}

	// Determine redirect policy: Follow only if -L is set
	followRedirects := *locationPtr // Direct mapping now

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
		FollowRedirects: followRedirects, // Updated logic
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