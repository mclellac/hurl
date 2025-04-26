// Package network handles making HTTP requests.
package network

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mclellac/hurl/config"
)

// akamaiPragmaValue is the static string used for the Akamai Pragma header.
const akamaiPragmaValue = "akamai-x-get-request-id,akamai-x-get-cache-key,akamai-x-cache-on,akamai-x-cache-remote-on,akamai-x-get-true-cache-key,akamai-x-check-cacheable,akamai-x-get-extracted-values,akamai-x-feo-trace,x-akamai-logging-mode: verbose"

// RequestOptions bundles parameters for making the HTTP request.
type RequestOptions struct {
	Method          string        // HTTP method (e.g., "GET", "POST")
	URL             string        // Target URL
	CustomHeaders   []string      // Custom headers in "Key: Value" format
	InsecureSkipTLS bool          // If true, skip TLS certificate verification
	FollowRedirects bool          // If true, follow HTTP 3xx redirects
	AddAkamaiPragma bool          // If true, add the Akamai debug Pragma header
	Verbose         bool          // If true, enable verbose output to stderr
	Config          config.Config // Color configuration
}

// Fetch performs an HTTP request based on the provided options.
// The caller is responsible for closing the response body if the returned response is non-nil.
func Fetch(opts RequestOptions) (*http.Response, error) {

	keyColor := config.GetAnsiCode(opts.Config.HeaderKeyColor)
	valueColor := config.GetAnsiCode(opts.Config.HeaderValueColor)
	traceColor := config.ColorWhite
	errorColor := config.ColorRed
	successColor := config.ColorGreen
	warningColor := config.ColorYellow
	resetColor := config.ColorReset

	tr := http.DefaultTransport.(*http.Transport).Clone()
	if tr.TLSClientConfig == nil {
		tr.TLSClientConfig = &tls.Config{}
	}
	tr.TLSClientConfig.InsecureSkipVerify = opts.InsecureSkipTLS

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}

	// This logic remains correct: if FollowRedirects is false (now the default unless -L is passed),
	// set CheckRedirect to prevent following. Otherwise, use default behavior.
	if !opts.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "%s* Ignoring redirect response from %s%s\n", traceColor, req.URL, resetColor)
			}
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
	}

	if opts.AddAkamaiPragma {
		req.Header.Set("Pragma", akamaiPragmaValue)
	}

	var trace *httptrace.ClientTrace
	currentReq := req
	if opts.Verbose {
		trace = &httptrace.ClientTrace{
			GetConn: func(hostPort string) {
				fmt.Fprintf(os.Stderr, "%s* Trying %s...%s\n", traceColor, hostPort, resetColor)
			},
			DNSStart: func(info httptrace.DNSStartInfo) {
				fmt.Fprintf(os.Stderr, "%s* Resolving %s...%s\n", traceColor, info.Host, resetColor)
			},
			DNSDone: func(info httptrace.DNSDoneInfo) {
				if info.Err != nil {
					fmt.Fprintf(os.Stderr, "%s* Error resolving host %s: %v%s\n", errorColor, currentReq.URL.Host, info.Err, resetColor)
					return
				}
				addrs := []string{}
				for _, ip := range info.Addrs {
					addrs = append(addrs, ip.String())
				}
				fmt.Fprintf(os.Stderr, "%s* Resolved %s to %s%v%s\n", traceColor, currentReq.URL.Host, valueColor, addrs, resetColor)
			},
			ConnectStart: func(network, addr string) {
				fmt.Fprintf(os.Stderr, "%s* Connecting to %s%s (%s)%s\n", traceColor, valueColor, addr, network, resetColor)
			},
			ConnectDone: func(network, addr string, err error) {
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s* Error connecting to %s: %v%s\n", errorColor, addr, err, resetColor)
				} else {
					fmt.Fprintf(os.Stderr, "%s* Connected to %s%s (%s)%s\n", traceColor, valueColor, addr, currentReq.URL.Host, resetColor)
				}
			},
			TLSHandshakeStart: func() {
				fmt.Fprintf(os.Stderr, "%s* Performing TLS handshake...%s\n", traceColor, resetColor)
			},
			TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s* TLS handshake error: %v%s\n", errorColor, err, resetColor)
					if cs.Version == 0 {
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
				fmt.Fprintf(os.Stderr, "%s* TLS handshake complete%s\n", traceColor, resetColor)
				fmt.Fprintf(os.Stderr, "%s* Protocol: %s%s%s\n", traceColor, valueColor, proto, resetColor)
				fmt.Fprintf(os.Stderr, "%s* Cipher Suite: %s%s%s\n", traceColor, valueColor, tls.CipherSuiteName(cs.CipherSuite), resetColor)
				if len(cs.PeerCertificates) > 0 {
					cert := cs.PeerCertificates[0]
					fmt.Fprintf(os.Stderr, "%s* Server certificate:%s\n", traceColor, resetColor)
					fmt.Fprintf(os.Stderr, "%s* Subject: %s%s%s\n", traceColor, valueColor, cert.Subject.String(), resetColor)
					fmt.Fprintf(os.Stderr, "%s* Issuer: %s%s%s\n", traceColor, valueColor, cert.Issuer.String(), resetColor)
					fmt.Fprintf(os.Stderr, "%s* Expiry: %s%s%s\n", traceColor, valueColor, cert.NotAfter.Format(time.RFC1123), resetColor)
				}
				if cs.NegotiatedProtocol != "" {
					fmt.Fprintf(os.Stderr, "%s* ALPN: server accepted %s%s%s\n", traceColor, valueColor, cs.NegotiatedProtocol, resetColor)
				}

			},
			GotConn: func(info httptrace.GotConnInfo) {
				fmt.Fprintf(os.Stderr, "%s* Connection established to %s%s%s\n", traceColor, valueColor, info.Conn.RemoteAddr(), resetColor)
			},
			GotFirstResponseByte: func() {
				fmt.Fprintf(os.Stderr, "%s* Receiving response headers...%s\n", traceColor, resetColor)
			},
		}
		traceCtx := httptrace.WithClientTrace(currentReq.Context(), trace)
		currentReq = currentReq.WithContext(traceCtx)
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "> ")
		fmt.Fprintf(os.Stderr, "%s%s%s ", keyColor, currentReq.Method, resetColor)
		fmt.Fprintf(os.Stderr, "%s%s%s ", valueColor, currentReq.URL.RequestURI(), resetColor)
		fmt.Fprintf(os.Stderr, "%s%s%s\n", valueColor, currentReq.Proto, resetColor)

		fmt.Fprintf(os.Stderr, "> ")
		fmt.Fprintf(os.Stderr, "%s%s%s: ", keyColor, "Host", resetColor)
		fmt.Fprintf(os.Stderr, "%s%s%s\n", valueColor, currentReq.Host, resetColor)

		printHeadersVerboseColor(os.Stderr, '>', currentReq.Header, opts.Config)
		fmt.Fprintf(os.Stderr, "> \n")
	}

	resp, err := client.Do(currentReq)

	if opts.Verbose && resp != nil {
		statusCodeColor := errorColor
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			statusCodeColor = successColor
		} else if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			statusCodeColor = warningColor
		}

		statusParts := strings.SplitN(resp.Status, " ", 2)
		statusCodeStr := statusParts[0]
		statusText := ""
		if len(statusParts) > 1 {
			statusText = statusParts[1]
		}

		fmt.Fprintf(os.Stderr, "< ")
		fmt.Fprintf(os.Stderr, "%s%s%s ", valueColor, resp.Proto, resetColor)
		fmt.Fprintf(os.Stderr, "%s%s%s ", statusCodeColor, statusCodeStr, resetColor)
		fmt.Fprintf(os.Stderr, "%s%s%s\n", valueColor, statusText, resetColor)

		printHeadersVerboseColor(os.Stderr, '<', resp.Header, opts.Config)
		fmt.Fprintf(os.Stderr, "< \n")
	}

	if err != nil {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "%s* Request failed: %v%s\n", errorColor, err, resetColor)
		}
		return resp, fmt.Errorf("error performing request: %w", err)
	}

	return resp, nil
}

// printHeadersVerboseColor prints headers to the specified writer with a prefix and colors.
func printHeadersVerboseColor(w io.Writer, prefix rune, headers http.Header, cfg config.Config) {
	keyColor := config.GetAnsiCode(cfg.HeaderKeyColor)
	valueColor := config.GetAnsiCode(cfg.HeaderValueColor)
	resetColor := config.ColorReset

	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		values := headers[k]
		for _, v := range values {
			fmt.Fprintf(w, "%c ", prefix) // Print prefix plainly
			fmt.Fprintf(w, "%s%s%s: ", keyColor, k, resetColor)
			fmt.Fprintf(w, "%s%s%s\n", valueColor, v, resetColor)
		}
	}
}