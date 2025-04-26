package network

import (
	"context" // Import context
	"crypto/tls"
	"fmt"
	"io" // Import io
	"net/http"
	"net/http/httptrace" // Import httptrace
	"os"                 // Import os for Stderr
	"sort"               // Import sort for printing headers
	"strings"
	"time"
)

// akamaiPragmaValue is the static string used for the Akamai Pragma header.
const akamaiPragmaValue = "akamai-x-get-request-id,akamai-x-get-cache-key,akamai-x-cache-on,akamai-x-cache-remote-on,akamai-x-get-true-cache-key,akamai-x-check-cacheable,akamai-x-get-extracted-values,akamai-x-feo-trace,x-akamai-logging-mode: verbose"

// RequestOptions bundles parameters for making the HTTP request.
type RequestOptions struct {
	Method          string
	URL             string
	CustomHeaders   []string
	InsecureSkipTLS bool
	FollowRedirects bool
	AddAkamaiPragma bool
	Verbose         bool // Added verbose flag
}

// Fetch performs an HTTP request based on the provided options.
// The caller is responsible for closing the response body.
func Fetch(opts RequestOptions) (*http.Response, error) {

	tr := http.DefaultTransport.(*http.Transport).Clone()
	if tr.TLSClientConfig == nil {
		tr.TLSClientConfig = &tls.Config{}
	}
	tr.TLSClientConfig.InsecureSkipVerify = opts.InsecureSkipTLS

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}

	if !opts.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "* Ignoring redirect response from %s\n", req.URL)
			}
			return http.ErrUseLastResponse
		}
	}

	req, err := http.NewRequest(opts.Method, opts.URL, nil) // Body nil for now
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36"
	req.Header.Set("User-Agent", userAgent)

	for _, h := range opts.CustomHeaders {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" {
				req.Header.Add(key, value)
			}
		} else if len(parts) == 1 && strings.TrimSpace(parts[0]) != "" {
			// Handle headers with empty value, like "X-Custom-Flag;"
			key := strings.TrimRight(strings.TrimSpace(parts[0]), ";")
			req.Header.Add(key, "")
		}
	}

	if opts.AddAkamaiPragma {
		req.Header.Set("Pragma", akamaiPragmaValue)
	}

	var trace *httptrace.ClientTrace
	if opts.Verbose {
		trace = &httptrace.ClientTrace{
			GetConn: func(hostPort string) {
				fmt.Fprintf(os.Stderr, "* Trying %s...\n", hostPort) // Note: This might resolve to multiple IPs
			},
			DNSStart: func(info httptrace.DNSStartInfo) {
				fmt.Fprintf(os.Stderr, "* Resolving timed out or error: %s...\n", info.Host)
			},
			DNSDone: func(info httptrace.DNSDoneInfo) {
				if info.Err != nil {
					fmt.Fprintf(os.Stderr, "* Error resolving host %s: %v\n", req.URL.Host, info.Err)
					return
				}
				addrs := []string{}
				for _, ip := range info.Addrs {
					addrs = append(addrs, ip.String())
				}
				fmt.Fprintf(os.Stderr, "* Resolved %s to %v\n", req.URL.Host, addrs)
			},
			ConnectStart: func(network, addr string) {
				fmt.Fprintf(os.Stderr, "* Connecting to %s (%s)\n", addr, network)
			},
			ConnectDone: func(network, addr string, err error) {
				if err != nil {
					fmt.Fprintf(os.Stderr, "* Error connecting to %s: %v\n", addr, err)
				} else {
					fmt.Fprintf(os.Stderr, "* Connected to %s (%s)\n", addr, req.URL.Host) // Show hostname too
				}
			},
			TLSHandshakeStart: func() {
				fmt.Fprintf(os.Stderr, "* Performing TLS handshake...\n")
			},
			TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
				if err != nil {
					fmt.Fprintf(os.Stderr, "* TLS handshake error: %v\n", err)
					if cs.Version == 0 { // Handle cases where ConnectionState might be partially populated on error
						return
					}
				}
				proto := ""
				switch cs.Version {
				case tls.VersionTLS10: proto = "TLSv1.0"
				case tls.VersionTLS11: proto = "TLSv1.1"
				case tls.VersionTLS12: proto = "TLSv1.2"
				case tls.VersionTLS13: proto = "TLSv1.3"
				default: proto = fmt.Sprintf("TLS Unknown (0x%x)", cs.Version)
				}
				fmt.Fprintf(os.Stderr, "* TLS handshake complete\n")
				fmt.Fprintf(os.Stderr, "* Protocol: %s\n", proto)
				fmt.Fprintf(os.Stderr, "* Cipher Suite: %s\n", tls.CipherSuiteName(cs.CipherSuite))
				if len(cs.PeerCertificates) > 0 {
					cert := cs.PeerCertificates[0]
					fmt.Fprintf(os.Stderr, "* Server certificate:\n")
					fmt.Fprintf(os.Stderr, "* Subject: %s\n", cert.Subject.String())
					fmt.Fprintf(os.Stderr, "* Issuer: %s\n", cert.Issuer.String())
					fmt.Fprintf(os.Stderr, "* Expiry: %s\n", cert.NotAfter.Format(time.RFC1123))
					// Could add SANs etc. if needed
				}
				if cs.NegotiatedProtocol != "" {
					fmt.Fprintf(os.Stderr, "* ALPN: server accepted %s\n", cs.NegotiatedProtocol)
				}

			},
			GotConn: func(info httptrace.GotConnInfo) {
				proto := "HTTP/1.1" // Assume 1.1 unless ALPN negotiated H2
				if info.NegotiatedProtocol == "h2" {
					proto = "HTTP/2.0"
				}
				fmt.Fprintf(os.Stderr, "* Using connection to %s, protocol %s\n", info.Conn.RemoteAddr(), proto)
			},
			// WroteHeaderField is too noisy for standard -v, skip it.
			// WroteHeaders isn't granular enough. We print manually below.
			// WroteRequest is called after body write, maybe useful later.
			GotFirstResponseByte: func() {
				// This signifies TTFB (Time To First Byte)
				fmt.Fprintf(os.Stderr, "* Waiting for response headers...\n")
			},
		}
		// Attach trace to request context
		ctx := httptrace.WithClientTrace(req.Context(), trace)
		req = req.WithContext(ctx)
	}

	// Print request headers if verbose
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "> %s %s %s\n", req.Method, req.URL.RequestURI(), req.Proto)
		// Manually format Host header as Go's req.Header doesn't include it directly here
		fmt.Fprintf(os.Stderr, "> Host: %s\n", req.Host)
		printHeadersVerbose(os.Stderr, '>', req.Header) // Print other headers
		fmt.Fprintln(os.Stderr, ">")                     // Blank line like curl
	}

	resp, err := client.Do(req)

	// Print response headers if verbose *and* we got a response object
	if opts.Verbose && resp != nil {
		fmt.Fprintf(os.Stderr, "< %s %s\n", resp.Proto, resp.Status)
		printHeadersVerbose(os.Stderr, '<', resp.Header) // Print response headers
		fmt.Fprintln(os.Stderr, "<")                     // Blank line
	}

	// Handle errors *after* potentially printing verbose info
	if err != nil {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "* Request failed: %v\n", err)
		}
		// Return potentially non-nil resp even on error, caller handles Close
		return resp, fmt.Errorf("error performing request: %w", err)
	}

	// Note: Caller (main.go) is responsible for closing resp.Body
	return resp, nil
}


// printHeadersVerbose prints headers to the specified writer with a prefix.
// Headers are sorted for consistent output. Used only for verbose mode.
func printHeadersVerbose(w io.Writer, prefix rune, headers http.Header) {
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		values := headers[k]
		for _, v := range values {
			fmt.Fprintf(w, "%c %s: %s\n", prefix, k, v)
		}
	}
}