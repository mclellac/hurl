# hurl - HTTP URL Inspector

`hurl` is a simple cURL clone, written in Go to support my own needs - particularly Akamai pragma header fetching.

## Installation

Requires Go (tested with version 1.18+).

```bash
go install [github.com/mclellac/hurl@latest](https://github.com/mclellac/hurl@latest)
```

Ensure your Go binary path (usually $GOPATH/bin or $HOME/go/bin) is included in your system's PATH environment variable.
Usage

hurl [flags] <URL>

By default, hurl performs a GET request to the specified <URL> and displays only the colored HTTP response headers to standard output. It does not follow redirects by default.
Options

The command accepts the following flags:

    --akamai-pragma: Send Akamai Pragma debug headers with the request.
    -H, --header value: Add a custom header to the request (e.g., -H "Accept: application/json"). This flag can be specified multiple times.
    -I, --head: Perform an HTTP HEAD request instead of GET. This overrides the -X flag if both are used.
    -k, --insecure: Allow connections to SSL sites without verifying the server certificate.
    -L, --location: Follow HTTP redirects (responses with 3xx status codes). Default behavior is not to follow redirects.
    -X, --request string: Specify the request method to use (e.g., POST, PUT, DELETE). (default: "GET")
    -v, --verbose: Enable verbose output. This prints detailed connection information, TLS handshake details, request headers (>), and response headers (<) to standard error (stderr) with coloring. Standard header output to stdout is suppressed in verbose mode.
    --help: Display this help message.

Configuration

The colors used for displaying the default response headers (key vs. value) can be configured via a JSON file. hurl looks for this file at:

    Linux/macOS: `~/.config/hurl/config.json`
    Windows: `%APPDATA%\hurl\config.json` (usually C:\Users\<YourUser>\AppData\Roaming\hurl\config.json)

The directory structure (hurl/) will be created if it doesn't exist on first run (or if config loading fails).

Example config.json:
JSON

{
  "header_key_color": "yellow",
  "header_value_color": "cyan"
}

Supported color names: red, green, yellow, blue, purple, cyan, white. If the file doesn't exist or a color name is invalid, default colors (yellow key, cyan value) are used.
Examples

1. Get default headers (colored):
```bash
hurl https://www.example.com
```

2. Verbose output (connection details, req/resp headers to stderr):

```bash
hurl -v https://example.com
```

3. Send a HEAD request:
```bash
hurl -I https://example.com
# Equivalent to: hurl -X HEAD https://example.com
```

4. Follow redirects:

```bash
# See initial 302 response headers (default behavior)
hurl [http://httpbin.org/redirect/1](http://httpbin.org/redirect/1)

# Follow redirect and see final 200 response headers
hurl -L [http://httpbin.org/redirect/1](http://httpbin.org/redirect/1)
```

5. Add custom headers:
```bash
hurl -H "Accept: application/json" -H "X-Custom: my-value" https://example.com
```

6. Allow self-signed certificate:
```bash
hurl -k https://self-signed.badssl.com/
```

7. Use Akamai debug headers:
```bash
hurl --akamai-pragma https://www.example.com
```
