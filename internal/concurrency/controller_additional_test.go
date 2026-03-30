package concurrency

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/concurrent"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

type probeFallbackFetcher struct {
	mu sync.Mutex
	// failOffset0FirstN controls how many first calls for offset=0 fail.
	failOffset0FirstN int
	callOffset0       int
	cases             []data.Case
}

func (f *probeFallbackFetcher) FetchPageCtx(_ context.Context, req PageRequest) ([]data.Case, int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if req.Offset == 0 && f.callOffset0 < f.failOffset0FirstN {
		f.callOffset0++
		return nil, -1, errors.New("probe failed")
	}

	total := int64(len(f.cases))
	if req.Offset >= len(f.cases) {
		return []data.Case{}, total, nil
	}

	end := req.Offset + req.Limit
	if end > len(f.cases) {
		end = len(f.cases)
	}
	return f.cases[req.Offset:end], total, nil
}

type permanentPageErrorFetcher struct {
	failOffset int
	cases      []data.Case
}

func (f *permanentPageErrorFetcher) FetchPageCtx(_ context.Context, req PageRequest) ([]data.Case, int64, error) {
	if req.Offset == f.failOffset {
		return nil, int64(len(f.cases)), errors.New("permanent page failure")
	}
	if req.Offset >= len(f.cases) {
		return []data.Case{}, int64(len(f.cases)), nil
	}
	end := req.Offset + req.Limit
	if end > len(f.cases) {
		end = len(f.cases)
	}
	return f.cases[req.Offset:end], int64(len(f.cases)), nil
}

func makeCases(n int, suiteID int64) []data.Case {
	out := make([]data.Case, n)
	for i := 0; i < n; i++ {
		out[i] = data.Case{ID: int64(i + 1), SuiteID: suiteID}
	}
	return out
}

type unknownTotalSinglePageFetcher struct {
	mu    sync.Mutex
	calls int
	cases []data.Case
}

func (f *unknownTotalSinglePageFetcher) FetchPageCtx(_ context.Context, req PageRequest) ([]data.Case, int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++

	if req.Offset > 0 {
		return []data.Case{}, -1, nil
	}

	end := req.Limit
	if end > len(f.cases) {
		end = len(f.cases)
	}
	return f.cases[:end], -1, nil
}

func (f *unknownTotalSinglePageFetcher) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

type probeFailThenKnownTotalFetcher struct {
	probeCalls int32
	cases      []data.Case
}

func (f *probeFailThenKnownTotalFetcher) FetchPageCtx(_ context.Context, req PageRequest) ([]data.Case, int64, error) {
	if req.Offset == 0 {
		if atomic.AddInt32(&f.probeCalls, 1) == 1 {
			return nil, -1, errors.New("probe failed")
		}
		end := req.Limit
		if end > len(f.cases) {
			end = len(f.cases)
		}
		return f.cases[:end], int64(len(f.cases)), nil
	}

	if req.Offset >= len(f.cases) {
		return []data.Case{}, int64(len(f.cases)), nil
	}
	end := req.Offset + req.Limit
	if end > len(f.cases) {
		end = len(f.cases)
	}
	return f.cases[req.Offset:end], int64(len(f.cases)), nil
}

type alwaysFailFetcher struct{}

func (f *alwaysFailFetcher) FetchPageCtx(_ context.Context, _ PageRequest) ([]data.Case, int64, error) {
	return nil, -1, errors.New("temporary outage")
}

type recordingKnownTotalFetcher struct {
	mu      sync.Mutex
	offsets []int
	total   int
	cases   []data.Case
}

func (f *recordingKnownTotalFetcher) FetchPageCtx(_ context.Context, req PageRequest) ([]data.Case, int64, error) {
	f.mu.Lock()
	f.offsets = append(f.offsets, req.Offset)
	f.mu.Unlock()

	if req.Offset >= f.total {
		return []data.Case{}, int64(f.total), nil
	}
	end := req.Offset + req.Limit
	if end > f.total {
		end = f.total
	}
	return f.cases[req.Offset:end], int64(f.total), nil
}

func (f *recordingKnownTotalFetcher) Offsets() []int {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]int, len(f.offsets))
	copy(out, f.offsets)
	return out
}

type unknownTotalEmptyPagesFetcher struct {
	cases []data.Case
}

func (f *unknownTotalEmptyPagesFetcher) FetchPageCtx(_ context.Context, req PageRequest) ([]data.Case, int64, error) {
	if req.Offset >= len(f.cases) {
		return []data.Case{}, -1, nil
	}
	end := req.Offset + req.Limit
	if end > len(f.cases) {
		end = len(f.cases)
	}
	return f.cases[req.Offset:end], -1, nil
}

