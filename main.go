package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mclellac/hurl/config"
	"github.com/mclellac/hurl/display"
	"github.com/mclellac/hurl/flagvar" // Import is now used correctly
	"github.com/mclellac/hurl/network"
)

func main() {
	// Use the exported type HeaderFlags from the flagvar package.
	var customHeaders flagvar.HeaderFlags // Corrected: Use exported type name

	methodPtr := flag.String("X", "GET", "HTTP method (GET, POST, PUT, DELETE, etc.)")
	flag.StringVar(methodPtr, "request", "GET", "HTTP method (GET, POST, PUT, DELETE, etc.)")
	flag.Var(&customHeaders, "H", "Add custom request header (e.g., \"Content-Type: application/json\") (can be used multiple times)")
	flag.Var(&customHeaders, "header", "Add custom request header (e.g., \"Content-Type: application/json\") (can be used multiple times)")

	insecurePtr := flag.Bool("k", false, "Allow insecure server connections when using SSL")
	flag.BoolVar(insecurePtr, "insecure", false, "Allow insecure server connections when using SSL")

	noRedirectPtr := flag.Bool("no-redirect", false, "Do not follow HTTP redirects (HTTP 3xx)")
	akamaiPragmaPtr := flag.Bool("akamai-pragma", false, "Send Akamai Pragma debug headers")

	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <URL>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s -X POST -H \"Content-Type: application/json\" https://httpbin.org/post\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	url := flag.Arg(0)
	method := strings.ToUpper(*methodPtr)
	followRedirects := !(*noRedirectPtr)

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
		CustomHeaders:   customHeaders.Get(), // Get() method still works
		InsecureSkipTLS: *insecurePtr,
		FollowRedirects: followRedirects,
		AddAkamaiPragma: *akamaiPragmaPtr,
	}

	resp, err := network.Fetch(reqOptions)
	// err stored from Fetch is checked later

	if resp != nil {
		defer resp.Body.Close()
	}

	// Check error from Fetch *after* attempting Close() via defer
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing request: %v\n", err)
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