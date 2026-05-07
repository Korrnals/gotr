package client

import (
	"context"
	"errors"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/Korrnals/gotr/internal/debug"
)

// RetryPolicy controls automatic retries for transient failures on idempotent
// (GET) HTTP calls.
//
// A request is retried when the underlying error indicates a network timeout
// or a transient server condition (HTTP 5xx, 429). Backoff is exponential
// (BaseDelay * 2^attempt) with full jitter, capped at MaxDelay.
//
// MaxAttempts counts the initial request as attempt #1; setting it to 1
// effectively disables retries.
type RetryPolicy struct {
	MaxAttempts int           // total attempts including the first; <=1 disables retry
	BaseDelay   time.Duration // initial backoff delay
	MaxDelay    time.Duration // upper bound on a single sleep
}

// DefaultRetryPolicy returns sensible defaults for TestRail GET endpoints.
//
// 4 attempts × exponential backoff (1s, 2s, 4s) with full jitter — total
// worst-case extra wait ≈ 7s before the final attempt. Tuned for transient
// gateway timeouts on heavy legacy endpoints (e.g. get_results_for_run for
// huge runs).
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 4,
		BaseDelay:   1 * time.Second,
		MaxDelay:    16 * time.Second,
	}
}

// isRetryableErr reports whether a transport-level error indicates a
// transient condition worth retrying. Caller-canceled contexts are NOT
// retried.
func isRetryableErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	// Per-request deadline exceeded (e.g. http.Client.Timeout).
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var nerr net.Error
	if errors.As(err, &nerr) && nerr.Timeout() {
		return true
	}
	var uerr *url.Error
	if errors.As(err, &uerr) {
		if uerr.Timeout() {
			return true
		}
		// Connection reset / EOF mid-stream from flaky upstream.
		return true
	}
	return false
}

// isRetryableStatus reports whether an HTTP status indicates a transient
// server-side condition.
func isRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests ||
		status == http.StatusBadGateway ||
		status == http.StatusServiceUnavailable ||
		status == http.StatusGatewayTimeout ||
		(status >= 500 && status <= 599)
}

// nextBackoff returns the jittered sleep duration for the given attempt
// (0-indexed). Uses full jitter — the actual sleep is in [0, capped).
func nextBackoff(attempt int, p RetryPolicy) time.Duration {
	// Exponential, capped.
	d := p.BaseDelay
	for i := 0; i < attempt && d < p.MaxDelay; i++ {
		d *= 2
	}
	if d > p.MaxDelay {
		d = p.MaxDelay
	}
	if d <= 0 {
		return 0
	}
	// Full jitter: uniform in [0, d).
	return time.Duration(rand.Int64N(int64(d)))
}

// sleepWithCtx waits for the given duration or returns ctx.Err() if the
// context is canceled first.
func sleepWithCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// logRetry emits a debug line describing why a retry is happening.
func logRetry(endpoint string, attempt, maxAttempts int, delay time.Duration, reason string) {
	debug.DebugPrint("{retry} %s attempt %d/%d in %s: %s",
		endpoint, attempt+1, maxAttempts, delay.Round(time.Millisecond), reason)
}
