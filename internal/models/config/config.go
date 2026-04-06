// internal/models/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Korrnals/gotr/internal/ui"
)

// DefaultConfigValues are default placeholders used in the configuration template.
// These values are used both when creating config and for validity checks.
const (
	DefaultBaseURL  = "https://yourcompany.testrail.io/"
	DefaultUsername = "your-email@example.com"
	DefaultAPIKey   = "your_api_key_here"
)

// ConfigData stores serialized gotr configuration fields.
type ConfigData struct {
	BaseURL  string `yaml:"base_url"`
	Username string `yaml:"username"`
	APIKey   string `yaml:"api_key"`
	Insecure bool   `yaml:"insecure"`
	JqFormat bool   `yaml:"jq_format"`
	Debug    bool   `yaml:"debug"`
}

// Config represents a single configuration file.
type Config struct {
	Path string      // Full path to the file
	Data *ConfigData // Configuration data
}

// New creates a Config instance at the given path.
func New(path string) *Config {
	return &Config{
		Path: path,
		Data: &ConfigData{},
	}
}

// Default returns a Config at the standard path (~/.gotr/config/default.yaml).
func Default() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ".gotr", "config", "default.yaml")
	return New(path), nil
}

// WithDefaults populates the config with default placeholder values.
func (c *Config) WithDefaults() *Config {
	c.Data = &ConfigData{
		BaseURL:  DefaultBaseURL,
		Username: DefaultUsername,
		APIKey:   DefaultAPIKey,
		Insecure: false,
		JqFormat: false,
		Debug:    false,
	}
	return c
}

// Create writes the configuration file to disk.
func (c *Config) Create() error {
	dir := filepath.Dir(c.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	content := []byte(c.renderTemplate())

	if err := os.WriteFile(c.Path, content, 0600); err != nil {
		return fmt.Errorf("failed to write file %s: %w", c.Path, err)
	}

	ui.Infof(os.Stdout, "Config file created: %s", c.Path)
	return nil
}

func (c *Config) renderTemplate() string {
	data := c.Data
	if data == nil {
		data = (&Config{}).WithDefaults().Data
	}

	return fmt.Sprintf(`# gotr configuration file
#
# Configuration source priority:
#   1) CLI flags
#   2) Environment variables (TESTRAIL_*)
#   3) This file

# TestRail base URL.
# Cloud example:  https://yourcompany.testrail.io
# Server example: https://testrail.example.local
base_url: %q

# Login (usually the TestRail user's email).
username: %q

# TestRail user API key.
api_key: %q

# true  -> skip TLS certificate verification (insecure, use only for internal environments)
# false -> standard secure TLS verification
insecure: %v

# Enable jq-formatted output (if the embedded jq binary is available).
jq_format: %v

# Enable gotr debug output.
debug: %v

compare:
  # Deployment mode for compare requests:
  #   auto   - attempt to detect from URL (cloud/server)
  #   cloud  - force cloud profile
  #   server - force server profile
  deployment: "auto"

  # For cloud profile: professional|enterprise
  cloud_tier: "professional"

  # Global rate limit (requests per minute) for compare.
  #   -1 -> automatic based on profile (cloud/server)
  #    0 -> rate limiting disabled
  #   >0 -> fixed value in req/min
  rate_limit: -1

  # Default for cloud when rate_limit=-1.
  # professional: 180, enterprise: 300
  cloud_rate_limit: 300

  # Default for server when rate_limit=-1.
  # Typically 0 (no limit).
  server_rate_limit: 0

  cases:
    # Parallelism across suites (between suites).
    parallel_suites: 10

    # Parallelism for pages within a single suite.
    parallel_pages: 6

    # Number of retries per page during the main compare cases fetch stage.
    page_retries: 5

    # Timeout for the entire compare cases operation.
    timeout: "30m"

    retry:
      # Retry attempts for a single failed page.
      attempts: 5

      # Number of parallel retry workers.
      workers: 12

      # Delay between retry attempts for a single page.
      delay: "200ms"

    # Always attempt to automatically retry failed pages after the main compare cases stage.
    auto_retry_failed_pages: true
`, data.BaseURL, data.Username, data.APIKey, data.Insecure, data.JqFormat, data.Debug)
}

// PathString returns the config file path.
func (c *Config) PathString() string {
	return c.Path
}

// IsValid checks that the config contains real data, not default placeholders.
func (c *Config) IsValid() bool {
	if c.Data == nil {
		return false
	}
	return c.Data.BaseURL != "" && c.Data.BaseURL != DefaultBaseURL &&
		c.Data.Username != "" && c.Data.Username != DefaultUsername &&
		c.Data.APIKey != "" && c.Data.APIKey != DefaultAPIKey
}

// IsDefaultValue checks whether the given value matches a default placeholder.
func IsDefaultValue(value, defaultValue string) bool {
	return value == "" || value == defaultValue
}
