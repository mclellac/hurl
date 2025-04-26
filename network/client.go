package network

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

func FetchHeaders(url string) (*http.Response, error) {
	// Ensure you are using a transport that allows HTTP/2 (like default or a clone)
	// If you previously forced HTTP/1.1, revert that change.
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{} // Basic TLS config
	// Ensure these lines that disable HTTP/2 are commented out or removed:
	// tr.TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)
	// tr.ForceAttemptHTTP2 = false

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %s: %w", url, err)
	}

	// Set the User-Agent to mimic current Chrome on Windows
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36"
	req.Header.Set("User-Agent", userAgent) // <--- SET CHROME USER AGENT HERE

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing GET request to %s: %w", url, err)
	}
	defer resp.Body.Close() // Ensure body is closed here or in caller (main.go)

	return resp, nil
}
