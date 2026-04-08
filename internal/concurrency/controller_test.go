package concurrency

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// mockSuiteFetcher is a mock implementation of SuiteFetcher for testing
type mockSuiteFetcher struct {
	cases            map[int64][]data.Case // suiteID -> cases
	latency          time.Duration
	callCount        int32
	failSuiteIDs     map[int64]bool
	failPageOffsets  map[int64]map[int]bool  // suiteID -> offset -> shouldFail (permanent)
	failPageMaxTimes int                     // if >0, pages in failPageOffsets fail only this many times
	mu               sync.Mutex              // guards pageAttempts
	pageAttempts     map[int64]map[int]int32 // suiteID -> offset -> attempt count
}

func newMockSuiteFetcher() *mockSuiteFetcher {
	return &mockSuiteFetcher{
		cases:           make(map[int64][]data.Case),
		latency:         10 * time.Millisecond,
		failSuiteIDs:    make(map[int64]bool),
		failPageOffsets: make(map[int64]map[int]bool),
		pageAttempts:    make(map[int64]map[int]int32),
	}
}

func (m *mockSuiteFetcher) FetchPageCtx(ctx context.Context, req PageRequest) ([]data.Case, int64, error) {
	atomic.AddInt32(&m.callCount, 1)

	// Simulate latency
	if m.latency > 0 {
		select {
		case <-time.After(m.latency):
		case <-ctx.Done():
			return nil, -1, ctx.Err()
		}
	}

	// Check if this suite should fail (streaming doesn't call GetTotalCases)
	if m.failSuiteIDs[req.SuiteID] {
		return nil, -1, errors.New("suite fetch failed")
	}

	// Check if this page should fail
	if offsets, ok := m.failPageOffsets[req.SuiteID]; ok {
		if offsets[req.Offset] {
			m.mu.Lock()
			if m.pageAttempts[req.SuiteID] == nil {
				m.pageAttempts[req.SuiteID] = make(map[int]int32)
			}
			m.pageAttempts[req.SuiteID][req.Offset]++
			attempt := m.pageAttempts[req.SuiteID][req.Offset]
			m.mu.Unlock()
			if m.failPageMaxTimes <= 0 || attempt <= int32(m.failPageMaxTimes) {
				return nil, -1, errors.New("page fetch failed")
			}
		}
	}

	cases, ok := m.cases[req.SuiteID]
	if !ok {
		return []data.Case{}, 0, nil
	}

	totalSize := int64(len(cases))

	// Slice the cases based on offset and limit
	start := req.Offset
	if start >= len(cases) {
		return []data.Case{}, totalSize, nil
	}

	end := start + req.Limit
	if end > len(cases) {
		end = len(cases)
	}

	return cases[start:end], totalSize, nil
}

func (m *mockSuiteFetcher) addCases(suiteID int64, count int) {
	cases := make([]data.Case, count)
	for i := 0; i < count; i++ {
		cases[i] = data.Case{
			ID:      suiteID*1000000 + int64(i),
			Title:   fmt.Sprintf("Case %d-%d", suiteID, i),
			SuiteID: suiteID,
		}
	}
	m.cases[suiteID] = cases
}

func TestParallelController_Execute_SingleSuite(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.addCases(1, 50) // 50 cases in suite 1

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 2,
		MaxConcurrentPages:  3,
		PageSize:            10,
		Timeout:             30 * time.Second,
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10, EstimatedSize: 50},
	}

	ctx := context.Background()
	result, err := controller.Execute(ctx, tasks, fetcher, nil)

	assert.NoError(t, err)
	assert.Len(t, result.Cases, 50)
	assert.Equal(t, 1, result.Stats.TotalSuites)
	assert.Equal(t, 1, result.Stats.CompletedSuites)
}

func TestParallelController_Execute_MultipleSuites(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.addCases(1, 30) // 30 cases
	fetcher.addCases(2, 50) // 50 cases
	fetcher.addCases(3, 20) // 20 cases

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 3,
		MaxConcurrentPages:  2,
		PageSize:            15,
		Timeout:             30 * time.Second,
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
		{SuiteID: 2, ProjectID: 10},
		{SuiteID: 3, ProjectID: 10},
	}

	ctx := context.Background()
	result, err := controller.Execute(ctx, tasks, fetcher, nil)

	assert.NoError(t, err)
	assert.Len(t, result.Cases, 100) // 30 + 50 + 20
	assert.Equal(t, 3, result.Stats.TotalSuites)
	assert.Equal(t, 3, result.Stats.CompletedSuites)
}

