// Package concurrency provides the ParallelController for orchestrating concurrent API requests.
package concurrency

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Korrnals/gotr/internal/concurrent"
	"github.com/Korrnals/gotr/internal/log"
	"github.com/Korrnals/gotr/internal/models/data"
	"golang.org/x/sync/errgroup"
)

// ParallelController orchestrates parallel fetching of cases across multiple suites
type ParallelController struct {
	config  *ControllerConfig
	limiter *concurrent.AdaptiveRateLimiter
}

// NewController creates a new parallel controller with the given configuration
func NewController(config *ControllerConfig) *ParallelController {
	if config == nil {
		config = DefaultControllerConfig()
	}
	config.Validate()

	return &ParallelController{
		config: config,
	}
}

// Execute executes parallel fetching for the given suite tasks.
//
// Key design decisions (Stage 6.7 v2):
//   - Streaming pagination — pages fetched in waves until data exhausted
//   - Reporter callbacks on every page — enables real-time progress display
//   - Post-factum integrity log: total cases, pages, errors, duplicates
func (pc *ParallelController) Execute(
	ctx context.Context,
	tasks []SuiteTask,
	fetcher SuiteFetcher,
	_ interface{}, // backward compat: was *progress.Monitor, now unused
) (*ExecutionResult, error) {
	if len(tasks) == 0 {
		return &ExecutionResult{
			Cases:  []data.Case{},
			Errors: []error{},
			Stats:  AggregationStats{},
		}, nil
	}

	log.Debug(fmt.Sprintf("[ParallelController] Starting execution for %d suites (streaming mode)", len(tasks)))

	// Apply timeout if configured
	if pc.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, pc.config.Timeout)
		defer cancel()
		log.Debug(fmt.Sprintf("[ParallelController] Timeout set to %v", pc.config.Timeout))
	}

	// Initialize rate limiter
	pc.limiter = concurrent.NewAdaptiveRateLimiter(pc.config.RequestsPerMinute)

	// Create priority queue — use simple FIFO
	pq := NewPriorityQueue()
	for _, task := range tasks {
		if task.EstimatedSize == 0 {
			task.EstimatedSize = 100 // default priority bucket
		}
		pq.Push(task)
	}
	pq.Close()

	// Create result aggregator
	aggregator := NewResultAggregator(len(tasks) * 10)
	aggregator.StartCtx(ctx)

	// Track statistics
	startTime := time.Now()
	var completedSuites int32
	totalSuites := int32(len(tasks))
	var failedPagesMu sync.Mutex
	failedPages := make([]FailedPage, 0)

	// Expected cases tracking (sum of API totalSize per suite)
	var expectedCasesTotal int64
	var suitesWithTotal int32
	var suitesVerified int32
	var suiteResultsMu sync.Mutex
	suiteResults := make([]SuiteResultInfo, 0, len(tasks))

	// Create worker pool for suites
	maxWorkers := pc.config.MaxConcurrentSuites
	if maxWorkers > len(tasks) {
		maxWorkers = len(tasks)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxWorkers)

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		g.Go(func() error {
			return pc.suiteWorker(ctx, pq, fetcher, aggregator, &completedSuites, totalSuites, &failedPagesMu, &failedPages, &expectedCasesTotal, &suitesWithTotal, &suitesVerified, &suiteResultsMu, &suiteResults)
		})
	}

	// Wait for all workers to complete
	err := g.Wait()

	// Get results from aggregator
	cases, errors := aggregator.Stop()

	// Build execution result
	stats := aggregator.Stats()
	stats.TotalSuites = len(tasks)
	stats.CompletedSuites = int(completedSuites)
	stats.StartTime = startTime
	stats.EndTime = time.Now()
	stats.ExpectedCases = atomic.LoadInt64(&expectedCasesTotal)
	stats.SuitesWithTotal = int(atomic.LoadInt32(&suitesWithTotal))
	stats.SuitesVerified = int(atomic.LoadInt32(&suitesVerified))
	stats.SuiteResults = suiteResults

	log.Debug(fmt.Sprintf("[ParallelController] Execution complete: got %d cases from %d suites",
		len(cases), completedSuites))

	if len(errors) > 0 {
		log.Debug(fmt.Sprintf("[ParallelController] Errors encountered: %d", len(errors)))
	}

	result := &ExecutionResult{
		Cases:       cases,
		Errors:      errors,
		FailedPages: failedPages,
		Stats:       stats,
		Partial:     err != nil && len(cases) > 0,
	}

	if err != nil && len(cases) == 0 {
		return result, fmt.Errorf("parallel execution failed: %w", err)
	}

	// Post-factum integrity log
	log.Debug(fmt.Sprintf("[ParallelController] Loaded %d cases from %d suites, %d pages, %d errors",
		len(cases), stats.CompletedSuites, stats.TotalPages, len(errors)))

	return result, nil
}

