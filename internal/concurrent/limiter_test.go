package concurrent

import (
	"context"
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

func TestRateLimiter_WaitCtx_Canceled(t *testing.T) {
	rl := NewRateLimiter(60) // 1 req/sec, burst 10

	// Consume initial burst so next wait needs to block.
	for i := 0; i < rl.burst; i++ {
		rl.Wait()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := rl.WaitCtx(ctx)
	assert.ErrorIs(t, err, context.Canceled)
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

func TestReservation_Delay_NilReservation(t *testing.T) {
	var rsv *Reservation
	assert.Equal(t, time.Duration(0), rsv.Delay())
}

func TestReservation_OK_NilReservation(t *testing.T) {
	var rsv *Reservation
	assert.False(t, rsv.OK())
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

func TestNewAdaptiveRateLimiter_DefaultUnlimitedConfig(t *testing.T) {
	for _, rpm := range []int{0, -10} {
		arl := NewAdaptiveRateLimiter(rpm)
		assert.NotNil(t, arl)
		assert.Nil(t, arl.baseLimiter)
		assert.Equal(t, 0.0, arl.targetRPS)
		assert.Equal(t, 0.0, arl.currentRPS)
		assert.Equal(t, 10, arl.sampleSize)
		assert.Len(t, arl.responseTimes, 0)
		assert.Equal(t, 10, cap(arl.responseTimes))
	}
}

func TestNewAdaptiveRateLimiter_DefaultConfiguredSampleForLimitedMode(t *testing.T) {
	arl := NewAdaptiveRateLimiter(120)
	assert.NotNil(t, arl)
	assert.NotNil(t, arl.baseLimiter)
	assert.Equal(t, 10, arl.sampleSize)
	assert.Len(t, arl.responseTimes, 0)
	assert.Equal(t, 10, cap(arl.responseTimes))
}

func TestAdaptiveRateLimiter_Wait(t *testing.T) {
	arl := NewAdaptiveRateLimiter(600)
	// Should not panic
	arl.Wait()
}

func TestAdaptiveRateLimiter_WaitCtx_Canceled(t *testing.T) {
	arl := NewAdaptiveRateLimiter(60) // 1 req/sec, burst 10

	// Consume initial burst so next wait needs to block.
	for i := 0; i < arl.baseLimiter.burst; i++ {
		arl.Wait()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := arl.WaitCtx(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestAdaptiveRateLimiter_WaitCtx_UnlimitedMode(t *testing.T) {
	arl := NewAdaptiveRateLimiter(0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := arl.WaitCtx(ctx)
	assert.NoError(t, err)
}

func TestAdaptiveRateLimiter_WaitCtx_NilReceiver(t *testing.T) {
	var arl *AdaptiveRateLimiter
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := arl.WaitCtx(ctx)
	assert.NoError(t, err)
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

func TestAdaptiveRateLimiter_RecordResponseTime_NilAndUnlimitedMode(t *testing.T) {
	var nilARL *AdaptiveRateLimiter
	// Nil receiver should be a no-op.
	nilARL.RecordResponseTime(100 * time.Millisecond)

	unlimited := NewAdaptiveRateLimiter(0)
	unlimited.RecordResponseTime(2 * time.Second)
	assert.Nil(t, unlimited.baseLimiter)
	assert.Len(t, unlimited.responseTimes, 0)
}

func TestAdaptiveRateLimiter_RecordResponseTime_TrimsBySampleSize(t *testing.T) {
	arl := NewAdaptiveRateLimiter(60)
	arl.sampleSize = 3

	arl.RecordResponseTime(1 * time.Second)
	arl.RecordResponseTime(2 * time.Second)
	arl.RecordResponseTime(3 * time.Second)
	arl.RecordResponseTime(4 * time.Second)
	arl.RecordResponseTime(5 * time.Second)

	assert.Len(t, arl.responseTimes, 3)
	assert.Equal(t, 3*time.Second, arl.responseTimes[0])
	assert.Equal(t, 4*time.Second, arl.responseTimes[1])
	assert.Equal(t, 5*time.Second, arl.responseTimes[2])
}

func TestAdaptiveRateLimiter_AdjustRate_Branches(t *testing.T) {
	t.Run("slow path is clamped by floor", func(t *testing.T) {
		arl := NewAdaptiveRateLimiter(600) // targetRPS = 10
		arl.currentRPS = 5.1

		arl.adjustRate(11 * time.Second)

		assert.InDelta(t, 5.0, arl.currentRPS, 0.0001)
	})

	t.Run("fast path is clamped by target", func(t *testing.T) {
		arl := NewAdaptiveRateLimiter(600) // targetRPS = 10
		arl.currentRPS = 9.8

		arl.adjustRate(500 * time.Millisecond)

		assert.InDelta(t, 10.0, arl.currentRPS, 0.0001)
	})

	t.Run("normal latency keeps rate unchanged", func(t *testing.T) {
		arl := NewAdaptiveRateLimiter(600) // targetRPS = 10
		arl.currentRPS = 7.3

		arl.adjustRate(2 * time.Second)

		assert.InDelta(t, 7.3, arl.currentRPS, 0.0001)
	})
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
func TestMax(t *testing.T) {
	assert.Equal(t, 2.0, max(1.0, 2.0))
	assert.Equal(t, 2.0, max(2.0, 1.0))
	assert.Equal(t, 3.0, max(3.0, 3.0))
}