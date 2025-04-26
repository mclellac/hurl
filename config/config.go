// config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config defines the structure for our configuration file.
type Config struct {
	HeaderKeyColor   string `json:"header_key_color"`
	HeaderValueColor string `json:"header_value_color"`
}

// DefaultConfig returns the default configuration settings.
func DefaultConfig() Config {
	return Config{
		HeaderKeyColor:   "yellow", // Default key color
		HeaderValueColor: "cyan",   // Default value color
	}
}

// LoadConfig loads configuration from a JSON file.
// If the file doesn't exist or is invalid, it returns default settings.
func LoadConfig() (Config, error) {
	cfg := DefaultConfig() // Start with defaults

	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback if user config dir is not available
		fmt.Fprintf(os.Stderr, "Warning: Could not find user config directory: %v. Using default colors.\n", err)
		return cfg, nil // Not a fatal error, just use defaults
	}

	configPath := filepath.Join(configDir, "hurl", "config.json")

	configFile, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, which is fine. Use defaults.
			// Optionally, you could create a default config file here.
			// fmt.Fprintf(os.Stderr, "Info: Config file not found at %s. Using default colors.\n", configPath)
			return cfg, nil
		}
		// Other error opening file
		fmt.Fprintf(os.Stderr, "Warning: Error opening config file %s: %v. Using default colors.\n", configPath, err)
		return cfg, nil // Use defaults on error
	}
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error decoding config file %s: %v. Using default colors.\n", configPath, err)
		return DefaultConfig(), nil // Reset to defaults on decode error
	}

	// Basic validation (ensure colors are not empty strings, etc.)
	if cfg.HeaderKeyColor == "" {
		cfg.HeaderKeyColor = DefaultConfig().HeaderKeyColor
	}
	if cfg.HeaderValueColor == "" {
		cfg.HeaderValueColor = DefaultConfig().HeaderValueColor
	}

	return cfg, nil
}

// EnsureConfigDir checks if the config directory exists and creates it if not.
// This can be called once at startup if you want to ensure the dir exists
// for users to place their config file.
func EnsureConfigDir() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("could not find user config directory: %w", err)
	}

	hurlConfigDir := filepath.Join(configDir, "hurl")
	if _, err := os.Stat(hurlConfigDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Info: Creating config directory: %s\n", hurlConfigDir)
		err = os.MkdirAll(hurlConfigDir, 0750) // Read/write/execute for user, read/execute for group
		if err != nil {
			return fmt.Errorf("could not create config directory %s: %w", hurlConfigDir, err)
		}
	} else if err != nil {
		return fmt.Errorf("could not check config directory %s: %w", hurlConfigDir, err)
	}
	return nil
}
