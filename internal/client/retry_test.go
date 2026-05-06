package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fastRetry returns a RetryPolicy with the requested attempt count and
// negligible delays so tests stay fast.
func fastRetry(maxAttempts int) RetryPolicy {
	return RetryPolicy{
		MaxAttempts: maxAttempts,
		BaseDelay:   1 * time.Millisecond,
		MaxDelay:    2 * time.Millisecond,
	}
}

func TestRetryPolicy_RetryOn5xxThenSuccess(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "u", "k", false, WithRetryPolicy(fastRetry(4)))
	require.NoError(t, err)

	resp, err := c.Get(context.Background(), "get_runs/1", nil)
	require.NoError(t, err)
	resp.Body.Close()
	assert.EqualValues(t, 3, calls.Load(), "expected 2 retries before success")
}

func TestRetryPolicy_GivesUpAfterMaxAttempts(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "u", "k", false, WithRetryPolicy(fastRetry(3)))
	require.NoError(t, err)

	_, err = c.Get(context.Background(), "get_runs/1", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "after 3 attempts")
	assert.EqualValues(t, 3, calls.Load())
}

func TestRetryPolicy_DoesNotRetryNon5xx(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad"}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "u", "k", false, WithRetryPolicy(fastRetry(5)))
	require.NoError(t, err)

	_, err = c.Get(context.Background(), "get_runs/1", nil)
	require.Error(t, err)
	assert.EqualValues(t, 1, calls.Load(), "4xx must not be retried")
}

func TestRetryPolicy_RetriesOn429(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "u", "k", false, WithRetryPolicy(fastRetry(3)))
	require.NoError(t, err)

	resp, err := c.Get(context.Background(), "get_runs/1", nil)
	require.NoError(t, err)
	resp.Body.Close()
	assert.EqualValues(t, 2, calls.Load())
}

func TestRetryPolicy_HonorsContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	// Slow backoff so context cancellation will trigger first.
	c, err := NewClient(srv.URL, "u", "k", false, WithRetryPolicy(RetryPolicy{
		MaxAttempts: 5,
		BaseDelay:   500 * time.Millisecond,
		MaxDelay:    500 * time.Millisecond,
	}))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = c.Get(ctx, "get_runs/1", nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled),
		"expected ctx error, got %v", err)
}

func TestIsRetryable(t *testing.T) {
	assert.False(t, isRetryableErr(nil))
	assert.False(t, isRetryableErr(context.Canceled))
	assert.True(t, isRetryableErr(context.DeadlineExceeded))
	assert.True(t, isRetryableErr(fmt.Errorf("wrap: %w", context.DeadlineExceeded)))

	assert.True(t, isRetryableStatus(500))
	assert.True(t, isRetryableStatus(502))
	assert.True(t, isRetryableStatus(503))
	assert.True(t, isRetryableStatus(504))
	assert.True(t, isRetryableStatus(429))
	assert.False(t, isRetryableStatus(200))
	assert.False(t, isRetryableStatus(400))
	assert.False(t, isRetryableStatus(404))
}

func TestNextBackoff_GrowsAndCaps(t *testing.T) {
	p := RetryPolicy{MaxAttempts: 10, BaseDelay: 10 * time.Millisecond, MaxDelay: 80 * time.Millisecond}
	for attempt := 0; attempt < 8; attempt++ {
		d := nextBackoff(attempt, p)
		assert.GreaterOrEqual(t, d, time.Duration(0))
		assert.LessOrEqual(t, d, p.MaxDelay)
	}
}
