// Package concurrent provides utilities for concurrent API requests with rate limiting.
package concurrent

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter implements token bucket rate limiting.
type RateLimiter struct {
	limiter *rate.Limiter
	burst   int
}

// NewRateLimiter creates a new rate limiter.
// requestsPerMinute: maximum requests allowed per minute.
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 180 // Default: 180 req/min (TestRail max limit)
	}

	// Convert requests per minute to rate per second
	ratePerSecond := rate.Limit(float64(requestsPerMinute) / 60.0)
	
	// Burst size: allow 15% of the rate or minimum 10
	burst := requestsPerMinute * 15 / 100
	if burst < 10 {
		burst = 10
	}

	return &RateLimiter{
		limiter: rate.NewLimiter(ratePerSecond, burst),
		burst:   burst,
	}
}

// Wait blocks until a token is available.
func (rl *RateLimiter) Wait() {
	if rl != nil && rl.limiter != nil {
		rl.limiter.Wait(context.Background())
	}
}

// WaitWithTimeout blocks until a token is available or timeout is reached.
// Returns true if token was acquired, false if timeout.
func (rl *RateLimiter) WaitWithTimeout(timeout time.Duration) bool {
	if rl == nil || rl.limiter == nil {
		return true
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := rl.limiter.Wait(ctx)
	return err == nil
}

// Allow checks if a request is allowed without blocking.
func (rl *RateLimiter) Allow() bool {
	if rl == nil || rl.limiter == nil {
		return true
	}
	return rl.limiter.Allow()
}

// Tokens returns the current number of available tokens.
func (rl *RateLimiter) Tokens() int {
	if rl == nil || rl.limiter == nil {
		return 0
	}
	return int(rl.limiter.Tokens())
}

// Reservation represents a reservation of tokens.
type Reservation struct {
	limiter   *RateLimiter
	reserved  bool
	delay     time.Duration
}

// Reserve reserves a token for future use.
func (rl *RateLimiter) Reserve() *Reservation {
	if rl == nil || rl.limiter == nil {
		return &Reservation{reserved: true, delay: 0}
	}

	rsv := rl.limiter.Reserve()
	return &Reservation{
		limiter:  rl,
		reserved: rsv.OK(),
		delay:    rsv.Delay(),
	}
}

// Delay returns the duration to wait before the reservation can be used.
func (r *Reservation) Delay() time.Duration {
	if r == nil {
		return 0
	}
	return r.delay
}

// OK returns true if the reservation is valid.
func (r *Reservation) OK() bool {
	if r == nil {
		return false
	}
	return r.reserved
}

// Cancel cancels the reservation.
func (r *Reservation) Cancel() {
	if r != nil && r.limiter != nil && r.limiter.limiter != nil && r.reserved {
		r.limiter.limiter.Reserve().Cancel()
	}
}

// AdaptiveRateLimiter adjusts rate based on API responses.
type AdaptiveRateLimiter struct {
	baseLimiter *RateLimiter
	targetRPS   float64
	currentRPS  float64
	sampleSize  int
	responseTimes []time.Duration
	mu          sync.Mutex
}

// NewAdaptiveRateLimiter creates an adaptive rate limiter.
func NewAdaptiveRateLimiter(initialRequestsPerMinute int) *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		baseLimiter:   NewRateLimiter(initialRequestsPerMinute),
		targetRPS:     float64(initialRequestsPerMinute) / 60.0,
		currentRPS:    float64(initialRequestsPerMinute) / 60.0,
		sampleSize:    10,
		responseTimes: make([]time.Duration, 0, 10),
	}
}

// Wait blocks until a token is available.
func (arl *AdaptiveRateLimiter) Wait() {
	if arl != nil && arl.baseLimiter != nil {
		arl.baseLimiter.Wait()
	}
}

// RecordResponseTime records a response time for adaptive adjustment.
func (arl *AdaptiveRateLimiter) RecordResponseTime(d time.Duration) {
	if arl == nil {
		return
	}

	arl.mu.Lock()
	defer arl.mu.Unlock()

	arl.responseTimes = append(arl.responseTimes, d)
	if len(arl.responseTimes) > arl.sampleSize {
		arl.responseTimes = arl.responseTimes[1:]
	}

	// Adjust rate based on average response time
	if len(arl.responseTimes) >= 5 {
		avg := averageDuration(arl.responseTimes)
		arl.adjustRate(avg)
	}
}

func averageDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	return sum / time.Duration(len(durations))
}

func (arl *AdaptiveRateLimiter) adjustRate(avgResponseTime time.Duration) {
	// If average response time is high, reduce rate
	// If low, we could increase rate, but be conservative
	if avgResponseTime > 2*time.Second {
		// Slow down
		arl.currentRPS = arl.currentRPS * 0.9
	} else if avgResponseTime < 500*time.Millisecond {
		// Speed up slightly, but don't exceed target
		arl.currentRPS = min(arl.currentRPS*1.05, arl.targetRPS)
	}

	// Update the underlying limiter
	newRate := rate.Limit(arl.currentRPS)
	arl.baseLimiter.limiter.SetLimit(newRate)
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
