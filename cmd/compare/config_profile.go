package compare

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type compareCasesRuntimeConfig struct {
	compareHeavyRuntimeConfig
	RetryAttempts        int
	RetryWorkers         int
	RetryDelay           time.Duration
	AutoRetryFailedPages bool
}

type compareHeavyRuntimeConfig struct {
	RateLimit            int
	ParallelSuites       int
	ParallelPages        int
	PageRetries          int
	Timeout              time.Duration
}

func ensureCompareConfigDefaults() {
	viper.SetDefault("compare.deployment", "auto")
	viper.SetDefault("compare.cloud_tier", "professional")
	viper.SetDefault("compare.rate_limit", -1) // -1 = auto: server→0 (unlimited), cloud→by plan tier

	viper.SetDefault("compare.cloud_rate_limit", 300) // enterprise tier
	viper.SetDefault("compare.server_rate_limit", 0)  // server = unlimited

	viper.SetDefault("compare.cases.parallel_suites", 10)
	viper.SetDefault("compare.cases.parallel_pages", 6)
	viper.SetDefault("compare.cases.page_retries", 5)
	viper.SetDefault("compare.cases.timeout", "30m")
	viper.SetDefault("compare.cases.retry.attempts", 5)
	viper.SetDefault("compare.cases.retry.workers", 12)
	viper.SetDefault("compare.cases.retry.delay", "200ms")
	viper.SetDefault("compare.cases.auto_retry_failed_pages", true)
}

func resolveCompareCasesRuntimeConfig(
	cmdFlags map[string]any,
	baseURL string,
) (compareCasesRuntimeConfig, error) {
	heavyCfg, err := resolveCompareHeavyRuntimeConfig(cmdFlags, baseURL)
	if err != nil {
		return compareCasesRuntimeConfig{}, err
	}

	ensureCompareConfigDefaults()

	retryAttempts := viper.GetInt("compare.cases.retry.attempts")
	if retryAttempts <= 0 {
		retryAttempts = 5
	}
	if isFlagProvided(cmdFlags, "retry_attempts") {
		retryAttempts = cmdFlags["retry_attempts"].(int)
	}

	retryWorkers := viper.GetInt("compare.cases.retry.workers")
	if retryWorkers <= 0 {
		retryWorkers = 12
	}
	if isFlagProvided(cmdFlags, "retry_workers") {
		retryWorkers = cmdFlags["retry_workers"].(int)
	}

	retryDelay := 500 * time.Millisecond
	retryDelayText := strings.TrimSpace(viper.GetString("compare.cases.retry.delay"))
	if retryDelayText != "" {
		parsed, err := time.ParseDuration(retryDelayText)
		if err != nil {
			return compareCasesRuntimeConfig{}, fmt.Errorf("invalid compare.cases.retry.delay in config: %w", err)
		}
		retryDelay = parsed
	}
	if isFlagProvided(cmdFlags, "retry_delay") {
		retryDelay = cmdFlags["retry_delay"].(time.Duration)
	}

	autoRetry := viper.GetBool("compare.cases.auto_retry_failed_pages")

	return compareCasesRuntimeConfig{
		compareHeavyRuntimeConfig: heavyCfg,
		RetryAttempts:             retryAttempts,
		RetryWorkers:              retryWorkers,
		RetryDelay:                retryDelay,
		AutoRetryFailedPages:      autoRetry,
	}, nil
}

func resolveCompareHeavyRuntimeConfig(
	cmdFlags map[string]any,
	baseURL string,
) (compareHeavyRuntimeConfig, error) {
	ensureCompareConfigDefaults()

	deployment := strings.ToLower(strings.TrimSpace(viper.GetString("compare.deployment")))
	if deployment == "" {
		deployment = "auto"
	}

	if deployment == "auto" {
		deployment = detectDeploymentByURL(baseURL)
	}

	cloudTier := strings.ToLower(strings.TrimSpace(viper.GetString("compare.cloud_tier")))
	if cloudTier == "" {
		cloudTier = "professional"
	}

	rateLimit := viper.GetInt("compare.rate_limit")
	if isFlagProvided(cmdFlags, "rate_limit") {
		rateLimit = cmdFlags["rate_limit"].(int)
	}

	if rateLimit < 0 {
		rateLimit = deriveRateLimitByProfile(deployment, cloudTier)
	}

	parallelSuites := viper.GetInt("compare.cases.parallel_suites")
	if parallelSuites <= 0 {
		parallelSuites = 12
	}
	if isFlagProvided(cmdFlags, "parallel_suites") {
		parallelSuites = cmdFlags["parallel_suites"].(int)
	}

	parallelPages := viper.GetInt("compare.cases.parallel_pages")
	if parallelPages <= 0 {
		parallelPages = 8
	}
	if isFlagProvided(cmdFlags, "parallel_pages") {
		parallelPages = cmdFlags["parallel_pages"].(int)
	}

	pageRetries := viper.GetInt("compare.cases.page_retries")
	if pageRetries <= 0 {
		pageRetries = 5
	}
	if isFlagProvided(cmdFlags, "page_retries") {
		pageRetries = cmdFlags["page_retries"].(int)
	}

	timeoutText := strings.TrimSpace(viper.GetString("compare.cases.timeout"))
	timeout := 30 * time.Minute
	if timeoutText != "" {
		parsed, err := time.ParseDuration(timeoutText)
		if err != nil {
			return compareHeavyRuntimeConfig{}, fmt.Errorf("invalid compare.cases.timeout in config: %w", err)
		}
		timeout = parsed
	}
	if isFlagProvided(cmdFlags, "timeout") {
		timeout = cmdFlags["timeout"].(time.Duration)
	}

	return compareHeavyRuntimeConfig{
		RateLimit:      rateLimit,
		ParallelSuites: parallelSuites,
		ParallelPages:  parallelPages,
		PageRetries:    pageRetries,
		Timeout:        timeout,
	}, nil
}

func detectDeploymentByURL(baseURL string) string {
	url := strings.ToLower(strings.TrimSpace(baseURL))
	if strings.Contains(url, ".testrail.io") {
		return "cloud"
	}
	return "server"
}

func deriveRateLimitByProfile(deployment, cloudTier string) int {
	if deployment == "server" {
		value := viper.GetInt("compare.server_rate_limit")
		if !viper.IsSet("compare.server_rate_limit") {
			return 0
		}
		return value
	}

	value := viper.GetInt("compare.cloud_rate_limit")
	if viper.IsSet("compare.cloud_rate_limit") {
		return value
	}

	if cloudTier == "enterprise" {
		return 300
	}
	return 180
}

func isFlagProvided(flags map[string]any, key string) bool {
	_, ok := flags[key]
	return ok
}
