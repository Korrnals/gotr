package concurrency

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// testItem is a simple type for testing generic fetch strategies.
type testItem struct {
	ID   int64
	Name string
}

// mockReporter tracks ProgressReporter calls for verification.
type mockReporter struct {
	itemsCompleted atomic.Int32
	batchesTotal   atomic.Int32
	errorsTotal    atomic.Int32
}

func (m *mockReporter) OnItemComplete()    { m.itemsCompleted.Add(1) }
func (m *mockReporter) OnBatchReceived(n int) { m.batchesTotal.Add(int32(n)) }
func (m *mockReporter) OnError()           { m.errorsTotal.Add(1) }

// --- FetchParallel tests ---

func TestFetchParallel_Basic(t *testing.T) {
	ctx := context.Background()
	projectIDs := []int64{10, 20, 30}

	results, err := FetchParallel(ctx, projectIDs,
		func(pid int64) ([]testItem, error) {
			return []testItem{
				{ID: pid*10 + 1, Name: fmt.Sprintf("item-%d-1", pid)},
				{ID: pid*10 + 2, Name: fmt.Sprintf("item-%d-2", pid)},
			}, nil
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 project results, got %d", len(results))
	}

	for _, pid := range projectIDs {
		items, ok := results[pid]
		if !ok {
			t.Fatalf("missing results for project %d", pid)
		}
		if len(items) != 2 {
			t.Fatalf("expected 2 items for project %d, got %d", pid, len(items))
		}
	}
}

func TestFetchParallel_Empty(t *testing.T) {
	ctx := context.Background()
	results, err := FetchParallel(ctx, []int64{},
		func(pid int64) ([]testItem, error) {
			t.Fatal("should not be called")
			return nil, nil
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty map, got %d entries", len(results))
	}
}

func TestFetchParallel_ErrorStopsAll(t *testing.T) {
	ctx := context.Background()

	_, err := FetchParallel(ctx, []int64{1, 2, 3},
		func(pid int64) ([]testItem, error) {
			if pid == 2 {
				return nil, fmt.Errorf("api error for project %d", pid)
			}
			return []testItem{{ID: pid}}, nil
		},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFetchParallel_ContinueOnError(t *testing.T) {
	ctx := context.Background()
	reporter := &mockReporter{}

	results, err := FetchParallel(ctx, []int64{1, 2, 3},
		func(pid int64) ([]testItem, error) {
			if pid == 2 {
				return nil, fmt.Errorf("api error for project %d", pid)
			}
			return []testItem{{ID: pid}}, nil
		},
		WithContinueOnError(),
		WithReporter(reporter),
	)

	// Should return partial results with error
	if err == nil {
		t.Fatal("expected error with continue-on-error, got nil")
	}

	// Should have results for projects 1 and 3
	if len(results) != 2 {
		t.Fatalf("expected 2 project results (partial), got %d", len(results))
	}

	// Reporter should have recorded the error
	if reporter.errorsTotal.Load() != 1 {
		t.Fatalf("expected 1 error reported, got %d", reporter.errorsTotal.Load())
	}

	// Reporter should have 2 items completed
	if reporter.itemsCompleted.Load() != 2 {
		t.Fatalf("expected 2 items completed, got %d", reporter.itemsCompleted.Load())
	}
}

func TestFetchParallel_WithReporter(t *testing.T) {
	ctx := context.Background()
	reporter := &mockReporter{}

	_, err := FetchParallel(ctx, []int64{1, 2},
		func(pid int64) ([]testItem, error) {
			return []testItem{{ID: 1}, {ID: 2}, {ID: 3}}, nil
		},
		WithReporter(reporter),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reporter.itemsCompleted.Load() != 2 {
		t.Fatalf("expected 2 items completed, got %d", reporter.itemsCompleted.Load())
	}

	if reporter.batchesTotal.Load() != 6 { // 3 items × 2 projects
		t.Fatalf("expected 6 total batch items, got %d", reporter.batchesTotal.Load())
	}
}

func TestFetchParallel_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel

	_, err := FetchParallel(ctx, []int64{1, 2, 3},
		func(pid int64) ([]testItem, error) {
			// Check context
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				time.Sleep(100 * time.Millisecond)
				return []testItem{{ID: pid}}, nil
			}
		},
	)

	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestFetchParallel_MaxConcurrency(t *testing.T) {
	ctx := context.Background()
	var concurrent atomic.Int32
	var maxConcurrent atomic.Int32

	_, err := FetchParallel(ctx, []int64{1, 2, 3, 4, 5},
		func(pid int64) ([]testItem, error) {
			cur := concurrent.Add(1)
			// Track max concurrency
			for {
				old := maxConcurrent.Load()
				if cur <= old || maxConcurrent.CompareAndSwap(old, cur) {
					break
				}
			}
			time.Sleep(50 * time.Millisecond)
			concurrent.Add(-1)
			return []testItem{{ID: pid}}, nil
		},
		WithMaxConcurrency(2),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if maxConcurrent.Load() > 2 {
		t.Fatalf("expected max concurrency 2, got %d", maxConcurrent.Load())
	}
}

// --- FetchParallelBySuite tests ---

func TestFetchParallelBySuite_Basic(t *testing.T) {
	ctx := context.Background()
	suiteIDs := []int64{100, 200, 300}

	items, err := FetchParallelBySuite(ctx, suiteIDs,
		func(suiteID int64) ([]testItem, error) {
			return []testItem{
				{ID: suiteID*10 + 1, Name: fmt.Sprintf("section-%d-1", suiteID)},
				{ID: suiteID*10 + 2, Name: fmt.Sprintf("section-%d-2", suiteID)},
			}, nil
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 3 suites × 2 items = 6 total
	if len(items) != 6 {
		t.Fatalf("expected 6 items, got %d", len(items))
	}
}

func TestFetchParallelBySuite_Empty(t *testing.T) {
	ctx := context.Background()
	items, err := FetchParallelBySuite(ctx, []int64{},
		func(suiteID int64) ([]testItem, error) {
			t.Fatal("should not be called")
			return nil, nil
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if items != nil {
		t.Fatalf("expected nil, got %v", items)
	}
}

func TestFetchParallelBySuite_ErrorStopsAll(t *testing.T) {
	ctx := context.Background()

	_, err := FetchParallelBySuite(ctx, []int64{1, 2, 3},
		func(suiteID int64) ([]testItem, error) {
			if suiteID == 2 {
				return nil, fmt.Errorf("suite error %d", suiteID)
			}
			return []testItem{{ID: suiteID}}, nil
		},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFetchParallelBySuite_ContinueOnError(t *testing.T) {
	ctx := context.Background()
	reporter := &mockReporter{}

	items, err := FetchParallelBySuite(ctx, []int64{1, 2, 3},
		func(suiteID int64) ([]testItem, error) {
			if suiteID == 2 {
				return nil, fmt.Errorf("suite error %d", suiteID)
			}
			return []testItem{{ID: suiteID}}, nil
		},
		WithContinueOnError(),
		WithReporter(reporter),
	)

	if err == nil {
		t.Fatal("expected error with continue-on-error, got nil")
	}

	// Should have partial results from suites 1 and 3
	if len(items) != 2 {
		t.Fatalf("expected 2 items (partial), got %d", len(items))
	}

	if reporter.errorsTotal.Load() != 1 {
		t.Fatalf("expected 1 error, got %d", reporter.errorsTotal.Load())
	}
}

func TestFetchParallelBySuite_WithReporter(t *testing.T) {
	ctx := context.Background()
	reporter := &mockReporter{}

	_, err := FetchParallelBySuite(ctx, []int64{1, 2},
		func(suiteID int64) ([]testItem, error) {
			return []testItem{{ID: 1}, {ID: 2}}, nil
		},
		WithReporter(reporter),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reporter.itemsCompleted.Load() != 2 {
		t.Fatalf("expected 2 items completed, got %d", reporter.itemsCompleted.Load())
	}

	if reporter.batchesTotal.Load() != 4 { // 2 items × 2 suites
		t.Fatalf("expected 4 total batch items, got %d", reporter.batchesTotal.Load())
	}
}
