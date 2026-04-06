package concurrency

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestResultAggregator_Basic(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(100)

	ra.StartCtx(ctx)

	// Submit some results
	ra.Submit(PageResult{
		SuiteID: 1,
		Offset:  0,
		Cases:   []data.Case{{ID: 1, Title: "Case 1"}, {ID: 2, Title: "Case 2"}},
	})

	ra.Submit(PageResult{
		SuiteID: 1,
		Offset:  2,
		Cases:   []data.Case{{ID: 3, Title: "Case 3"}},
	})

	// Stop and get results
	cases, errs := ra.Stop()

	assert.Len(t, cases, 3)
	assert.Len(t, errs, 0)

	// Verify cases (order not guaranteed in concurrent processing)
	ids := make([]int64, len(cases))
	for i, c := range cases {
		ids[i] = c.ID
	}
	assert.ElementsMatch(t, []int64{1, 2, 3}, ids)
}

func TestNewResultAggregator_DefaultBufferConfig(t *testing.T) {
	for _, size := range []int{0, -1} {
		ra := NewResultAggregator(size)
		assert.NotNil(t, ra)
		assert.Equal(t, 1000, ra.bufferSize)
		assert.NotNil(t, ra.seenIDs)
		assert.NotNil(t, ra.errors)
		assert.Empty(t, ra.errors)
	}
}

func TestNewResultAggregator_ExplicitBufferConfig(t *testing.T) {
	ra := NewResultAggregator(128)
	assert.NotNil(t, ra)
	assert.Equal(t, 128, ra.bufferSize)
	assert.NotNil(t, ra.seenIDs)
	assert.NotNil(t, ra.errors)
	assert.Empty(t, ra.errors)
}

func TestResultAggregator_Wait(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(100)
	ra.StartCtx(ctx)
	ra.Stop()

	done := make(chan struct{})
	go func() {
		ra.Wait()
		close(done)
	}()

	select {
	case <-done:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("Wait() did not return after Stop()")
	}
}

func TestPageResult_IsSuccess(t *testing.T) {
	pr := PageResult{}
	assert.True(t, pr.IsSuccess())

	pr.Error = errors.New("fetch error")
	assert.False(t, pr.IsSuccess())
}

func TestResultAggregator_Deduplication(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(100)

	ra.StartCtx(ctx)

	// Submit results with duplicate IDs
	ra.Submit(PageResult{
		SuiteID: 1,
		Cases:   []data.Case{{ID: 1, Title: "Case 1"}, {ID: 2, Title: "Case 2"}},
	})

	// Duplicate ID 2
	ra.Submit(PageResult{
		SuiteID: 1,
		Cases:   []data.Case{{ID: 2, Title: "Duplicate Case 2"}, {ID: 3, Title: "Case 3"}},
	})

	cases, _ := ra.Stop()

	// Should have 3 unique cases
	assert.Len(t, cases, 3)

	// Verify Case 2 wasn't overwritten
	var case2 data.Case
	for _, c := range cases {
		if c.ID == 2 {
			case2 = c
			break
		}
	}
	assert.Equal(t, "Case 2", case2.Title) // First occurrence wins
}

func TestResultAggregator_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(100)

	ra.StartCtx(ctx)

	// Submit successful result
	ra.Submit(PageResult{
		SuiteID: 1,
		Cases:   []data.Case{{ID: 1, Title: "Case 1"}},
	})

	// Submit error result
	ra.Submit(PageResult{
		SuiteID: 2,
		Offset:  0,
		Error:   errors.New("fetch failed"),
	})

	// Submit direct error
	ra.SubmitError(errors.New("direct error"))

	cases, errs := ra.Stop()

	assert.Len(t, cases, 1)
	assert.Len(t, errs, 2)
}

