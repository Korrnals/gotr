package concurrent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(60) // 60 req/min = 1 req/sec
	assert.NotNil(t, rl)
	assert.NotNil(t, rl.limiter)
	assert.Greater(t, rl.burst, 0)
}

func TestNewRateLimiter_Default(t *testing.T) {
	rl := NewRateLimiter(0) // Should use default
	assert.NotNil(t, rl)
}

func TestRateLimiter_Wait(t *testing.T) {
	rl := NewRateLimiter(600) // 10 req/sec
	
	// Should not panic
	rl.Wait()
	rl.Wait()
	rl.Wait()
}

func TestRateLimiter_Wait_Nil(t *testing.T) {
	var rl *RateLimiter
	// Should not panic
	rl.Wait()
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(600) // 10 req/sec
	assert.True(t, rl.Allow())
}

func TestRateLimiter_Allow_Nil(t *testing.T) {
	var rl *RateLimiter
	assert.True(t, rl.Allow()) // Nil limiter always allows
}

func TestRateLimiter_Tokens(t *testing.T) {
	rl := NewRateLimiter(60)
	tokens := rl.Tokens()
	assert.GreaterOrEqual(t, tokens, 0)
}

func TestRateLimiter_Tokens_Nil(t *testing.T) {
	var rl *RateLimiter
	assert.Equal(t, 0, rl.Tokens())
}

func TestRateLimiter_WaitWithTimeout(t *testing.T) {
	rl := NewRateLimiter(60)
	
	// Should succeed within timeout
	result := rl.WaitWithTimeout(1 * time.Second)
	assert.True(t, result)
}

func TestRateLimiter_WaitWithTimeout_Nil(t *testing.T) {
	var rl *RateLimiter
	result := rl.WaitWithTimeout(1 * time.Second)
	assert.True(t, result) // Nil limiter always succeeds
}

func TestReservation(t *testing.T) {
	rl := NewRateLimiter(60)
	rsv := rl.Reserve()
	
	assert.NotNil(t, rsv)
	assert.True(t, rsv.OK())
	assert.GreaterOrEqual(t, rsv.Delay(), time.Duration(0))
}

func TestReservation_NilLimiter(t *testing.T) {
	var rl *RateLimiter
	rsv := rl.Reserve()
	
	assert.NotNil(t, rsv)
	assert.True(t, rsv.OK())
	assert.Equal(t, time.Duration(0), rsv.Delay())
}

func TestReservation_Cancel(t *testing.T) {
	rl := NewRateLimiter(60)
	rsv := rl.Reserve()
	
	// Should not panic
	rsv.Cancel()
}

func TestNewAdaptiveRateLimiter(t *testing.T) {
	arl := NewAdaptiveRateLimiter(60)
	assert.NotNil(t, arl)
	assert.NotNil(t, arl.baseLimiter)
	assert.Equal(t, 1.0, arl.targetRPS) // 60/60
}

func TestAdaptiveRateLimiter_Wait(t *testing.T) {
	arl := NewAdaptiveRateLimiter(600)
	// Should not panic
	arl.Wait()
}

func TestAdaptiveRateLimiter_RecordResponseTime(t *testing.T) {
	arl := NewAdaptiveRateLimiter(600)
	
	// Record some fast response times
	for i := 0; i < 5; i++ {
		arl.RecordResponseTime(100 * time.Millisecond)
	}
	
	// Record some slow response times
	for i := 0; i < 5; i++ {
		arl.RecordResponseTime(3 * time.Second)
	}
	
	// Should have adjusted the rate
	assert.NotZero(t, arl.currentRPS)
}

func TestAverageDuration(t *testing.T) {
	durations := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
	}
	
	avg := averageDuration(durations)
	assert.Equal(t, 200*time.Millisecond, avg)
}

func TestAverageDuration_Empty(t *testing.T) {
	avg := averageDuration([]time.Duration{})
	assert.Equal(t, time.Duration(0), avg)
}

func TestMin(t *testing.T) {
	assert.Equal(t, 1.0, min(1.0, 2.0))
	assert.Equal(t, 1.0, min(2.0, 1.0))
	assert.Equal(t, 1.0, min(1.0, 1.0))
}
