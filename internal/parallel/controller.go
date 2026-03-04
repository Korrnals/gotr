// Package parallel provides the ParallelController for orchestrating concurrent API requests.
package parallel

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
			return pc.suiteWorker(ctx, pq, fetcher, aggregator, &completedSuites, totalSuites, &failedPagesMu, &failedPages)
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
		err := pc.fetchSuiteStreaming(ctx, task, fetcher, aggregator, failedPagesMu, failedPages)
		if err != nil {
			log.Debug(fmt.Sprintf("Error fetching suite %d: %v", task.SuiteID, err))
			aggregator.SubmitError(fmt.Errorf("suite %d: %w", task.SuiteID, err))
		}

		completed := atomic.AddInt32(completedSuites, 1)

		// Report suite completion via Reporter
		if pc.config.Reporter != nil {
			pc.config.Reporter.OnSuiteComplete()
		}

		log.Debug(fmt.Sprintf("Completed suite %d (%d/%d)", task.SuiteID, completed, totalSuites))
	}
}

// fetchSuiteStreaming fetches all pages from a suite using pipeline-based parallel pagination.
// Instead of wave-based synchronization (fetch N → wait all → fetch N), this uses
// N independent workers that continuously claim the next offset via an atomic counter.
//
// Algorithm:
//   - N workers (MaxConcurrentPages) run in parallel
//   - Each worker atomically claims the next offset, fetches the page, submits results
//   - When a worker receives an empty page (0 cases, no error) → sets exhausted flag
//   - All workers check the exhausted flag before claiming the next offset
//   - Failed offsets are collected for a sequential recovery pass
//
// This eliminates wave synchronization overhead: no worker waits for the slowest
// page in a batch. Throughput is limited only by the rate limiter and network latency.
func (pc *ParallelController) fetchSuiteStreaming(
	ctx context.Context,
	task SuiteTask,
	fetcher SuiteFetcher,
	aggregator *ResultAggregator,
	failedPagesMu *sync.Mutex,
	failedPages *[]FailedPage,
) error {
	pageSize := pc.config.PageSize
	numWorkers := pc.config.MaxConcurrentPages
	if numWorkers <= 0 {
		numWorkers = 3
	}

	reporter := pc.config.Reporter

	// Atomic offset counter — each worker claims the next offset atomically
	var nextOffset int64
	// Exhausted flag — set when we confidently reached end of data
	var exhausted int32
	// Known total size from API "size" field (-1 = unknown yet)
	var knownTotal int64 = -1
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

	// Stats
	var totalCases int32

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

				// Use TotalSize from API to set exact bound (first response wins)
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
					reporter.OnCasesReceived(len(result.Cases))
					reporter.OnPageFetched()
				}
			}
		})
	}

	g.Wait()

	// === Recovery pass: re-fetch failed pages sequentially ===
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
					reporter.OnCasesReceived(len(result.Cases))
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

	cases := atomic.LoadInt32(&totalCases)
	log.Debug(fmt.Sprintf("[Suite %d] Pipeline complete: %d cases fetched", task.SuiteID, cases))

	return nil
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