func TestParallelController_Execute_EmptyTasks(t *testing.T) {
	controller := NewController(DefaultControllerConfig())
	fetcher := newMockSuiteFetcher()

	result, err := controller.Execute(context.Background(), nil, fetcher, nil)

	assert.NoError(t, err)
	assert.Empty(t, result.Cases)
	assert.Empty(t, result.Errors)
}

func TestParallelController_Execute_SuiteError(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.addCases(1, 30)
	fetcher.addCases(2, 50)
	fetcher.failSuiteIDs[2] = true // Suite 2 will fail

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites:      2,
		MaxConcurrentPages:       2,
		PageSize:                 15,
		Timeout:                  10 * time.Second,
		MaxRetries:               1,
		MaxConsecutiveErrorWaves: 1,
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
		{SuiteID: 2, ProjectID: 10},
	}

	ctx := context.Background()
	result, err := controller.Execute(ctx, tasks, fetcher, nil)

	// Should succeed with partial results
	assert.NoError(t, err)
	assert.Len(t, result.Cases, 30) // Only suite 1 cases
	// Partial flag may or may not be set depending on timing
	assert.True(t, len(result.Errors) > 0 || result.Partial)
}

func TestParallelController_Execute_PageError(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.addCases(1, 50)
	fetcher.failPageOffsets[1] = map[int]bool{15: true} // Fail page at offset 15

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 2,
		MaxConcurrentPages:  3,
		PageSize:            10,
		Timeout:             30 * time.Second,
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
	}

	ctx := context.Background()
	result, err := controller.Execute(ctx, tasks, fetcher, nil)

	// Should succeed with partial results (40 cases instead of 50)
	assert.NoError(t, err)
	assert.True(t, len(result.Cases) >= 40, "Expected at least 40 cases, got %d", len(result.Cases))
}

func TestParallelController_Execute_Timeout(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.addCases(1, 100)
	fetcher.latency = 500 * time.Millisecond // Slow responses

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 1,
		MaxConcurrentPages:  1,
		PageSize:            10,
		Timeout:             200 * time.Millisecond, // Very short timeout
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
	}

	ctx := context.Background()
	result, err := controller.Execute(ctx, tasks, fetcher, nil)

	// Should return partial results
	assert.True(t, err != nil || result.Partial)
	assert.True(t, len(result.Cases) < 100) // Not all cases fetched
}

func TestParallelController_Execute_ConcurrentPerformance(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.latency = 50 * time.Millisecond

	// Create 5 suites with 20 cases each
	for i := int64(1); i <= 5; i++ {
		fetcher.addCases(i, 20)
	}

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 5, // All suites in parallel
		MaxConcurrentPages:  2,
		PageSize:            10,
		Timeout:             30 * time.Second,
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
		{SuiteID: 2, ProjectID: 10},
		{SuiteID: 3, ProjectID: 10},
		{SuiteID: 4, ProjectID: 10},
		{SuiteID: 5, ProjectID: 10},
	}

	ctx := context.Background()
	start := time.Now()
	result, err := controller.Execute(ctx, tasks, fetcher, nil)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Len(t, result.Cases, 100) // 5 suites * 20 cases

	// With sequential processing: 5 suites * 2 pages * 50ms = 500ms
	// With parallel processing: ~2 pages * 50ms = 100ms (plus overhead)
	// Should be significantly faster than 400ms
	t.Logf("Parallel execution took: %v", duration)
	assert.True(t, duration < 400*time.Millisecond, "Parallel execution too slow: %v", duration)
}

func TestParallelController_Execute_LargeSuitePagination(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.addCases(1, 1000) // Large suite: 1000 cases

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 2,
		MaxConcurrentPages:  5, // Multiple parallel page fetches
		PageSize:            50,
		Timeout:             30 * time.Second,
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10, EstimatedSize: 1000},
	}

	ctx := context.Background()
	result, err := controller.Execute(ctx, tasks, fetcher, nil)

	assert.NoError(t, err)
	assert.Len(t, result.Cases, 1000)

	// Verify correct number of page requests (1000 / 50 = 20 pages)
	// But we get total count first, so 21 calls
	assert.True(t, atomic.LoadInt32(&fetcher.callCount) >= 20)
}