// suiteWorker processes suites from the priority queue
func (pc *ParallelController) suiteWorker(
	ctx context.Context,
	pq *PriorityQueue,
	fetcher SuiteFetcher,
	aggregator *ResultAggregator,
	completedSuites *int32,
	totalSuites int32,
	failedPagesMu *sync.Mutex,
	failedPages *[]FailedPage,
	expectedCasesTotal *int64,
	suitesWithTotal *int32,
	suitesVerified *int32,
	suiteResultsMu *sync.Mutex,
	suiteResults *[]SuiteResultInfo,
) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		task, ok := pq.Pop()
		if !ok {
			return nil
		}

		log.Debug(fmt.Sprintf("Processing suite %d", task.SuiteID))

		// Streaming fetch — no GetTotalCases needed
		casesFetched, suiteExpected, verified, err := pc.fetchSuiteStreaming(ctx, task, fetcher, aggregator, failedPagesMu, failedPages)
		if err != nil {
			log.Debug(fmt.Sprintf("Error fetching suite %d: %v", task.SuiteID, err))
			aggregator.SubmitError(fmt.Errorf("suite %d: %w", task.SuiteID, err))
		}
		if suiteExpected >= 0 {
			atomic.AddInt64(expectedCasesTotal, suiteExpected)
			atomic.AddInt32(suitesWithTotal, 1)
		}
		if verified {
			atomic.AddInt32(suitesVerified, 1)
		}

		// Track per-suite result for integrity verification
		suiteResultsMu.Lock()
		*suiteResults = append(*suiteResults, SuiteResultInfo{
			SuiteID:      int(task.SuiteID),
			CasesFetched: casesFetched,
			Verified:     verified,
		})
		suiteResultsMu.Unlock()

		completed := atomic.AddInt32(completedSuites, 1)

		// Report suite completion via Reporter
		if pc.config.Reporter != nil {
			pc.config.Reporter.OnItemComplete()
		}

		log.Debug(fmt.Sprintf("Completed suite %d (%d/%d)", task.SuiteID, completed, totalSuites))
	}
}

