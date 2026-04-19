// internal/models/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultConfigValues are default placeholders used in the configuration template.
// These values are used both when creating config and for validity checks.
const (
	DefaultBaseURL  = "https://yourcompany.testrail.io/"
	DefaultUsername = "your-email@example.com"
	DefaultAPIKey   = "your_api_key_here"
)

// SnapshotRetention stores retention policy for snapshots
type SnapshotRetention struct {
	Enabled           bool     `yaml:"enabled"`
	DefaultTTLDays    int      `yaml:"default_ttl_days"`
	ProtectedPrefixes []string `yaml:"protected_prefixes"`
	FrozenSnapshots   []string `yaml:"frozen_snapshots"`
}

// ValidateTTL checks that DefaultTTLDays is positive
func (sr *SnapshotRetention) ValidateTTL() error {
	if sr.DefaultTTLDays <= 0 {
		return fmt.Errorf("snapshot retention: default_ttl_days must be positive, got %d", sr.DefaultTTLDays)
	}
	return nil
}

// ValidateProtectedPrefixes checks for duplicate or conflicting prefixes
func (sr *SnapshotRetention) ValidateProtectedPrefixes() error {
	seen := make(map[string]bool)
	for _, prefix := range sr.ProtectedPrefixes {
		if seen[prefix] {
			return fmt.Errorf("snapshot retention: duplicate protected_prefix: %q", prefix)
		}
		seen[prefix] = true
	}
	return nil
}

// Validate performs all validation checks on retention config
func (sr *SnapshotRetention) Validate() error {
	if sr == nil {
		return nil
	}
	if sr.Enabled {
		if err := sr.ValidateTTL(); err != nil {
			return err
		}
		if err := sr.ValidateProtectedPrefixes(); err != nil {
			return err
		}
	}
	return nil
}

// SnapshotConfig stores snapshot-related configuration
type SnapshotConfig struct {
	Enabled        bool                `yaml:"enabled"`
	MaxSnapshots   int                 `yaml:"max_snapshots"`
	Retention      *SnapshotRetention  `yaml:"retention"`
	Attachments    AttachmentConfig    `yaml:"attachments"`
}

// ValidateMaxSnapshots checks that MaxSnapshots is positive
func (sc *SnapshotConfig) ValidateMaxSnapshots() error {
	if sc.MaxSnapshots <= 0 {
		return fmt.Errorf("snapshot config: max_snapshots must be positive, got %d", sc.MaxSnapshots)
	}
	return nil
}

// Validate performs all validation checks on snapshot config
func (sc *SnapshotConfig) Validate() error {
	if sc == nil || !sc.Enabled {
		return nil
	}
	if err := sc.ValidateMaxSnapshots(); err != nil {
		return err
	}
	if err := sc.Retention.Validate(); err != nil {
		return err
	}
	return nil
}

// AttachmentConfig stores attachment handling configuration
type AttachmentConfig struct {
	SaveBinary            string `yaml:"save_binary"`
	MaxFileMB             int    `yaml:"max_file_mb"`
	Compress              bool   `yaml:"compress"`
	PromptAboveThreshold  bool   `yaml:"prompt_above_threshold"`
}

// ConfigData stores serialized gotr configuration fields.
type ConfigData struct {
	BaseURL  string            `yaml:"base_url"`
	Username string            `yaml:"username"`
	APIKey   string            `yaml:"api_key"`
	Insecure bool              `yaml:"insecure"`
	JqFormat bool              `yaml:"jq_format"`
	Debug    bool              `yaml:"debug"`
	Snap     *SnapshotConfig   `yaml:"snap"`
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
		Snap: &SnapshotConfig{
			Enabled:      true,
			MaxSnapshots: 100,
			Retention: &SnapshotRetention{
				Enabled:           true,
				DefaultTTLDays:    30,
				ProtectedPrefixes: []string{"pinned_", "archived_"},
				FrozenSnapshots:   []string{},
			},
			Attachments: AttachmentConfig{
				SaveBinary:           "auto",
				MaxFileMB:            10,
				Compress:             true,
				PromptAboveThreshold: true,
			},
		},
	}
	return c
}

// Create writes the configuration file to disk.
func (c *Config) Create() error {
	dir := filepath.Dir(c.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	content := []byte(c.renderTemplate())

	if err := os.WriteFile(c.Path, content, 0o600); err != nil {
		return fmt.Errorf("failed to write file %s: %w", c.Path, err)
	}

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

# Snapshot before mutations (enabled by default).
# Saves entity state to ~/.gotr/snaps/ for rollback.
snap:
  enabled: true
  max_snapshots: 100

  # Snapshot retention policy (auto-cleanup of old snapshots)
  retention:
    enabled: true
    # Default time-to-live for snapshots (days)
    # Snapshots older than this will be auto-deleted (except protected)
    default_ttl_days: 30
    # Prefixes for snapshots that should never be auto-deleted
    # Example: pinned_*, archived_*
    protected_prefixes:
      - "pinned_"
      - "archived_"
    # List of snapshot IDs to permanently freeze (never delete)
    # Example: ["snap-critical-migration-2026-04", "snap-backup-v1"]
    frozen_snapshots: []

  # Attachment handling when saving snapshots
  attachments:
    # Whether to save binary attachments: auto|always|never
    # auto: only if under max_file_mb
    save_binary: "auto"
    max_file_mb: 10
    compress: true
    prompt_above_threshold: true
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