type transientPageErrorFetcher struct {
	mu            sync.Mutex
	failOnceByOff map[int]bool
	cases         []data.Case
}

func (f *transientPageErrorFetcher) FetchPageCtx(_ context.Context, req PageRequest) ([]data.Case, int64, error) {
	f.mu.Lock()
	if f.failOnceByOff[req.Offset] {
		delete(f.failOnceByOff, req.Offset)
		f.mu.Unlock()
		return nil, int64(len(f.cases)), errors.New("transient page failure")
	}
	f.mu.Unlock()

	if req.Offset >= len(f.cases) {
		return []data.Case{}, int64(len(f.cases)), nil
	}
	end := req.Offset + req.Limit
	if end > len(f.cases) {
		end = len(f.cases)
	}
	return f.cases[req.Offset:end], int64(len(f.cases)), nil
}

type probeFailsThenKnownTotalFetcher struct {
	mu             sync.Mutex
	probeFailed    bool
	parallelOffset []int
	total          int
	cases          []data.Case
}

type emptyKnownTotalFetcher struct{}

func (f *emptyKnownTotalFetcher) FetchPageCtx(_ context.Context, _ PageRequest) ([]data.Case, int64, error) {
	return []data.Case{}, 0, nil
}

func (f *probeFailsThenKnownTotalFetcher) FetchPageCtx(_ context.Context, req PageRequest) ([]data.Case, int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if req.Offset == 0 && !f.probeFailed {
		f.probeFailed = true
		return nil, -1, errors.New("probe failure")
	}

	f.parallelOffset = append(f.parallelOffset, req.Offset)

	if req.Offset >= f.total {
		return []data.Case{}, int64(f.total), nil
	}
	end := req.Offset + req.Limit
	if end > f.total {
		end = f.total
	}
	return f.cases[req.Offset:end], int64(f.total), nil
}

func (f *probeFailsThenKnownTotalFetcher) Offsets() []int {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]int, len(f.parallelOffset))
	copy(out, f.parallelOffset)
	return out
}

func TestFetchPageWithRetry_CanceledContext(t *testing.T) {
	pc := NewController(DefaultControllerConfig())
	pc.limiter = concurrent.NewAdaptiveRateLimiter(180)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := pc.fetchPageWithRetry(ctx, PageRequest{SuiteTask: SuiteTask{SuiteID: 1}, Limit: 10, PageNum: 1}, &mockSuiteFetcher{})
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "context canceled")
}

func TestSuiteWorker_ContextCanceled(t *testing.T) {
	pc := NewController(DefaultControllerConfig())
	pq := NewPriorityQueue()
	pq.Close()

	agg := NewResultAggregator(10)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var completed int32
	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex
	var expectedCasesTotal int64
	var suitesWithTotal int32
	var suitesVerified int32
	var suiteResultsMu sync.Mutex
	suiteResults := make([]SuiteResultInfo, 0)

	err := pc.suiteWorker(
		ctx,
		pq,
		&mockSuiteFetcher{},
		agg,
		&completed,
		1,
		&failedPagesMu,
		&failedPages,
		&expectedCasesTotal,
		&suitesWithTotal,
		&suitesVerified,
		&suiteResultsMu,
		&suiteResults,
	)

	assert.ErrorIs(t, err, context.Canceled)
}

func TestSuiteWorker_QueueClosedReturnsNil(t *testing.T) {
	pc := NewController(DefaultControllerConfig())
	pq := NewPriorityQueue()
	pq.Close()

	agg := NewResultAggregator(10)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	var completed int32
	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex
	var expectedCasesTotal int64
	var suitesWithTotal int32
	var suitesVerified int32
	var suiteResultsMu sync.Mutex
	suiteResults := make([]SuiteResultInfo, 0)

	err := pc.suiteWorker(
		context.Background(),
		pq,
		&mockSuiteFetcher{},
		agg,
		&completed,
		1,
		&failedPagesMu,
		&failedPages,
		&expectedCasesTotal,
		&suitesWithTotal,
		&suitesVerified,
		&suiteResultsMu,
		&suiteResults,
	)

	assert.NoError(t, err)
	assert.Equal(t, int32(0), completed)
	assert.Empty(t, suiteResults)
	assert.Empty(t, failedPages)
}

