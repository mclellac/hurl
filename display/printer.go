// Package display handles printing output in specific formats.
package display

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/mclellac/hurl/config"
)

// PrintHeaders takes HTTP headers and configuration, then prints them
// to the specified writer with configured colors.
func PrintHeaders(w io.Writer, headers http.Header, cfg config.Config) {
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
		valueStr := strings.Join(values, ", ")
		fmt.Fprintf(w, "%s%s:%s %s%s%s\n",
			keyColor,
			k,
			resetColor,
			valueColor,
			valueStr,
			resetColor,
		)
	}
}