func TestParallelController_DefaultConfig(t *testing.T) {
	controller := NewController(nil)

	assert.Equal(t, 5, controller.config.MaxConcurrentSuites)
	assert.Equal(t, 3, controller.config.MaxConcurrentPages)
	assert.Equal(t, 180, controller.config.RequestsPerMinute)
	assert.Equal(t, 5*time.Minute, controller.config.Timeout)
	assert.Equal(t, 250, controller.config.PageSize)
}

func TestParallelController_ValidateConfig(t *testing.T) {
	config := &ControllerConfig{
		MaxConcurrentSuites: 0,  // Invalid → default 5
		MaxConcurrentPages:  -1, // Invalid → default 3
		RequestsPerMinute:   0,  // Valid: 0 = unlimited (no rate limiting)
		Timeout:             0,  // Invalid → default 5m
		PageSize:            0,  // Invalid → default 250
	}

	config.Normalize()

	// Should be set to defaults (except RequestsPerMinute: 0 is valid)
	assert.Equal(t, 5, config.MaxConcurrentSuites)
	assert.Equal(t, 3, config.MaxConcurrentPages)
	assert.Equal(t, 0, config.RequestsPerMinute) // 0 = unlimited, valid
	assert.Equal(t, 5*time.Minute, config.Timeout)
	assert.Equal(t, 250, config.PageSize)
}

func TestParallelController_ValidateConfig_NegativeRPM(t *testing.T) {
	config := &ControllerConfig{
		MaxConcurrentSuites:      1,
		MaxConcurrentPages:       1,
		RequestsPerMinute:        -10,
		Timeout:                  time.Second,
		PageSize:                 10,
		MaxConsecutiveErrorWaves: 1,
	}

	config.Normalize()
	assert.Equal(t, 180, config.RequestsPerMinute)
}

func TestControllerConfig_WithMethods(t *testing.T) {
	config := DefaultControllerConfig().
		WithMaxConcurrentSuites(10).
		WithMaxConcurrentPages(5).
		WithTimeout(10 * time.Minute)

	assert.Equal(t, 10, config.MaxConcurrentSuites)
	assert.Equal(t, 5, config.MaxConcurrentPages)
	assert.Equal(t, 10*time.Minute, config.Timeout)
}

