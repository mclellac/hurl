package network

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

const akamaiPragmaValue = "akamai-x-get-request-id,akamai-x-get-cache-key,akamai-x-cache-on,akamai-x-cache-remote-on,akamai-x-get-true-cache-key,akamai-x-check-cacheable,akamai-x-get-extracted-values,akamai-x-feo-trace,x-akamai-logging-mode: verbose"

func FetchHeaders(url string, addAkamaiPragma bool) (*http.Response, error) {

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %s: %w", url, err)
	}

	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36"
	req.Header.Set("User-Agent", userAgent)

	if addAkamaiPragma {
		req.Header.Set("Pragma", akamaiPragmaValue)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing GET request to %s: %w", url, err)
	}

	// Note: Caller (main.go) is responsible for closing resp.Body
	return resp, nil
}