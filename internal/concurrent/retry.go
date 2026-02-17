// Package concurrent provides utilities for concurrent API requests with rate limiting.
package concurrent

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/Korrnals/gotr/internal/log"
)

// RetryConfig configures retry behavior.
type RetryConfig struct {
	MaxRetries   int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	RetryableErrors []error // Specific errors that should trigger retry
}

// DefaultRetryConfig returns default retry configuration.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:   5,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// RetryableFunc is a function that can be retried.
type RetryableFunc func() error

// Retry executes a function with exponential backoff retry.
func Retry(config *RetryConfig, fn RetryableFunc) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Debug(fmt.Sprintf("Retry attempt %d/%d after %v", attempt, config.MaxRetries, delay))
			time.Sleep(delay)
			// Calculate next delay with exponential backoff
			delay = time.Duration(math.Min(
				float64(delay)*config.Multiplier,
				float64(config.MaxDelay),
			))
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err, config) {
			return fmt.Errorf("non-retryable error: %w", err)
		}

		log.Debug(fmt.Sprintf("Attempt %d failed: %v", attempt+1, err))
	}

	return fmt.Errorf("failed after %d attempts: %w", config.MaxRetries+1, lastErr)
}

// RetryWithContext executes a function with retry, respecting context cancellation.
func RetryWithContext(ctx context.Context, config *RetryConfig, fn RetryableFunc) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(delay):
				// Continue with retry
			}

			log.Debug(fmt.Sprintf("Retry attempt %d/%d after %v", attempt, config.MaxRetries, delay))
			// Calculate next delay with exponential backoff
			delay = time.Duration(math.Min(
				float64(delay)*config.Multiplier,
				float64(config.MaxDelay),
			))
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err, config) {
			return fmt.Errorf("non-retryable error: %w", err)
		}

		log.Debug(fmt.Sprintf("Attempt %d failed: %v", attempt+1, err))
	}

	return fmt.Errorf("failed after %d attempts: %w", config.MaxRetries+1, lastErr)
}

func isRetryableError(err error, config *RetryConfig) bool {
	if err == nil {
		return false
	}

	// Check specific retryable errors if configured
	if len(config.RetryableErrors) > 0 {
		for _, retryableErr := range config.RetryableErrors {
			if err == retryableErr {
				return true
			}
		}
		return false
	}

	// Default: all errors are retryable (for idempotent operations)
	return true
}

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	failures     int
	maxFailures  int
	timeout      time.Duration
	lastFailure  time.Time
	state        CircuitState
	mu           sync.Mutex
}

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	// CircuitClosed allows requests through.
	CircuitClosed CircuitState = iota
	// CircuitOpen blocks all requests.
	CircuitOpen
	// CircuitHalfOpen allows a test request through.
	CircuitHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
		state:       CircuitClosed,
	}
}

// Execute executes a function with circuit breaker protection.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if cb == nil {
		return fn()
	}

	if !cb.canExecute() {
		return fmt.Errorf("circuit breaker is open")
	}

	err := fn()
	cb.recordResult(err)
	return err
}

func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return true
	}
}

func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err == nil {
		// Success
		if cb.state == CircuitHalfOpen {
			cb.state = CircuitClosed
		}
		cb.failures = 0
	} else {
		// Failure
		cb.failures++
		cb.lastFailure = time.Now()

		if cb.failures >= cb.maxFailures {
			cb.state = CircuitOpen
		}
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() CircuitState {
	if cb == nil {
		return CircuitClosed
	}
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	if cb == nil {
		return
	}
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = CircuitClosed
	cb.failures = 0
}