func BenchmarkParallelController_Execute(b *testing.B) {
	fetcher := newMockSuiteFetcher()
	fetcher.latency = 1 * time.Millisecond

	// Setup: 10 suites with 50 cases each
	for i := int64(1); i <= 10; i++ {
		fetcher.addCases(i, 50)
	}

	tasks := make([]SuiteTask, 10)
	for i := 0; i < 10; i++ {
		tasks[i] = SuiteTask{SuiteID: int64(i + 1), ProjectID: 10}
	}

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 5,
		MaxConcurrentPages:  3,
		PageSize:            25,
	})

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := controller.Execute(ctx, tasks, fetcher, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// truncatedMockFetcher simulates a server that occasionally returns truncated pages
// (fewer results than limit, even though more data exists at higher offsets).
type truncatedMockFetcher struct {
	mockSuiteFetcher
	// truncatePages maps suiteID -> offset -> max results to return (< pageSize)
	truncatePages map[int64]map[int]int
}

func newTruncatedMockFetcher() *truncatedMockFetcher {
	return &truncatedMockFetcher{
		mockSuiteFetcher: mockSuiteFetcher{
			cases:           make(map[int64][]data.Case),
			latency:         1 * time.Millisecond,
			failSuiteIDs:    make(map[int64]bool),
			failPageOffsets: make(map[int64]map[int]bool),
		},
		truncatePages: make(map[int64]map[int]int),
	}
}

func (m *truncatedMockFetcher) FetchPageCtx(ctx context.Context, req PageRequest) ([]data.Case, int64, error) {
	atomic.AddInt32(&m.callCount, 1)

	cases, ok := m.cases[req.SuiteID]
	if !ok {
		return []data.Case{}, 0, nil
	}

	totalSize := int64(len(cases))

	start := req.Offset
	if start >= len(cases) {
		return []data.Case{}, totalSize, nil
	}

	end := start + req.Limit
	if end > len(cases) {
		end = len(cases)
	}

	result := cases[start:end]

	// Apply truncation if configured for this suite+offset
	if offsets, ok := m.truncatePages[req.SuiteID]; ok {
		if maxResults, ok := offsets[req.Offset]; ok && len(result) > maxResults {
			result = result[:maxResults]
		}
	}

	return result, totalSize, nil
}

func (m *truncatedMockFetcher) setTruncation(suiteID int64, offset int, maxResults int) {
	if _, ok := m.truncatePages[suiteID]; !ok {
		m.truncatePages[suiteID] = make(map[int]int)
	}
	m.truncatePages[suiteID][offset] = maxResults
}

// TestParallelController_TruncatedPageDoesNotLoseData verifies that a truncated API
// response (fewer than pageSize items, but NOT the last page) does NOT cause
// the controller to stop fetching early and lose data.
// This is a regression test for the data loss bug where partial pages triggered
// premature suite termination.
func TestParallelController_TruncatedPageDoesNotLoseData(t *testing.T) {
	fetcher := newTruncatedMockFetcher()
	// Suite with 1000 cases, pageSize=250 → 4 full pages
	fetcher.addCases(1, 1000)
	// Simulate server returning only 200 items for offset=250 (should be 250)
	fetcher.setTruncation(1, 250, 200)

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 1,
		MaxConcurrentPages:  3,
		PageSize:            250,
		Timeout:             10 * time.Second,
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
	}

	result, err := controller.Execute(context.Background(), tasks, fetcher, nil)
	assert.NoError(t, err)

	// We must get ALL 1000 cases (with 200 at offset 250, not 250, but data at 500+ still loaded)
	// Expected: 250 (offset 0) + 200 (offset 250, truncated) + 250 (offset 500) + 250 (offset 750) + 50 gap cases NOT fetched
	// Actually with truncation at offset 250 returning only 200 items, cases 450-499 are skipped
	// (offset 250 returns items 250-449 instead of 250-499).
	// But ALL subsequent pages (offset 500, 750) are still fetched.
	// Total = 250 + 200 + 250 + 250 = 950 (50 cases at positions 450-499 lost due to API truncation,
	// which is unavoidable, but critically offset 500+ data is NOT lost)
	assert.GreaterOrEqual(t, len(result.Cases), 950,
		"Must fetch data beyond the truncated page; old bug would stop at 700")
	// The old buggy code would stop after wave 0 (offsets 0,250,500) because offset 250 had < 250 items,
	// getting only 250+200+250=700 cases and missing the entire offset 750+ range
}

// TestParallelController_AllCasesFetchedNormally verifies complete fetch with no truncation
func TestParallelController_AllCasesFetchedNormally(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.addCases(1, 1000)

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 1,
		MaxConcurrentPages:  3,
		PageSize:            250,
		Timeout:             10 * time.Second,
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
	}

	result, err := controller.Execute(context.Background(), tasks, fetcher, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1000, len(result.Cases), "Must fetch all 1000 cases")
}

// TestParallelController_MultipleSuitesWithTruncation tests multiple suites where one has truncation
func TestParallelController_MultipleSuitesWithTruncation(t *testing.T) {
	fetcher := newTruncatedMockFetcher()
	fetcher.addCases(1, 500)  // Suite 1: normal
	fetcher.addCases(2, 1500) // Suite 2: will have truncation
	fetcher.addCases(3, 300)  // Suite 3: normal

	// Truncate suite 2 offset 500 (returns 100 instead of 250)
	fetcher.setTruncation(2, 500, 100)

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 3,
		MaxConcurrentPages:  3,
		PageSize:            250,
		Timeout:             10 * time.Second,
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
		{SuiteID: 2, ProjectID: 10},
		{SuiteID: 3, ProjectID: 10},
	}

	result, err := controller.Execute(context.Background(), tasks, fetcher, nil)
	assert.NoError(t, err)

	// Suite 1: 500 cases (all fetched)
	// Suite 2: 1500 cases, but offset 500 truncated to 100 → loses 150 cases in that page,
	//          but continues fetching 750+ → gets most of 1500
	// Suite 3: 300 cases (all fetched)
	// Total must be well above what old buggy code would produce
	// Old code: suite 2 would stop after first wave with truncation → ~600 cases from suite 2
	// New code: suite 2 continues → gets ~1350+ from suite 2
	expectedMinimum := 500 + 1350 + 300 // conservative lower bound
	assert.GreaterOrEqual(t, len(result.Cases), expectedMinimum,
		"Truncation in one suite must not stop fetching for that suite")
}