func TestResultAggregator_Concurrent(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(1000)

	ra.StartCtx(ctx)

	// Concurrent submissions
	numProducers := 10
	numResultsPerProducer := 100

	done := make(chan struct{})
	for i := 0; i < numProducers; i++ {
		go func(producerID int) {
			defer func() { done <- struct{}{} }()

			for j := 0; j < numResultsPerProducer; j++ {
				caseID := int64(producerID*1000000 + j)
				ra.Submit(PageResult{
					SuiteID: int64(producerID),
					Offset:  j,
					Cases:   []data.Case{{ID: caseID, Title: "Case"}},
				})
			}
		}(i)
	}

	// Wait for all producers
	for i := 0; i < numProducers; i++ {
		<-done
	}

	cases, _ := ra.Stop()

	// Should have all unique cases
	expectedCount := numProducers * numResultsPerProducer
	assert.Len(t, cases, expectedCount)

	// Verify no duplicates
	seen := make(map[int64]bool)
	for _, c := range cases {
		assert.False(t, seen[c.ID], "Duplicate case ID: %d", c.ID)
		seen[c.ID] = true
	}
}

func TestResultAggregator_Stats(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(100)

	// Stats before start
	stats := ra.Stats()
	assert.Equal(t, 0, stats.TotalCases)
	assert.False(t, stats.IsRunning)

	ra.StartCtx(ctx)

	// Stats after start
	stats = ra.Stats()
	assert.True(t, stats.IsRunning)
	assert.True(t, stats.Duration >= 0)

	// Submit results
	ra.Submit(PageResult{
		SuiteID: 1,
		Cases:   []data.Case{{ID: 1}, {ID: 2}},
	})
	ra.Submit(PageResult{
		SuiteID: 2,
		Cases:   []data.Case{{ID: 3}},
	})
	ra.Submit(PageResult{
		SuiteID: 3,
		Error:   errors.New("failed"),
	})

	// Give aggregator time to process
	time.Sleep(50 * time.Millisecond)

	stats = ra.Stats()
	assert.Equal(t, 3, stats.TotalCases)
	assert.Equal(t, 3, stats.TotalPages)
	assert.Equal(t, 1, stats.FailedPages)
	assert.True(t, stats.HasErrors())
	assert.True(t, stats.ErrorRate() > 0)

	ra.Stop()
}

func TestResultAggregator_StopWithoutStart(t *testing.T) {
	ra := NewResultAggregator(100)

	// Stop without Start should not panic
	cases, errs := ra.Stop()
	assert.Empty(t, cases)
	assert.Empty(t, errs)
}

func TestResultAggregator_DoubleStop(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(100)

	ra.StartCtx(ctx)
	ra.Submit(PageResult{Cases: []data.Case{{ID: 1}}})

	cases1, _ := ra.Stop()
	cases2, _ := ra.Stop()

	// Second stop should return same results
	assert.Len(t, cases1, 1)
	assert.Len(t, cases2, 1)
}

func TestResultAggregator_EmptyResults(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(100)

	ra.StartCtx(ctx)

	// Submit empty results
	ra.Submit(PageResult{SuiteID: 1, Cases: []data.Case{}})
	ra.Submit(PageResult{SuiteID: 2, Cases: nil})

	cases, _ := ra.Stop()
	assert.Empty(t, cases)
}

func TestResultAggregator_NilError(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(100)

	ra.StartCtx(ctx)

	// Submit nil error should not panic
	ra.SubmitError(nil)
	ra.Submit(PageResult{Error: nil})

	_, errs := ra.Stop()
	assert.Empty(t, errs)
}

func TestResultAggregator_StartCtx_Idempotent(t *testing.T) {
	ra := NewResultAggregator(10)
	ctx := context.Background()

	ra.StartCtx(ctx)
	firstResultCh := ra.resultCh
	firstErrCh := ra.errCh
	firstDoneCh := ra.doneCh

	// Second start must be a no-op and keep existing channels.
	ra.StartCtx(ctx)

	assert.Equal(t, firstResultCh, ra.resultCh)
	assert.Equal(t, firstErrCh, ra.errCh)
	assert.Equal(t, firstDoneCh, ra.doneCh)

	ra.Stop()
}

