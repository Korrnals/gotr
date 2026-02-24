package parallel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestResultAggregator_Basic(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(100)

	ra.Start(ctx)

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

func TestResultAggregator_Deduplication(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(100)

	ra.Start(ctx)

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

	ra.Start(ctx)

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

	ra.Start(ctx)

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

	ra.Start(ctx)

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

	ra.Start(ctx)
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

	ra.Start(ctx)

	// Submit empty results
	ra.Submit(PageResult{SuiteID: 1, Cases: []data.Case{}})
	ra.Submit(PageResult{SuiteID: 2, Cases: nil})

	cases, _ := ra.Stop()
	assert.Empty(t, cases)
}

func TestResultAggregator_NilError(t *testing.T) {
	ctx := context.Background()
	ra := NewResultAggregator(100)

	ra.Start(ctx)

	// Submit nil error should not panic
	ra.SubmitError(nil)
	ra.Submit(PageResult{Error: nil})

	_, errs := ra.Stop()
	assert.Empty(t, errs)
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

	ra.Start(ctx)

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
	ra.Start(ctx)

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
