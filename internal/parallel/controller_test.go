package parallel

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

// mockSuiteFetcher is a mock implementation of SuiteFetcher for testing
type mockSuiteFetcher struct {
	cases           map[int64][]data.Case // suiteID -> cases
	latency         time.Duration
	failAfter       int
	callCount       int32
	failSuiteIDs    map[int64]bool
	failPageOffsets map[int64]map[int]bool // suiteID -> offset -> shouldFail
}

func newMockSuiteFetcher() *mockSuiteFetcher {
	return &mockSuiteFetcher{
		cases:           make(map[int64][]data.Case),
		latency:         10 * time.Millisecond,
		failSuiteIDs:    make(map[int64]bool),
		failPageOffsets: make(map[int64]map[int]bool),
	}
}

func (m *mockSuiteFetcher) FetchPage(ctx context.Context, req PageRequest) ([]data.Case, error) {
	atomic.AddInt32(&m.callCount, 1)

	// Simulate latency
	if m.latency > 0 {
		select {
		case <-time.After(m.latency):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Check if this page should fail
	if offsets, ok := m.failPageOffsets[req.SuiteID]; ok {
		if offsets[req.Offset] {
			return nil, errors.New("page fetch failed")
		}
	}

	cases, ok := m.cases[req.SuiteID]
	if !ok {
		return []data.Case{}, nil
	}

	// Slice the cases based on offset and limit
	start := req.Offset
	if start >= len(cases) {
		return []data.Case{}, nil
	}

	end := start + req.Limit
	if end > len(cases) {
		end = len(cases)
	}

	return cases[start:end], nil
}

func (m *mockSuiteFetcher) GetTotalCases(ctx context.Context, projectID int64, suiteID int64) (int, error) {
	if m.failSuiteIDs[suiteID] {
		return 0, errors.New("get total cases failed")
	}
	return len(m.cases[suiteID]), nil
}

func (m *mockSuiteFetcher) addCases(suiteID int64, count int) {
	cases := make([]data.Case, count)
	for i := 0; i < count; i++ {
		cases[i] = data.Case{
			ID:      suiteID*1000000 + int64(i),
			Title:   "Case",
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
		MaxConcurrentSuites: 2,
		MaxConcurrentPages:  2,
		PageSize:            15,
		Timeout:             30 * time.Second,
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

func TestParallelController_estimateSize(t *testing.T) {
	fetcher := newMockSuiteFetcher()
	fetcher.addCases(1, 150)

	controller := NewController(DefaultControllerConfig())

	// Test with GetTotalCases working
	task := SuiteTask{SuiteID: 1, ProjectID: 10}
	size := controller.estimateSize(context.Background(), task, fetcher)
	assert.Equal(t, 150, size)

	// Test with GetTotalCases failing
	fetcher.failSuiteIDs[2] = true
	task = SuiteTask{SuiteID: 2, ProjectID: 10}
	size = controller.estimateSize(context.Background(), task, fetcher)
	assert.True(t, size > 0) // Should return some default
}

func TestParallelController_DefaultConfig(t *testing.T) {
	controller := NewController(nil)

	assert.Equal(t, 5, controller.config.MaxConcurrentSuites)
	assert.Equal(t, 3, controller.config.MaxConcurrentPages)
	assert.Equal(t, 150, controller.config.RequestsPerMinute)
	assert.Equal(t, 5*time.Minute, controller.config.Timeout)
	assert.Equal(t, 250, controller.config.PageSize)
}

func TestParallelController_ValidateConfig(t *testing.T) {
	config := &ControllerConfig{
		MaxConcurrentSuites: 0, // Invalid
		MaxConcurrentPages:  -1, // Invalid
		RequestsPerMinute:   0,  // Invalid
		Timeout:             0,  // Invalid
		PageSize:            0,  // Invalid
	}

	config.Validate()

	// Should be set to defaults
	assert.Equal(t, 5, config.MaxConcurrentSuites)
	assert.Equal(t, 3, config.MaxConcurrentPages)
	assert.Equal(t, 150, config.RequestsPerMinute)
	assert.Equal(t, 5*time.Minute, config.Timeout)
	assert.Equal(t, 250, config.PageSize)
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
