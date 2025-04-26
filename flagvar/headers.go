// Package flagvar provides custom types that satisfy the flag.Value interface.
package flagvar

import (
	"fmt"
)

// HeaderFlags implements pflag.Value interface for collecting multiple flag strings.
// It is exported because it starts with an uppercase letter.
type HeaderFlags []string

// String returns a string representation of the collected flags.
func (h *HeaderFlags) String() string {
	return fmt.Sprintf("%v", *h)
}

// Set appends a value to the collection. Called by flag.Parse() for each flag instance.
func (h *HeaderFlags) Set(value string) error {
	*h = append(*h, value)
	return nil
}

// Type returns the type description for pflag.
func (h *HeaderFlags) Type() string {
	return "stringSlice" // Added required Type() method
}

// Get returns the collected flag values as a slice of strings.
func (h *HeaderFlags) Get() []string {
	return *h
}