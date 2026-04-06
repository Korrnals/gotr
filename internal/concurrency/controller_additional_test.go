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

type recoveryEmptyPageFetcher struct {
	mu            sync.Mutex
	failOnceByOff map[int]bool
	cases         []data.Case
}

func (f *recoveryEmptyPageFetcher) FetchPageCtx(_ context.Context, req PageRequest) ([]data.Case, int64, error) {
	f.mu.Lock()
	if f.failOnceByOff[req.Offset] {
		delete(f.failOnceByOff, req.Offset)
		f.mu.Unlock()
		return nil, -1, errors.New("transient offset failure")
	}
	f.mu.Unlock()

	if req.Offset >= len(f.cases) {
		return []data.Case{}, -1, nil
	}
	end := req.Offset + req.Limit
	if end > len(f.cases) {
		end = len(f.cases)
	}
	return f.cases[req.Offset:end], -1, nil
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

type paginatedReporterMock struct {
	itemCount  atomic.Int32
	batchCount atomic.Int32
	errorCount atomic.Int32
	pageCount  atomic.Int32
}

func (r *paginatedReporterMock) OnItemComplete()       { r.itemCount.Add(1) }
func (r *paginatedReporterMock) OnBatchReceived(n int) { r.batchCount.Add(int32(n)) }
func (r *paginatedReporterMock) OnError()              { r.errorCount.Add(1) }
func (r *paginatedReporterMock) OnPageFetched()        { r.pageCount.Add(1) }

type cancelOnProbeFetcher struct {
	cancel context.CancelFunc
}

func (f *cancelOnProbeFetcher) FetchPageCtx(_ context.Context, req PageRequest) ([]data.Case, int64, error) {
	if req.Offset == 0 {
		f.cancel()
		return nil, -1, errors.New("probe failed with cancellation")
	}
	return nil, -1, context.Canceled
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

func TestFetchSuiteStreaming_RecoveryRetryReturnsEmptyPage(t *testing.T) {
	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       1,
		PageSize:                 10,
		Timeout:                  5 * time.Second,
		RequestsPerMinute:        180,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 2,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(pc.config.RequestsPerMinute)

	fetcher := &recoveryEmptyPageFetcher{
		failOnceByOff: map[int]bool{20: true},
		cases:         makeCases(12, 1),
	}

	agg := NewResultAggregator(40)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex

	fetched, expected, verified, err := pc.fetchSuiteStreaming(
		context.Background(),
		SuiteTask{SuiteID: 1, ProjectID: 10, EstimatedSize: 12},
		fetcher,
		agg,
		&failedPagesMu,
		&failedPages,
	)

	assert.NoError(t, err)
	assert.Equal(t, 12, fetched)
	assert.Equal(t, int64(-1), expected)
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

func TestFetchSuiteStreaming_ReporterProgressAndRecoveryError(t *testing.T) {
	reporter := &paginatedReporterMock{}
	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       2,
		PageSize:                 10,
		Timeout:                  5 * time.Second,
		RequestsPerMinute:        180,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 1,
		Reporter:                 reporter,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(pc.config.RequestsPerMinute)

	fetcher := &permanentPageErrorFetcher{
		failOffset: 20,
		cases:      makeCases(25, 1),
	}

	agg := NewResultAggregator(50)
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
	assert.Equal(t, 20, fetched)
	assert.Equal(t, int64(25), expected)
	assert.False(t, verified)
	assert.Len(t, failedPages, 1)
	assert.Equal(t, int32(1), reporter.errorCount.Load())
	assert.True(t, reporter.batchCount.Load() >= 20)
	assert.True(t, reporter.pageCount.Load() >= 2)
}

func TestFetchSuiteStreaming_DefaultWorkerAndErrorWaveFallback(t *testing.T) {
	pc := NewController(DefaultControllerConfig())
	pc.config.MaxConcurrentPages = 0
	pc.config.MaxConsecutiveErrorWaves = 0
	pc.config.PageSize = 10
	pc.config.MaxRetries = 0
	pc.limiter = concurrent.NewAdaptiveRateLimiter(pc.config.RequestsPerMinute)

	agg := NewResultAggregator(20)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex

	fetched, expected, verified, err := pc.fetchSuiteStreaming(
		context.Background(),
		SuiteTask{SuiteID: 1, ProjectID: 10},
		&alwaysFailFetcher{},
		agg,
		&failedPagesMu,
		&failedPages,
	)

	assert.NoError(t, err)
	assert.Equal(t, 0, fetched)
	assert.Equal(t, int64(-1), expected)
	assert.False(t, verified)
	assert.NotEmpty(t, failedPages)
}

func TestSuiteWorker_SubmitsStreamingErrorAndReportsCompletion(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reporter := &paginatedReporterMock{}
	pc := NewController(DefaultControllerConfig())
	pc.config.MaxRetries = 0
	pc.config.Reporter = reporter
	pc.limiter = concurrent.NewAdaptiveRateLimiter(pc.config.RequestsPerMinute)

	pq := NewPriorityQueue()
	pq.Push(SuiteTask{SuiteID: 1, ProjectID: 10, EstimatedSize: 1})
	pq.Close()

	agg := NewResultAggregator(10)
	agg.StartCtx(context.Background())

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
		&cancelOnProbeFetcher{cancel: cancel},
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
	_, errs := agg.Stop()
	assert.NotEmpty(t, errs)
	assert.Equal(t, int32(1), reporter.itemCount.Load())
}

// ── New fetcher helpers for coverage of remaining branches ─────────────────

// infiniteSequentialFetcher returns exactly 1 case per page with unknown total (-1),
// so the controller never exhausts naturally and will hit the 40 K page safety cap.
type infiniteSequentialFetcher struct{}

func (f *infiniteSequentialFetcher) FetchPageCtx(_ context.Context, req PageRequest) ([]data.Case, int64, error) {
	return []data.Case{{ID: int64(req.Offset + 1)}}, -1, nil
}

// probeFailThenEmptyFetcher fails the very first call (simulating a probe error) and
// then returns an empty page with a configurable totalSize for every subsequent call.
type probeFailThenEmptyFetcher struct {
	mu        sync.Mutex
	firstDone bool
	totalSize int64
}

func (f *probeFailThenEmptyFetcher) FetchPageCtx(_ context.Context, _ PageRequest) ([]data.Case, int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.firstDone {
		f.firstDone = true
		return nil, -1, errors.New("probe failure")
	}
	return []data.Case{}, f.totalSize, nil
}

// cancelOnCallNFetcher counts FetchPageCtx invocations and cancels a context when the
// configured call number is reached.  Offsets listed in failOffsets always return error.
type cancelOnCallNFetcher struct {
	mu          sync.Mutex
	callNum     int
	cancelOnN   int
	cancel      context.CancelFunc
	cases       []data.Case
	failOffsets map[int]bool
}

func (f *cancelOnCallNFetcher) FetchPageCtx(_ context.Context, req PageRequest) ([]data.Case, int64, error) {
	f.mu.Lock()
	f.callNum++
	n := f.callNum
	if n == f.cancelOnN && f.cancel != nil {
		f.cancel()
	}
	f.mu.Unlock()

	if f.failOffsets[req.Offset] {
		return nil, -1, errors.New("permanent failure")
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

// ── Tests ───────────────────────────────────────────────────────────────────

// TestFetchSuiteStreaming_40KPageSafetyCap exercises the "offset/pageSize > 40 000" safety
// guard inside the Phase-2 worker loop.  With pageSize=1 and an infinite data source the
// single worker claims offsets 1 … 40 001.  At offset 40 001 the guard fires, sets
// exhausted and exits; the function returns normally without error.
func TestFetchSuiteStreaming_40KPageSafetyCap(t *testing.T) {
	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       1,
		PageSize:                 1,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 1,
		RequestsPerMinute:        0, // unlimited — no rate-limiter delay
		Timeout:                  30 * time.Second,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(0)

	agg := NewResultAggregator(50000)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex

	fetched, expected, verified, err := pc.fetchSuiteStreaming(
		context.Background(),
		SuiteTask{SuiteID: 1, ProjectID: 10},
		&infiniteSequentialFetcher{},
		agg,
		&failedPagesMu,
		&failedPages,
	)

	assert.NoError(t, err)
	// probe=1 case + workers at offsets 1..40 000 = 40 001 total
	assert.Greater(t, fetched, 40000)
	assert.Equal(t, int64(-1), expected)
	assert.True(t, verified)
	assert.Empty(t, failedPages)
}

// TestFetchSuiteStreaming_CASRecheckExhaustedOnEmptySuite covers the re-check block
// inside the "if result.TotalSize >= 0" branch (lines ~383-386).
// When the probe fails and the worker's retry at offset 0 discovers TotalSize=0,
// the CAS sets knownTotal=0 and the re-check finds offset(0) >= knownTotal(0) → exhausted.
func TestFetchSuiteStreaming_CASRecheckExhaustedOnEmptySuite(t *testing.T) {
	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       1,
		PageSize:                 10,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 1,
		RequestsPerMinute:        0,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(0)

	agg := NewResultAggregator(10)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex

	// totalSize=0 → CAS(-1→0), re-check: 0 >= 0 → sets exhausted, returns nil
	fetched, expected, verified, err := pc.fetchSuiteStreaming(
		context.Background(),
		SuiteTask{SuiteID: 1, ProjectID: 10},
		&probeFailThenEmptyFetcher{totalSize: 0},
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

// TestFetchSuiteStreaming_EmptyPageKnownTotalExhausts covers the
// "empty page + knownTotal >= 0" exhaustion branch (lines ~400-403).
// The probe fails; the Phase-2 worker at offset 0 gets an empty page with TotalSize=25.
// Because offset(0) < knownTotal(25) the CAS re-check does NOT fire; instead the
// "len(cases)==0 with known total" branch sets exhausted and returns nil.
func TestFetchSuiteStreaming_EmptyPageKnownTotalExhausts(t *testing.T) {
	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       1,
		PageSize:                 10,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 1,
		RequestsPerMinute:        0,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(0)

	agg := NewResultAggregator(10)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex

	// totalSize=25 → CAS(-1→25), re-check: 0 >= 25 = false (passes);
	// then len(cases)==0 && knownTotal>=0 → exhausted
	fetched, expected, verified, err := pc.fetchSuiteStreaming(
		context.Background(),
		SuiteTask{SuiteID: 1, ProjectID: 10},
		&probeFailThenEmptyFetcher{totalSize: 25},
		agg,
		&failedPagesMu,
		&failedPages,
	)

	assert.NoError(t, err)
	assert.Equal(t, 0, fetched)
	assert.Equal(t, int64(25), expected)
	assert.True(t, verified)
	assert.Empty(t, failedPages)
}

// TestFetchSuiteStreaming_RecoveryCancellationBreaks covers the
// "if ctx.Err() != nil { break }" guard at the top of the Phase-3 recovery loop.
// Setup: probe succeeds (10 cases, total=30); Phase-2 workers permanently fail offsets
// 10 and 20; Phase-3 iteration 1 cancels the context via cancelOnCallNFetcher so that
// iteration 2 finds ctx.Err() != nil and breaks.
func TestFetchSuiteStreaming_RecoveryCancellationBreaks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       2,
		PageSize:                 10,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 2,
		RequestsPerMinute:        0,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(0)

	// Call 1 = probe (offset 0, succeeds), calls 2–3 = Phase-2 workers (both fail),
	// call 4 = Phase-3 iteration 1 (cancels ctx then returns error permanently).
	fetcher := &cancelOnCallNFetcher{
		cancelOnN:   4,
		cancel:      cancel,
		cases:       makeCases(30, 1),
		failOffsets: map[int]bool{10: true, 20: true},
	}

	agg := NewResultAggregator(50)
	agg.StartCtx(context.Background())
	defer agg.Stop()

	failedPages := make([]FailedPage, 0)
	var failedPagesMu sync.Mutex

	fetched, expected, _, err := pc.fetchSuiteStreaming(
		ctx,
		SuiteTask{SuiteID: 1, ProjectID: 10},
		fetcher,
		agg,
		&failedPagesMu,
		&failedPages,
	)

	assert.NoError(t, err)
	assert.Equal(t, 10, fetched) // only probe's 10 cases
	assert.Equal(t, int64(30), expected)
	assert.NotEmpty(t, failedPages) // Phase-3 iter 1 produced a permanent failure
}

// TestFetchSuiteStreaming_RecoveryReporterCallbacks covers the reporter callback lines
// inside the Phase-3 "recovered page with cases" branch (lines ~475-478).
// A transient failure on offset 20 causes Phase-2 to skip it; Phase-3 recovers it
// successfully.  With a reporter set, OnBatchReceived+OnPageFetched are invoked.
func TestFetchSuiteStreaming_RecoveryReporterCallbacks(t *testing.T) {
	reporter := &paginatedReporterMock{}
	pc := NewController(&ControllerConfig{
		MaxConcurrentPages:       2,
		PageSize:                 10,
		MaxRetries:               0,
		MaxConsecutiveErrorWaves: 2,
		RequestsPerMinute:        0,
		Reporter:                 reporter,
	})
	pc.limiter = concurrent.NewAdaptiveRateLimiter(0)

	// Offset 20 fails once in Phase 2; Phase-3 recovers it (5 cases).
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
		SuiteTask{SuiteID: 1, ProjectID: 10},
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
	// Phase-3 recovery invoked reporter callbacks for the 5 recovered cases
	assert.GreaterOrEqual(t, reporter.batchCount.Load(), int32(25))
	assert.GreaterOrEqual(t, reporter.pageCount.Load(), int32(1))
}
