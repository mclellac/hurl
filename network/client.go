// network/client.go
package network

import (
	"fmt"
	"net/http"
	"time"
)

// FetchHeaders performs an HTTP GET request to the specified URL
// and returns the response. The caller is responsible for closing
// the response body.
func FetchHeaders(url string) (*http.Response, error) {
	client := &http.Client{
		Timeout: 30 * time.Second, // Example timeout
		// Prevent following redirects automatically if you *only* want the first response's headers
		// CheckRedirect: func(req *http.Request, via []*http.Request) error {
		//  return http.ErrUseLastResponse
		// },
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %s: %w", url, err)
	}

	// Set a default User-Agent, good practice
	req.Header.Set("User-Agent", "hurl/0.1 (Go-http-client)")
	// Add any other default headers if needed

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing GET request to %s: %w", url, err)
	}

	// Note: We DON'T close resp.Body here. The caller (main.go) will do that.
	return resp, nil
}