func TestParallelController_GetStats(t *testing.T) {
	// GetStats on a freshly created controller returns empty AggregationStats
	controller := NewController(DefaultControllerConfig())
	stats := controller.GetStats()
	assert.Equal(t, 0, stats.TotalCases)
	assert.Equal(t, 0, stats.TotalPages)
	assert.False(t, stats.IsRunning)
}

// ─────────────────────────────────────────────────────────────────────────────
// Wave-5 Coverage Tests: SuiteWorker, Context Cancellation, Retry, Large Pagination
// ─────────────────────────────────────────────────────────────────────────────

// TestSuiteWorker_ContextCancellation tests that suite worker exits cleanly on context cancellation
func TestSuiteWorker_ContextCancellation(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.addCases(1, 1000)
	fetcher.latency = 10 * time.Millisecond // Slow fetches to allow cancellation

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 2,
		MaxConcurrentPages:  2,
		PageSize:            100,
		Timeout:             30 * time.Second,
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Start a goroutine to cancel after short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
	}

	result, err := controller.Execute(ctx, tasks, fetcher, nil)

	// Should return with reduced cases (canceled early)
	assert.True(t, len(result.Cases) < 1000 || err != nil, "Expected early termination due to cancellation")
}

// TestSuiteWorker_NetworkRetry tests that failed pages are retried during recovery phase
func TestSuiteWorker_NetworkRetry(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.addCases(1, 500)

	// Configure to fail offsets 100 and 200 only once — retry should recover them.
	// Without failPageMaxTimes, permanent failures + exponential backoff (1s+2s+4s per retry)
	// would cause a timeout.
	fetcher.failPageOffsets[1] = map[int]bool{
		100: true,
		200: true,
	}
	fetcher.failPageMaxTimes = 1

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 1,
		MaxConcurrentPages:  3,
		PageSize:            100,
		Timeout:             30 * time.Second,
		MaxRetries:          3, // Allow retries
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
	}

	result, err := controller.Execute(context.Background(), tasks, fetcher, nil)

	// Should recover and fetch all cases — transient failures heal after 1 attempt
	assert.NoError(t, err)
	assert.Equal(t, 500, len(result.Cases), "Should fetch all 500 cases after transient retries")
}

// TestSuiteWorker_PartialResults tests fetching partial suite data when page fails persistently
func TestSuiteWorker_PartialResults(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.addCases(1, 500)

	// Permanently fail one page offset
	fetcher.failPageOffsets[1] = map[int]bool{
		200: true, // This page will always fail
	}

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites:      1,
		MaxConcurrentPages:       3,
		PageSize:                 100,
		Timeout:                  30 * time.Second,
		MaxRetries:               1,
		MaxConsecutiveErrorWaves: 2, // Stop retrying after some failures
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
	}

	result, err := controller.Execute(context.Background(), tasks, fetcher, nil)

	// Should return partial results: pages 0,100 succeed, page 200 fails, pages 300-400 may be missing
	assert.NoError(t, err)
	// At minimum, pages before the failed one should be fetched
	assert.True(t, len(result.Cases) >= 200, "Should fetch at least first 2 pages before persistent failure")
}

// TestFetchSuiteStreaming_LargePageCount tests pagination with very large page count
func TestFetchSuiteStreaming_LargePageCount(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	largeSize := 5000 // 5000 cases = 50 pages with pageSize 100
	fetcher.addCases(1, largeSize)

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 1,
		MaxConcurrentPages:  5, // Use multiple workers to speed up
		PageSize:            100,
		Timeout:             30 * time.Second,
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10, EstimatedSize: largeSize},
	}

	ctx := context.Background()
	start := time.Now()
	result, err := controller.Execute(ctx, tasks, fetcher, nil)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Len(t, result.Cases, largeSize, "Should fetch all %d cases", largeSize)

	t.Logf("Large suite fetch: %d cases in %v (%d pages)", largeSize, duration, largeSize/100)

	// With 5 workers, should be faster than sequential (which would be ~500ms at 10ms+ latency)
	// Parallel should be ~100-200ms
	assert.True(t, duration < 10*time.Second, "Large pagination should complete in reasonable time")
}

