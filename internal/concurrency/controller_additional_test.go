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
