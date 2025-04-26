// config/colours.go
package config

import "strings"

// ANSI Color Codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

// DefaultColor is used if a configured color is invalid
const DefaultColor = ColorCyan

// colorMap maps color names (lowercase) to ANSI codes
var colorMap = map[string]string{
	"reset":  ColorReset,
	"red":    ColorRed,
	"green":  ColorGreen,
	"yellow": ColorYellow,
	"blue":   ColorBlue,
	"purple": ColorPurple,
	"cyan":   ColorCyan,
	"white":  ColorWhite,
}

// GetAnsiCode returns the ANSI code for a given color name.
// It defaults to DefaultColor if the name is not recognized.
func GetAnsiCode(name string) string {
	code, ok := colorMap[strings.ToLower(name)]
	if !ok {
		return DefaultColor // Return a default if color name is unknown
	}
	return code
}
