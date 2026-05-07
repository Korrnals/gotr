package client

import (
	"os"
	"testing"
)

// TestMain disables retry-with-backoff in unit tests by default. The retry
// loop is exercised explicitly by tests that need it via WithRetryPolicy or
// by directly calling the retry helpers; defaulting it off keeps the rest of
// the suite fast (no jittered sleeps on intentional 5xx fixtures).
func TestMain(m *testing.M) {
	defaultOptions.retryPolicy = RetryPolicy{MaxAttempts: 1}
	os.Exit(m.Run())
}