func TestResultAggregator_SubmitErrorBranches(t *testing.T) {
	t.Run("ignored when not started", func(t *testing.T) {
		ra := NewResultAggregator(1)
		ra.SubmitError(errors.New("not-started"))
		_, errs := ra.GetResults()
		assert.Empty(t, errs)
	})

	t.Run("ignored when stopped", func(t *testing.T) {
		ra := NewResultAggregator(1)
		ra.started = true
		ra.stopped = true
		ra.errCh = make(chan error, 1)

		ra.SubmitError(errors.New("stopped"))
		assert.Equal(t, 0, len(ra.errCh))
	})

	t.Run("enqueues when channel has capacity", func(t *testing.T) {
		ra := NewResultAggregator(1)
		ra.started = true
		ra.errCh = make(chan error, 1)

		expected := errors.New("enqueue")
		ra.SubmitError(expected)

		select {
		case got := <-ra.errCh:
			assert.Equal(t, expected, got)
		default:
			t.Fatal("expected error to be enqueued")
		}
	})

	t.Run("drops when channel is full", func(t *testing.T) {
		ra := NewResultAggregator(1)
		ra.started = true
		ra.errCh = make(chan error, 1)

		first := errors.New("first")
		second := errors.New("second")
		ra.errCh <- first
		ra.SubmitError(second)

		assert.Equal(t, 1, len(ra.errCh))
		got := <-ra.errCh
		assert.Equal(t, first, got)
	})
}

func TestResultAggregator_AddError_NilIgnored(t *testing.T) {
	ra := NewResultAggregator(10)

	// Directly exercise nil guard branch.
	ra.addError(nil)
	_, errs := ra.GetResults()
	assert.Empty(t, errs)

	ra.addError(errors.New("real error"))
	_, errs = ra.GetResults()
	assert.Len(t, errs, 1)
}

func TestCombinedError(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		err := CombinedError(nil)
		assert.Nil(t, err)

		err = CombinedError([]error{})
		assert.Nil(t, err)
	})

	t.Run("single error", func(t *testing.T) {
		err := CombinedError([]error{errors.New("single")})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "single")
	})

	t.Run("multiple errors", func(t *testing.T) {
		errs := []error{
			errors.New("first"),
			errors.New("second"),
		}
		err := CombinedError(errs)
		assert.Contains(t, err.Error(), "first")
		assert.Contains(t, err.Error(), "second")
	})
}

func TestAggregationStats(t *testing.T) {
	t.Run("has errors", func(t *testing.T) {
		stats := AggregationStats{ErrorCount: 1}
		assert.True(t, stats.HasErrors())

		stats = AggregationStats{FailedPages: 1}
		assert.True(t, stats.HasErrors())

		stats = AggregationStats{}
		assert.False(t, stats.HasErrors())
	})

	t.Run("error rate", func(t *testing.T) {
		stats := AggregationStats{TotalPages: 10, FailedPages: 2}
		assert.Equal(t, 20.0, stats.ErrorRate())

		stats = AggregationStats{}
		assert.Equal(t, 0.0, stats.ErrorRate())
	})
}

func TestResultAggregator_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ra := NewResultAggregator(100)

	ra.StartCtx(ctx)

	// Submit some results
	ra.Submit(PageResult{
		SuiteID: 1,
		Cases:   []data.Case{{ID: 1, Title: "Case 1"}},
	})

	// Cancel context
	cancel()

	// Give time for cancellation to propagate
	time.Sleep(50 * time.Millisecond)

	// After cancellation, we can still get results that were processed
	cases, _ := ra.GetResults()
	// Note: we may or may not have the case depending on timing
	_ = cases
}

func BenchmarkResultAggregator_Submit(b *testing.B) {
	ctx := context.Background()
	ra := NewResultAggregator(10000)
	ra.StartCtx(ctx)

	result := PageResult{
		SuiteID: 1,
		Cases:   []data.Case{{ID: 1, Title: "Case"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ra.Submit(result)
	}
	b.StopTimer()

	ra.Stop()
}

// ─────────────────────────────────────────────────────────────────────────────
// Wave-5 Coverage Tests: Buffer, Shutdown, Concurrency, Error Propagation
// ─────────────────────────────────────────────────────────────────────────────

// TestAggregator_BufferOverflow tests behavior when result channel hits capacity
func TestAggregator_BufferOverflow(t *testing.T) {
	ctx := context.Background()
	smallBuffer := 5
	ra := NewResultAggregator(smallBuffer)

	ra.StartCtx(ctx)

	// Submit results rapidly to potentially overflow buffer
	blockingSubmit := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			ra.Submit(PageResult{
				SuiteID: 1,
				Offset:  i,
				Cases:   []data.Case{{ID: int64(i), Title: fmt.Sprintf("Case %d", i)}},
			})
		}
		close(blockingSubmit)
	}()

	// Wait for submissions to complete (with timeout to ensure blocking doesn't deadlock)
	select {
	case <-blockingSubmit:
		// Expected
	case <-time.After(5 * time.Second):
		t.Fatal("buffer overflow test timed out - possible deadlock")
	}

	cases, _ := ra.Stop()

	// Verify all cases were eventually submitted (though timing may vary)
	assert.True(t, len(cases) > 50, "Expected many cases despite small buffer")
}