// fetchSuiteStreaming fetches all pages from a suite using probe-first parallel pagination.
//
// Algorithm (probe-first):
//   - Phase 1: Probe — single synchronous request for page 0 to learn totalSize
//   - Phase 2: Parallel — N workers fetch remaining pages with known bounds
//   - Phase 3: Recovery — re-fetch any failed pages sequentially
//
// The probe eliminates burst flood: instead of N workers speculatively claiming
// offsets before any API response, we learn the exact page count upfront.
// When totalSize is unknown (fallback), workers use exhaustion detection.
// fetchSuiteStreaming returns (casesFetched, expectedTotal, verified, error).
// casesFetched is the actual number of cases loaded from this suite.
// expectedTotal is the totalSize reported by the API (-1 if unknown).
// verified is true when all pages were fetched and exhaustion confirmed (no permanent errors).
func (pc *ParallelController) fetchSuiteStreaming(
	ctx context.Context,
	task SuiteTask,
	fetcher SuiteFetcher,
	aggregator *ResultAggregator,
	failedPagesMu *sync.Mutex,
	failedPages *[]FailedPage,
) (int, int64, bool, error) {
	pageSize := pc.config.PageSize
	numWorkers := pc.config.MaxConcurrentPages
	if numWorkers <= 0 {
		numWorkers = 3
	}

	reporter := pc.config.Reporter

	// Stats
	var totalCases int32

	// ── Phase 1: Probe — fetch page 0 to learn totalSize before launching workers ──
	probeReq := PageRequest{
		SuiteTask: task,
		Offset:    0,
		Limit:     pageSize,
		PageNum:   1,
	}
	probeResult := pc.fetchPageWithRetry(ctx, probeReq, fetcher)

	var knownTotal int64 = -1
	var probeFailed bool

	if probeResult.Error != nil {
		// Probe failed after retries — DON'T abandon the suite.
		// Fall back to Phase 2 starting from offset 0, giving workers
		// a chance to retry page 0 (server may recover by then).
		log.Debug(fmt.Sprintf("[Suite %d] WARNING: probe page 0 failed: %v — falling back to Phase 2",
			task.SuiteID, probeResult.Error))
		probeFailed = true
	} else {
		// Submit probe page results
		if len(probeResult.Cases) > 0 {
			aggregator.Submit(probeResult)
			atomic.AddInt32(&totalCases, int32(len(probeResult.Cases)))
			if reporter != nil {
				reporter.OnBatchReceived(len(probeResult.Cases))
				reporter.OnPageFetched()
			}
		}

		// Set known total from probe response
		if probeResult.TotalSize >= 0 {
			knownTotal = probeResult.TotalSize
		}

		// If suite is empty, fits in one page, or returned partial page (when total unknown) — done
		if len(probeResult.Cases) == 0 ||
			(knownTotal >= 0 && knownTotal <= int64(pageSize)) ||
			(knownTotal < 0 && len(probeResult.Cases) < pageSize) {
			log.Debug(fmt.Sprintf("[Suite %d] Probe: %d cases, totalSize=%d — single page, done",
				task.SuiteID, len(probeResult.Cases), knownTotal))
			return len(probeResult.Cases), knownTotal, true, nil // single page — verified
		}
	}

	// ── Phase 2: Parallel fetch of remaining pages ──

	// Cap workers to remaining pages when total is known
	if knownTotal >= 0 {
		totalPages := int((knownTotal + int64(pageSize) - 1) / int64(pageSize))
		remainingPages := totalPages - 1 // page 0 already fetched
		if remainingPages <= 0 {
			return len(probeResult.Cases), knownTotal, true, nil // known total, single page — verified
		}
		if numWorkers > remainingPages {
			numWorkers = remainingPages
		}
		log.Debug(fmt.Sprintf("[Suite %d] Probe: totalSize=%d, %d pages remaining, %d workers",
			task.SuiteID, knownTotal, remainingPages, numWorkers))
	}

	// If probe succeeded, start from page 1 (page 0 already fetched).
	// If probe failed, start from page 0 — workers will retry it.
	var nextOffset int64
	if probeFailed {
		nextOffset = 0
	} else {
		nextOffset = int64(pageSize)
	}
	// Exhausted flag — set when we confidently reached end of data
	var exhausted int32
	// Consecutive empty pages counter
	var consecutiveEmptyPages int32
	// Consecutive error counter — reset on any successful page
	var consecutiveErrors int32
	maxConsecutiveErrors := int32(pc.config.MaxConsecutiveErrorWaves * numWorkers)
	if maxConsecutiveErrors <= 0 {
		maxConsecutiveErrors = 9 // fallback
	}

	// Failed offsets for recovery pass
	var failedMu sync.Mutex
	var failedOffsets []int

	g, gctx := errgroup.WithContext(ctx)

	for w := 0; w < numWorkers; w++ {
		g.Go(func() error {
			for {
				// Check context
				if gctx.Err() != nil {
					return gctx.Err()
				}

				// Check if suite data is exhausted
				if atomic.LoadInt32(&exhausted) != 0 {
					return nil
				}

				// Check if too many consecutive errors (likely past end of data)
				if atomic.LoadInt32(&consecutiveErrors) >= maxConsecutiveErrors {
					return nil
				}

				// Claim next offset atomically
				offset := int(atomic.AddInt64(&nextOffset, int64(pageSize)) - int64(pageSize))

				// Exact bound: if API told us total size, skip offsets past it
				if kt := atomic.LoadInt64(&knownTotal); kt >= 0 && int64(offset) >= kt {
					atomic.StoreInt32(&exhausted, 1)
					return nil
				}

				// Safety: cap at 10M cases (40K pages)
				if offset/pageSize > 40000 {
					log.Debug(fmt.Sprintf("[Suite %d] WARNING: hit 40K page limit, stopping", task.SuiteID))
					atomic.StoreInt32(&exhausted, 1)
					return nil
				}

				req := PageRequest{
					SuiteTask: task,
					Offset:    offset,
					Limit:     pageSize,
					PageNum:   offset/pageSize + 1,
				}

				result := pc.fetchPageWithRetry(gctx, req, fetcher)

				// Use TotalSize from API to update bound (probe normally sets this,
				// but in fallback mode the first parallel response may set it)
				if result.TotalSize >= 0 {
					atomic.CompareAndSwapInt64(&knownTotal, -1, result.TotalSize)
					// Re-check: maybe this offset is already past the bound
					if kt := atomic.LoadInt64(&knownTotal); kt >= 0 && int64(offset) >= kt {
						atomic.StoreInt32(&exhausted, 1)
						return nil
					}
				}

				if result.Error != nil {
					// Record failed offset for recovery
					failedMu.Lock()
					failedOffsets = append(failedOffsets, offset)
					failedMu.Unlock()
					atomic.AddInt32(&consecutiveErrors, 1)
					continue
				}

				if len(result.Cases) == 0 {
					// Empty page — if we already know total, just mark exhausted
					if atomic.LoadInt64(&knownTotal) >= 0 {
						atomic.StoreInt32(&exhausted, 1)
						return nil
					}
					// Unknown total: stop after several consecutive empties
					emptyCount := atomic.AddInt32(&consecutiveEmptyPages, 1)
					if emptyCount >= int32(numWorkers) {
						atomic.StoreInt32(&exhausted, 1)
						return nil
					}
					continue
				}

				// Success — reset consecutive error counter
				atomic.StoreInt32(&consecutiveErrors, 0)
				atomic.StoreInt32(&consecutiveEmptyPages, 0)

				aggregator.Submit(result)
				atomic.AddInt32(&totalCases, int32(len(result.Cases)))

				if reporter != nil {
					reporter.OnBatchReceived(len(result.Cases))
					reporter.OnPageFetched()
				}
			}
		})
	}

	g.Wait()

	// ── Phase 3: Recovery pass — re-fetch failed pages sequentially ──
	if len(failedOffsets) > 0 {
		log.Debug(fmt.Sprintf("[Suite %d] Recovery: re-fetching %d failed pages", task.SuiteID, len(failedOffsets)))

		recovered := 0
		permanentErrors := 0
		for _, failedOffset := range failedOffsets {
			if ctx.Err() != nil {
				break
			}

			req := PageRequest{
				SuiteTask: task,
				Offset:    failedOffset,
				Limit:     pageSize,
				PageNum:   failedOffset/pageSize + 1,
			}

			result := pc.fetchPageWithRetry(ctx, req, fetcher)
			if result.Error != nil {
				// Still failed — submit as permanent error
				aggregator.Submit(result)
				failedPagesMu.Lock()
				*failedPages = append(*failedPages, FailedPage{
					ProjectID: task.ProjectID,
					SuiteID:   task.SuiteID,
					Offset:    failedOffset,
					Limit:     pageSize,
					PageNum:   failedOffset/pageSize + 1,
					Error:     result.Error.Error(),
				})
				failedPagesMu.Unlock()
				permanentErrors++
				if reporter != nil {
					reporter.OnError()
				}
				log.Debug(fmt.Sprintf("[Suite %d] Recovery FAILED for offset %d: %v",
					task.SuiteID, failedOffset, result.Error))
			} else if len(result.Cases) > 0 {
				aggregator.Submit(result)
				recovered++
				atomic.AddInt32(&totalCases, int32(len(result.Cases)))

				if reporter != nil {
					reporter.OnBatchReceived(len(result.Cases))
					reporter.OnPageFetched()
				}
			}
			// else: empty page — offset was past end of data, not an error
		}

		if recovered > 0 || permanentErrors > 0 {
			log.Debug(fmt.Sprintf("[Suite %d] Recovery: %d/%d recovered, %d permanent errors",
				task.SuiteID, recovered, len(failedOffsets), permanentErrors))
		}
	}

	// Determine verification: suite is verified if no permanent page failures
	suiteVerified := true
	if len(failedOffsets) > 0 {
		// Check if Phase 3 left permanent errors (scan shared failedPages for this suite)
		failedPagesMu.Lock()
		for _, fp := range *failedPages {
			if fp.SuiteID == task.SuiteID {
				suiteVerified = false
				break
			}
		}
		failedPagesMu.Unlock()
	}

	fetchedCount := int(atomic.LoadInt32(&totalCases))
	log.Debug(fmt.Sprintf("[Suite %d] Complete: %d cases fetched (expected %d, verified=%v)", task.SuiteID, fetchedCount, knownTotal, suiteVerified))

	return fetchedCount, knownTotal, suiteVerified, nil
}

