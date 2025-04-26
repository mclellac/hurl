// display/printer.go
package display

import (
	"fmt"
	"hurl/config" // Import the local config package
	"net/http"
	"sort"
	"strings"
)

// PrintHeaders takes HTTP headers and configuration, then prints them
// to standard output with configured colors.
func PrintHeaders(headers http.Header, cfg config.Config) {
	keyColor := config.GetAnsiCode(cfg.HeaderKeyColor)
	valueColor := config.GetAnsiCode(cfg.HeaderValueColor)
	resetColor := config.ColorReset

	// Sort header keys for consistent output order
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Print Status Line (optional but often useful, like curl -i)
	// If resp is available here, you could print resp.Status
	// fmt.Printf("%sHTTP/%d.%d %s%s\n", valueColor, resp.ProtoMajor, resp.ProtoMinor, resp.Status, resetColor)

	for _, k := range keys {
		values := headers[k]
		// Join multiple values for the same header key, separated by comma+space
		// Or print each on a new line if preferred
		valueStr := strings.Join(values, ", ")
		fmt.Printf("%s%s:%s %s%s%s\n",
			keyColor,   // Color for key
			k,          // Header key
			resetColor, // Reset color after key
			valueColor, // Color for value
			valueStr,   // Header value(s)
			resetColor, // Reset color at end of line
		)
	}
}
