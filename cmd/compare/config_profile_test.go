package compare

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestResolveCompareCasesRuntimeConfig_ConfigOnly(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	viper.Set("compare.deployment", "cloud")
	viper.Set("compare.cloud_tier", "enterprise")
	viper.Set("compare.rate_limit", -1)
	viper.Set("compare.cases.parallel_suites", 12)
	viper.Set("compare.cases.parallel_pages", 14)
	viper.Set("compare.cases.page_retries", 7)
	viper.Set("compare.cases.timeout", "45m")
	viper.Set("compare.cases.retry.attempts", 5)
	viper.Set("compare.cases.retry.workers", 9)
	viper.Set("compare.cases.retry.delay", "700ms")
	viper.Set("compare.cases.auto_retry_failed_pages", false)

	cfg, err := resolveCompareCasesRuntimeConfig(nil, "https://team.testrail.io")
	assert.NoError(t, err)
	assert.Equal(t, 300, cfg.RateLimit)
	assert.Equal(t, 12, cfg.ParallelSuites)
	assert.Equal(t, 14, cfg.ParallelPages)
	assert.Equal(t, 7, cfg.PageRetries)
	assert.Equal(t, 45*time.Minute, cfg.Timeout)
	assert.Equal(t, 5, cfg.RetryAttempts)
	assert.Equal(t, 9, cfg.RetryWorkers)
	assert.Equal(t, 700*time.Millisecond, cfg.RetryDelay)
	assert.False(t, cfg.AutoRetryFailedPages)
}

func TestResolveCompareCasesRuntimeConfig_FlagOverrides(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	viper.Set("compare.deployment", "cloud")
	viper.Set("compare.cloud_tier", "professional")
	viper.Set("compare.rate_limit", -1)

	overrides := map[string]any{
		"rate_limit":      240,
		"parallel_suites": 16,
		"parallel_pages":  20,
		"page_retries":    11,
		"timeout":         20 * time.Minute,
		"retry_attempts":  6,
		"retry_workers":   10,
		"retry_delay":     2 * time.Second,
	}

	cfg, err := resolveCompareCasesRuntimeConfig(overrides, "https://team.testrail.io")
	assert.NoError(t, err)
	assert.Equal(t, 240, cfg.RateLimit)
	assert.Equal(t, 16, cfg.ParallelSuites)
	assert.Equal(t, 20, cfg.ParallelPages)
	assert.Equal(t, 11, cfg.PageRetries)
	assert.Equal(t, 20*time.Minute, cfg.Timeout)
	assert.Equal(t, 6, cfg.RetryAttempts)
	assert.Equal(t, 10, cfg.RetryWorkers)
	assert.Equal(t, 2*time.Second, cfg.RetryDelay)
}

func TestResolveCompareCasesRuntimeConfig_DefaultsAndAutoDeployment(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	cfg, err := resolveCompareCasesRuntimeConfig(nil, "https://mycompany.testrail.io")
	assert.NoError(t, err)
	assert.Equal(t, 300, cfg.RateLimit)
	assert.Equal(t, 10, cfg.ParallelSuites)
	assert.Equal(t, 6, cfg.ParallelPages)
	assert.Equal(t, 5, cfg.PageRetries)
	assert.Equal(t, 30*time.Minute, cfg.Timeout)
	assert.True(t, cfg.AutoRetryFailedPages)

	cfgServer, err := resolveCompareCasesRuntimeConfig(nil, "https://testrail.internal.local")
	assert.NoError(t, err)
	assert.Equal(t, 0, cfgServer.RateLimit)
}

func TestResolveCompareCasesRuntimeConfig_InvalidDurations(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	viper.Set("compare.cases.timeout", "not-a-duration")
	_, err := resolveCompareCasesRuntimeConfig(nil, "https://team.testrail.io")
	assert.Error(t, err)

	viper.Reset()
	viper.Set("compare.cases.retry.delay", "bad")
	_, err = resolveCompareCasesRuntimeConfig(nil, "https://team.testrail.io")
	assert.Error(t, err)
}


