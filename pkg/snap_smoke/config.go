//go:build smoke

package snap_smoke

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Korrnals/gotr/internal/client"
)

// SmokeConfig holds all environment-based configuration for smoke tests.
type SmokeConfig struct {
	BaseURL   string
	Username  string
	APIKey    string
	ProjectID int64
	SuiteID   int64
	Insecure  bool
}

// LoadConfig reads smoke test configuration from environment variables.
// Returns an error if required variables are missing.
func LoadConfig() (*SmokeConfig, error) {
	cfg := &SmokeConfig{}

	cfg.BaseURL = os.Getenv("GOTR_SMOKE_URL")
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("GOTR_SMOKE_URL is required")
	}

	cfg.Username = os.Getenv("GOTR_SMOKE_USER")
	if cfg.Username == "" {
		return nil, fmt.Errorf("GOTR_SMOKE_USER is required")
	}

	cfg.APIKey = os.Getenv("GOTR_SMOKE_KEY")
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("GOTR_SMOKE_KEY is required")
	}

	pidStr := os.Getenv("GOTR_SMOKE_PROJECT")
	if pidStr == "" {
		return nil, fmt.Errorf("GOTR_SMOKE_PROJECT is required")
	}
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("GOTR_SMOKE_PROJECT: invalid integer: %w", err)
	}
	cfg.ProjectID = pid

	if sidStr := os.Getenv("GOTR_SMOKE_SUITE"); sidStr != "" {
		sid, err := strconv.ParseInt(sidStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("GOTR_SMOKE_SUITE: invalid integer: %w", err)
		}
		cfg.SuiteID = sid
	}

	insecure := os.Getenv("GOTR_SMOKE_INSECURE")
	cfg.Insecure = insecure == "true" || insecure == "1"

	return cfg, nil
}

// NewClient creates a real TestRail API client from smoke config.
func NewClient(cfg *SmokeConfig) (*client.HTTPClient, error) {
	opts := []client.ClientOption{}
	if cfg.Insecure {
		opts = append(opts, client.WithSkipTlsVerify(true))
	}
	return client.NewClient(cfg.BaseURL, cfg.Username, cfg.APIKey, false, opts...)
}