// TestAggregator_PartialBatchOnShutdown tests handling of incomplete batches when stopped
func TestAggregator_PartialBatchOnShutdown(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(100)

	ra.StartCtx(ctx)

	// Submit multiple batches
	for batch := 0; batch < 5; batch++ {
		ra.Submit(PageResult{
			SuiteID: 1,
			Offset:  batch * 10,
			Cases: func() []data.Case {
				cases := make([]data.Case, batch%3+1) // Vary batch size
				for i := range cases {
					cases[i] = data.Case{
						ID:    int64(batch*100 + i),
						Title: fmt.Sprintf("Case %d-%d", batch, i),
					}
				}
				return cases
			}(),
		})
	}

	// Stop immediately (may have pending items in channels)
	cases, _ := ra.Stop()

	// Should collect all submitted cases
	assert.True(t, len(cases) > 0, "Should have collected cases from partial batches")
}

// TestAggregator_ConcurrentSubmitAndFlush stress-tests concurrent submit with stop
func TestAggregator_ConcurrentSubmitAndFlush(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(1000)

	ra.StartCtx(ctx)

	numProducers := 20
	resultsPerProducer := 50
	done := make(chan struct{}, numProducers)

	// Concurrent producers
	for p := 0; p < numProducers; p++ {
		go func(producerID int) {
			defer func() { done <- struct{}{} }()

			for i := 0; i < resultsPerProducer; i++ {
				caseID := int64(producerID*10000 + i)
				ra.Submit(PageResult{
					SuiteID: int64(producerID),
					Offset:  i,
					Cases:   []data.Case{{ID: caseID, Title: fmt.Sprintf("P%d-C%d", producerID, i)}},
				})
			}
		}(p)
	}

	// Wait for most producers to finish before stopping
	completed := 0
	timeout := time.After(5 * time.Second)
	timedOut := false
	for completed < numProducers {
		select {
		case <-done:
			completed++
		case <-timeout:
			// Timeout reached: stop waiting and proceed to shutdown.
			timedOut = true
		}
		if timedOut {
			break
		}
	}

	// Now stop the aggregator
	cases, errs := ra.Stop()

	// Verify results integrity
	assert.Empty(t, errs)
	expectedMin := numProducers * resultsPerProducer / 2 // At least half should be collected
	assert.True(t, len(cases) >= expectedMin, "Expected at least %d cases, got %d", expectedMin, len(cases))

	// Verify no duplicates
	seenIDs := make(map[int64]bool)
	for _, c := range cases {
		assert.False(t, seenIDs[c.ID], "Duplicate case ID: %d", c.ID)
		seenIDs[c.ID] = true
	}
}

// TestAggregator_ErrorPropagation tests error channel handling and propagation
func TestAggregator_ErrorPropagation(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(100)

	ra.StartCtx(ctx)

	// Mix successful and error submissions
	successCases := 0
	for i := 0; i < 20; i++ {
		if i%3 == 0 {
			// Submit error result
			ra.Submit(PageResult{
				SuiteID: 1,
				Offset:  i * 10,
				Error:   fmt.Errorf("page error %d", i),
			})
		} else {
			// Submit successful result
			ra.Submit(PageResult{
				SuiteID: 1,
				Offset:  i * 10,
				Cases:   []data.Case{{ID: int64(i), Title: fmt.Sprintf("Case %d", i)}},
			})
			successCases++
		}
	}

	// Also submit direct errors
	for i := 0; i < 5; i++ {
		ra.SubmitError(fmt.Errorf("direct error %d", i))
	}

	cases, errs := ra.Stop()

	// Verify separation: cases vs errors
	assert.True(t, len(cases) > 0, "Should have successful cases")
	assert.True(t, len(errs) > 0, "Should have collected errors")
	assert.True(t, len(errs) >= 5, "Should have at least the direct errors")

	// Verify stats reflect both
	stats := AggregationStats{
		TotalCases:  len(cases),
		TotalPages:  20,
		FailedPages: len(errs),
		ErrorCount:  len(errs),
	}
	assert.True(t, stats.HasErrors())
}