func TestFetchSuiteStreaming_ProbeFailFallbackThenSuccess(t *testing.T) {
	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       2,
		PageSize:                 10,
		Timeout:                  5 * time.Second,
		RequestsPerMinute:        180,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 1,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(pc.config.RequestsPerMinute)

	fetcher := &probeFallbackFetcher{
		failOffset0FirstN: 1,
		cases:             makeCases(20, 1),
	}

	agg := NewResultAggregator(50)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex

	fetched, expected, verified, err := pc.fetchSuiteStreaming(
		context.Background(),
		SuiteTask{SuiteID: 1, ProjectID: 10, EstimatedSize: 20},
		fetcher,
		agg,
		&failedPagesMu,
		&failedPages,
	)

	assert.NoError(t, err)
	assert.Equal(t, 20, fetched)
	assert.Equal(t, int64(20), expected)
	assert.True(t, verified)
	assert.Empty(t, failedPages)
}

func TestFetchSuiteStreaming_EmptyKnownTotalOnProbe(t *testing.T) {
	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       2,
		PageSize:                 10,
		Timeout:                  5 * time.Second,
		RequestsPerMinute:        180,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 1,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(pc.config.RequestsPerMinute)

	agg := NewResultAggregator(10)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex

	fetched, expected, verified, err := pc.fetchSuiteStreaming(
		context.Background(),
		SuiteTask{SuiteID: 1, ProjectID: 10},
		&emptyKnownTotalFetcher{},
		agg,
		&failedPagesMu,
		&failedPages,
	)

	assert.NoError(t, err)
	assert.Equal(t, 0, fetched)
	assert.Equal(t, int64(0), expected)
	assert.True(t, verified)
	assert.Empty(t, failedPages)
}

func TestFetchSuiteStreaming_PermanentRecoveryFailure(t *testing.T) {
	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       2,
		PageSize:                 10,
		Timeout:                  5 * time.Second,
		RequestsPerMinute:        180,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 1,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(pc.config.RequestsPerMinute)

	fetcher := &permanentPageErrorFetcher{
		failOffset: 10,
		cases:      makeCases(20, 1),
	}

	agg := NewResultAggregator(50)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex

	fetched, expected, verified, err := pc.fetchSuiteStreaming(
		context.Background(),
		SuiteTask{SuiteID: 1, ProjectID: 10, EstimatedSize: 20},
		fetcher,
		agg,
		&failedPagesMu,
		&failedPages,
	)

	assert.NoError(t, err)
	assert.Equal(t, 10, fetched)
	assert.Equal(t, int64(20), expected)
	assert.False(t, verified)
	assert.Len(t, failedPages, 1)
	assert.Equal(t, int64(1), failedPages[0].SuiteID)
	assert.Equal(t, 10, failedPages[0].Offset)
}

func TestFetchSuiteStreaming_UnknownTotalSinglePageCompletes(t *testing.T) {
	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       3,
		PageSize:                 10,
		Timeout:                  5 * time.Second,
		RequestsPerMinute:        180,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 1,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(pc.config.RequestsPerMinute)

	fetcher := &unknownTotalSinglePageFetcher{cases: makeCases(7, 1)}

	agg := NewResultAggregator(20)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex

	fetched, expected, verified, err := pc.fetchSuiteStreaming(
		context.Background(),
		SuiteTask{SuiteID: 1, ProjectID: 10, EstimatedSize: 7},
		fetcher,
		agg,
		&failedPagesMu,
		&failedPages,
	)

	assert.NoError(t, err)
	assert.Equal(t, 7, fetched)
	assert.Equal(t, int64(-1), expected)
	assert.True(t, verified)
	assert.Empty(t, failedPages)
	assert.Equal(t, 1, fetcher.CallCount())
}

func TestFetchSuiteStreaming_ProbeFallbackSetsKnownTotalInParallel(t *testing.T) {
	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       1,
		PageSize:                 10,
		Timeout:                  5 * time.Second,
		RequestsPerMinute:        180,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 1,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(pc.config.RequestsPerMinute)

	fetcher := &probeFailThenKnownTotalFetcher{cases: makeCases(5, 1)}

	agg := NewResultAggregator(20)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex

	fetched, expected, verified, err := pc.fetchSuiteStreaming(
		context.Background(),
		SuiteTask{SuiteID: 1, ProjectID: 10, EstimatedSize: 5},
		fetcher,
		agg,
		&failedPagesMu,
		&failedPages,
	)

	assert.NoError(t, err)
	assert.Equal(t, 5, fetched)
	assert.Equal(t, int64(5), expected)
	assert.True(t, verified)
	assert.Empty(t, failedPages)
}

