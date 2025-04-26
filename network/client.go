package network

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// akamaiPragmaValue is the static string used for the Akamai Pragma header.
const akamaiPragmaValue = "akamai-x-get-request-id,akamai-x-get-cache-key,akamai-x-cache-on,akamai-x-cache-remote-on,akamai-x-get-true-cache-key,akamai-x-check-cacheable,akamai-x-get-extracted-values,akamai-x-feo-trace,x-akamai-logging-mode: verbose"

// RequestOptions bundles parameters for making the HTTP request.
type RequestOptions struct {
	Method          string   // HTTP method (e.g., "GET", "POST")
	URL             string   // Target URL
	CustomHeaders   []string // Custom headers in "Key: Value" format
	InsecureSkipTLS bool     // If true, skip TLS certificate verification
	FollowRedirects bool     // If true, follow HTTP 3xx redirects
	AddAkamaiPragma bool     // If true, add the Akamai debug Pragma header
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
			return http.ErrUseLastResponse
		}
	}

	req, err := http.NewRequest(opts.Method, opts.URL, nil)
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
		// Silently ignore invalid header formats for now
	}

	if opts.AddAkamaiPragma {
		req.Header.Set("Pragma", akamaiPragmaValue)
	}

	resp, err := client.Do(req)
	if err != nil {
		// Note: Caller (main.go) is responsible for closing resp.Body if resp is non-nil
		return resp, fmt.Errorf("error performing request: %w", err) // Return potentially non-nil resp on error
	}

	return resp, nil
}