package concurrent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()
	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.InitialDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
}

func TestRetry_Success(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	err := Retry(nil, fn)
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestRetry_SuccessAfterFailures(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	err := Retry(config, fn)
	assert.NoError(t, err)
	assert.Equal(t, 3, callCount)
}

func TestRetry_MaxRetriesExceeded(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:   2,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	fn := func() error {
		return errors.New("permanent error")
	}

	err := Retry(config, fn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permanent error")
	assert.Contains(t, err.Error(), "failed after 3 attempts")
}

func TestRetry_NonRetryableError(t *testing.T) {
	retryableErr := errors.New("retryable")
	nonRetryableErr := errors.New("non-retryable")

	config := &RetryConfig{
		MaxRetries:      3,
		InitialDelay:    10 * time.Millisecond,
		RetryableErrors: []error{retryableErr},
	}

	fn := func() error {
		return nonRetryableErr
	}

	err := Retry(config, fn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-retryable error")
}

func TestRetryWithContext_Success(t *testing.T) {
	ctx := context.Background()
	config := &RetryConfig{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
	}

	fn := func() error {
		return nil
	}

	err := RetryWithContext(ctx, config, fn)
	assert.NoError(t, err)
}

func TestRetryWithContext_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := &RetryConfig{
		MaxRetries:   5,
		InitialDelay: 1 * time.Second,
	}

	callCount := 0
	fn := func() error {
		callCount++
		if callCount == 1 {
			cancel() // Cancel context after first attempt
		}
		return errors.New("error")
	}

	err := RetryWithContext(ctx, config, fn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
}

func TestIsRetryableError(t *testing.T) {
	// No specific retryable errors configured - all are retryable
	config := &RetryConfig{}
	assert.True(t, isRetryableError(errors.New("any error"), config))
	assert.False(t, isRetryableError(nil, config))

	// Specific retryable errors configured
	specificErr := errors.New("specific")
	config.RetryableErrors = []error{specificErr}
	assert.True(t, isRetryableError(specificErr, config))
	assert.False(t, isRetryableError(errors.New("other"), config))
}

func TestNewCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(5, 30*time.Second)
	assert.NotNil(t, cb)
	assert.Equal(t, 5, cb.maxFailures)
	assert.Equal(t, 30*time.Second, cb.timeout)
	assert.Equal(t, CircuitClosed, cb.state)
}

func TestCircuitBreaker_Execute_Success(t *testing.T) {
	cb := NewCircuitBreaker(3, 30*time.Second)
	
	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	err := cb.Execute(fn)
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
	assert.Equal(t, CircuitClosed, cb.State())
}

func TestCircuitBreaker_Execute_Failure(t *testing.T) {
	cb := NewCircuitBreaker(3, 30*time.Second)
	
	err := cb.Execute(func() error {
		return errors.New("error")
	})
	
	assert.Error(t, err)
	assert.Equal(t, 1, cb.failures)
	assert.Equal(t, CircuitClosed, cb.State()) // Still closed, not enough failures
}

func TestCircuitBreaker_OpensAfterFailures(t *testing.T) {
	cb := NewCircuitBreaker(3, 30*time.Second)
	
	// Cause 3 failures
	for i := 0; i < 3; i++ {
		cb.Execute(func() error {
			return errors.New("error")
		})
	}
	
	assert.Equal(t, CircuitOpen, cb.State())
	
	// Next execution should be blocked
	err := cb.Execute(func() error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")
}

func TestCircuitBreaker_HalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(1, 50*time.Millisecond)
	
	// Cause failure to open circuit
	cb.Execute(func() error {
		return errors.New("error")
	})
	
	assert.Equal(t, CircuitOpen, cb.State())
	
	// Wait for timeout
	time.Sleep(100 * time.Millisecond)
	
	// Execute should transition to half-open and allow execution
	err := cb.Execute(func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, CircuitClosed, cb.State())
}

func TestCircuitBreaker_State_Nil(t *testing.T) {
	var cb *CircuitBreaker
	assert.Equal(t, CircuitClosed, cb.State())
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(1, 30*time.Second)
	
	// Open circuit
	cb.Execute(func() error {
		return errors.New("error")
	})
	assert.Equal(t, CircuitOpen, cb.State())
	
	// Reset
	cb.Reset()
	assert.Equal(t, CircuitClosed, cb.State())
	assert.Equal(t, 0, cb.failures)
}

func TestCircuitBreaker_Reset_Nil(t *testing.T) {
	var cb *CircuitBreaker
	// Should not panic
	cb.Reset()
}

func TestCircuitBreaker_Execute_Nil(t *testing.T) {
	var cb *CircuitBreaker
	
	executed := false
	err := cb.Execute(func() error {
		executed = true
		return nil
	})
	
	assert.NoError(t, err)
	assert.True(t, executed)
}