func TestFetchSuiteStreaming_StopsOnConsecutiveErrorsAndMarksFailedPage(t *testing.T) {
	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       1,
		PageSize:                 10,
		Timeout:                  5 * time.Second,
		RequestsPerMinute:        180,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 1,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(pc.config.RequestsPerMinute)

	fetcher := &alwaysFailFetcher{}

	agg := NewResultAggregator(20)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex

	fetched, expected, verified, err := pc.fetchSuiteStreaming(
		context.Background(),
		SuiteTask{SuiteID: 1, ProjectID: 10},
		fetcher,
		agg,
		&failedPagesMu,
		&failedPages,
	)

	assert.NoError(t, err)
	assert.Equal(t, 0, fetched)
	assert.Equal(t, int64(-1), expected)
	assert.False(t, verified)
	assert.Len(t, failedPages, 1)
	assert.Equal(t, int64(1), failedPages[0].SuiteID)
	assert.Equal(t, 0, failedPages[0].Offset)
}

func TestFetchSuiteStreaming_KnownTotalCapsWorkersAndStopsByBound(t *testing.T) {
	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       8,
		PageSize:                 10,
		Timeout:                  5 * time.Second,
		RequestsPerMinute:        180,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 1,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(pc.config.RequestsPerMinute)

	fetcher := &recordingKnownTotalFetcher{cases: makeCases(15, 1), total: 15}

	agg := NewResultAggregator(20)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex

	fetched, expected, verified, err := pc.fetchSuiteStreaming(
		context.Background(),
		SuiteTask{SuiteID: 1, ProjectID: 10, EstimatedSize: 15},
		fetcher,
		agg,
		&failedPagesMu,
		&failedPages,
	)

	assert.NoError(t, err)
	assert.Equal(t, 15, fetched)
	assert.Equal(t, int64(15), expected)
	assert.True(t, verified)
	assert.Empty(t, failedPages)
	assert.ElementsMatch(t, []int{0, 10}, fetcher.Offsets())
}

func TestFetchSuiteStreaming_UnknownTotalStopsOnConsecutiveEmptyPages(t *testing.T) {
	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       2,
		PageSize:                 10,
		Timeout:                  5 * time.Second,
		RequestsPerMinute:        180,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 1,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(pc.config.RequestsPerMinute)

	fetcher := &unknownTotalEmptyPagesFetcher{cases: makeCases(10, 1)}

	agg := NewResultAggregator(20)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex

	fetched, expected, verified, err := pc.fetchSuiteStreaming(
		context.Background(),
		SuiteTask{SuiteID: 1, ProjectID: 10, EstimatedSize: 10},
		fetcher,
		agg,
		&failedPagesMu,
		&failedPages,
	)

	assert.NoError(t, err)
	assert.Equal(t, 10, fetched)
	assert.Equal(t, int64(-1), expected)
	assert.True(t, verified)
	assert.Empty(t, failedPages)
}

func TestFetchSuiteStreaming_PartialRetryRecoveredInPhase3(t *testing.T) {
	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       2,
		PageSize:                 10,
		Timeout:                  5 * time.Second,
		RequestsPerMinute:        180,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 2,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(pc.config.RequestsPerMinute)

	fetcher := &transientPageErrorFetcher{
		failOnceByOff: map[int]bool{20: true},
		cases:         makeCases(25, 1),
	}

	agg := NewResultAggregator(40)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex

	fetched, expected, verified, err := pc.fetchSuiteStreaming(
		context.Background(),
		SuiteTask{SuiteID: 1, ProjectID: 10, EstimatedSize: 25},
		fetcher,
		agg,
		&failedPagesMu,
		&failedPages,
	)

	assert.NoError(t, err)
	assert.Equal(t, 25, fetched)
	assert.Equal(t, int64(25), expected)
	assert.True(t, verified)
	assert.Empty(t, failedPages)
}

func TestFetchSuiteStreaming_ProbeFailThenParallelDiscoversKnownTotal(t *testing.T) {
	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       2,
		PageSize:                 10,
		Timeout:                  5 * time.Second,
		RequestsPerMinute:        180,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 2,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(pc.config.RequestsPerMinute)

	fetcher := &probeFailsThenKnownTotalFetcher{
		total: 5,
		cases: makeCases(5, 1),
	}

	agg := NewResultAggregator(20)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex

	fetched, expected, verified, err := pc.fetchSuiteStreaming(
		context.Background(),
		SuiteTask{SuiteID: 1, ProjectID: 10, EstimatedSize: 5},
		fetcher,
		agg,
		&failedPagesMu,
		&failedPages,
	)

	assert.NoError(t, err)
	assert.Equal(t, 5, fetched)
	assert.Equal(t, int64(5), expected)
	assert.True(t, verified)
	assert.Empty(t, failedPages)
	assert.Contains(t, fetcher.Offsets(), 0)
}