// fetchPageWithRetry fetches a single page with rate limiting and retry
func (pc *ParallelController) fetchPageWithRetry(
	ctx context.Context,
	req PageRequest,
	fetcher SuiteFetcher,
) PageResult {
	start := time.Now()
	result := PageResult{
		SuiteID:   req.SuiteID,
		Offset:    req.Offset,
		Limit:     req.Limit,
		FetchedAt: start,
	}

	// Apply rate limiting (wait time is NOT counted in response time)
	pc.limiter.Wait()

	// Measure only the actual server response time
	fetchStart := time.Now()

	// Fetch with retry
	retryConfig := concurrent.DefaultRetryConfig()
	if pc.config.MaxRetries >= 0 {
		retryConfig.MaxRetries = pc.config.MaxRetries
	}
	result.TotalSize = -1
	err := concurrent.RetryWithContext(ctx, retryConfig, func() error {
		cases, totalSize, err := fetcher.FetchPageCtx(ctx, req)
		if err != nil {
			return err
		}
		result.Cases = cases
		result.TotalSize = totalSize
		return nil
	})

	serverTime := time.Since(fetchStart)
	result.Duration = time.Since(start) // total time including rate limiter wait

	if err != nil {
		result.Error = fmt.Errorf("page %d (offset %d): %w", req.PageNum, req.Offset, err)
	}

	// Record only server response time (not rate limiter wait time)
	pc.limiter.RecordResponseTime(serverTime)

	return result
}

// GetStats returns statistics about the current execution
func (pc *ParallelController) GetStats() AggregationStats {
	return AggregationStats{}
}