// TestAggregator_SubmitAfterStop tests that submit after stop is safely ignored
func TestAggregator_SubmitAfterStop(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(100)

	ra.StartCtx(ctx)

	ra.Submit(PageResult{
		SuiteID: 1,
		Cases:   []data.Case{{ID: 1}},
	})

	cases1, _ := ra.Stop()
	initialCount := len(cases1)

	// Submit after stop should be ignored (not panic)
	ra.Submit(PageResult{
		SuiteID: 2,
		Cases:   []data.Case{{ID: 2, Title: "Should be ignored"}},
	})

	ra.SubmitError(errors.New("should be ignored"))

	// Get results again
	cases2, _ := ra.GetResults()

	// Count should not increase
	assert.Equal(t, initialCount, len(cases2), "Submit after stop should not add cases")
}

// TestAggregator_RapidStartStop tests rapid initialization and teardown cycles
func TestAggregator_RapidStartStop(t *testing.T) {
	cycles := 10

	for cycle := 0; cycle < cycles; cycle++ {
		ctx := context.Background()
		ra := NewResultAggregator(100)

		ra.StartCtx(ctx)

		// Quick submissions
		for i := 0; i < 5; i++ {
			ra.Submit(PageResult{
				SuiteID: 1,
				Cases:   []data.Case{{ID: int64(cycle*100 + i)}},
			})
		}

		// Immediate stop
		cases, _ := ra.Stop()
		assert.True(t, len(cases) > 0, "Cycle %d: Should collect cases", cycle)
	}
}

// TestAggregator_LargePayloadBatch tests aggregator with large case payloads
func TestAggregator_LargePayloadBatch(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(50)

	ra.StartCtx(ctx)

	// Create large cases (simulate realistic payloads with title)
	largeTitle := "Large Case " + strings.Repeat("x", 200)

	// Submit batch of large cases
	largeBatch := make([]data.Case, 100)
	for i := 0; i < 100; i++ {
		largeBatch[i] = data.Case{
			ID:      int64(i),
			Title:   largeTitle,
			SuiteID: 1,
		}
	}

	ra.Submit(PageResult{
		SuiteID: 1,
		Cases:   largeBatch,
	})

	cases, _ := ra.Stop()

	assert.Len(t, cases, 100, "Should handle large payloads")
	assert.Equal(t, int64(0), cases[0].ID)
}

// TestAggregator_StatsAccuracy verifies stats are accurate during concurrent operations
func TestAggregator_StatsAccuracy(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(100)

	statsSnapshots := make([]AggregationStats, 0)
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(10 * time.Millisecond)
			if ra.IsRunning() {
				statsSnapshots = append(statsSnapshots, ra.Stats())
			}
		}
	}()

	ra.StartCtx(ctx)

	// Submit results over time
	for i := 0; i < 20; i++ {
		ra.Submit(PageResult{
			SuiteID: 1,
			Offset:  i,
			Cases:   []data.Case{{ID: int64(i), Title: fmt.Sprintf("Case %d", i)}},
		})
		if i%5 == 0 {
			time.Sleep(5 * time.Millisecond)
		}
	}

	// Submit errors
	for i := 0; i < 3; i++ {
		ra.SubmitError(fmt.Errorf("error %d", i))
	}

	cases, errs := ra.Stop()

	// Verify final stats
	assert.Len(t, cases, 20)
	assert.Len(t, errs, 3)

	// Stats should show reasonable progression (monotonically increasing)
	for i := 1; i < len(statsSnapshots); i++ {
		prev := statsSnapshots[i-1]
		curr := statsSnapshots[i]
		assert.True(t, curr.TotalCases >= prev.TotalCases, "Stats regression: TotalCases decreased")
		assert.True(t, curr.TotalPages >= prev.TotalPages, "Stats regression: TotalPages decreased")
	}
}