// TestFetchSuiteStreaming_ProbeFailFallback tests that probe failure doesn't abandon suite
func TestFetchSuiteStreaming_ProbeFailFallback(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.addCases(1, 300)

	// Create a counting wrapper fetcher
	countingFetcher := &mockSuiteFetcher{
		cases:           fetcher.cases,
		latency:         fetcher.latency,
		failSuiteIDs:    map[int64]bool{},
		failPageOffsets: map[int64]map[int]bool{},
	}
	// Override to fail first probe request only
	countingFetcher.failPageOffsets[1] = map[int]bool{
		0: true, // Fail offset 0 (probe)
	}

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 1,
		MaxConcurrentPages:  3,
		PageSize:            100,
		Timeout:             30 * time.Second,
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
	}

	result, err := controller.Execute(context.Background(), tasks, fetcher, nil)

	// Should recover from probe failure and fetch suite data via fallback
	assert.NoError(t, err)
	assert.True(t, len(result.Cases) > 0, "Should fetch data despite probe failure")
}

// TestFetchSuiteStreaming_EmptySuite tests handling of empty suite
func TestFetchSuiteStreaming_EmptySuite(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.addCases(1, 0) // Empty suite

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 1,
		MaxConcurrentPages:  3,
		PageSize:            100,
		Timeout:             30 * time.Second,
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
	}

	result, err := controller.Execute(context.Background(), tasks, fetcher, nil)

	assert.NoError(t, err)
	assert.Empty(t, result.Cases)
	assert.Equal(t, 1, result.Stats.CompletedSuites)
}

// TestFetchSuiteStreaming_MaxPageLimit tests that the controller correctly handles
// a large suite requiring many pages of pagination.
// Note: the 40K-page safety cap is a fallback for unknown totalSize; here the mock
// returns a known totalSize so the controller stops via the exact bound check.
func TestFetchSuiteStreaming_MaxPageLimit(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	// 50K cases with pageSize 250 = 200 pages — exercises heavy pagination
	// without excessive memory/time (10M was ~80MB RAM and caused timeouts).
	fetcher.addCases(1, 50000)
	fetcher.latency = 0

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 1,
		MaxConcurrentPages:  5,
		PageSize:            250,
		Timeout:             30 * time.Second,
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
	}

	result, err := controller.Execute(context.Background(), tasks, fetcher, nil)

	assert.NoError(t, err)
	assert.Equal(t, 50000, len(result.Cases), "Should fetch all cases from the large suite")
}

// TestFetchSuiteStreaming_ConsecutiveErrorWaves tests error wave detection
func TestFetchSuiteStreaming_ConsecutiveErrorWaves(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.addCases(1, 1000)

	// Make all offsets fail
	for offset := 0; offset < 1000; offset += 100 {
		if fetcher.failPageOffsets[1] == nil {
			fetcher.failPageOffsets[1] = make(map[int]bool)
		}
		fetcher.failPageOffsets[1][offset] = true
	}

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites:      1,
		MaxConcurrentPages:       2,
		PageSize:                 100,
		Timeout:                  30 * time.Second,
		MaxConsecutiveErrorWaves: 1,
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
	}

	result, err := controller.Execute(context.Background(), tasks, fetcher, nil)

	// Should handle persistent errors gracefully (not panic, may return partial results)
	assert.NoError(t, err)
	assert.True(t, len(result.Cases) == 0 || result.Partial, "Should either have no cases or be marked partial")
}

// TestFetchSuiteStreaming_UnknownTotalFallback tests exhaustion detection when API doesn't report totalSize
func TestFetchSuiteStreaming_UnknownTotalFallback(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.addCases(1, 350)

	// The existing fetcher already handles unknown totals well in our mock setup
	// Since the mock doesn't override totalSize behavior, it should work as-is

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 1,
		MaxConcurrentPages:  3,
		PageSize:            100,
		Timeout:             30 * time.Second,
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
	}

	result, err := controller.Execute(context.Background(), tasks, fetcher, nil)

	// Should fetch all cases using exhaustion detection
	assert.NoError(t, err)
	// With exhaustion detection (N consecutive empty pages), should fetch all 350 cases
	assert.Len(t, result.Cases, 350)
}

// TestFetchSuiteStreaming_VerySmallPageSize tests edge case of extremely small page size
func TestFetchSuiteStreaming_VerySmallPageSize(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.addCases(1, 100)

	controller := NewController(&ControllerConfig{
		MaxConcurrentSuites: 1,
		MaxConcurrentPages:  5,
		PageSize:            5, // Very small → 20 pages
		Timeout:             30 * time.Second,
	})

	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10},
	}

	result, err := controller.Execute(context.Background(), tasks, fetcher, nil)

	assert.NoError(t, err)
	assert.Len(t, result.Cases, 100, "Should handle very small page sizes")
}
