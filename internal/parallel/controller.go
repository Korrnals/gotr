// Package parallel provides the ParallelController for orchestrating concurrent API requests.
package parallel

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Korrnals/gotr/internal/concurrent"
	"github.com/Korrnals/gotr/internal/log"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/progress"
	"golang.org/x/sync/errgroup"
)

// ParallelController orchestrates parallel fetching of cases across multiple suites
type ParallelController struct {
	config      *ControllerConfig
	workerPool  *concurrent.WorkerPool
	limiter     *concurrent.AdaptiveRateLimiter
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

// Execute executes parallel fetching for the given suite tasks
func (pc *ParallelController) Execute(
	ctx context.Context,
	tasks []SuiteTask,
	fetcher SuiteFetcher,
	progressMonitor *progress.Monitor,
) (*ExecutionResult, error) {
	if len(tasks) == 0 {
		return &ExecutionResult{
			Cases:  []data.Case{},
			Errors: []error{},
			Stats:  AggregationStats{},
		}, nil
	}

	// Get total expected count first (for verification)
	totalExpected := 0
	for _, task := range tasks {
		if task.EstimatedSize > 0 {
			totalExpected += task.EstimatedSize
		}
	}
	log.Debug(fmt.Sprintf("[ParallelController] Starting execution for %d suites, estimated total: %d cases", 
		len(tasks), totalExpected))

	// Apply timeout if configured
	if pc.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, pc.config.Timeout)
		defer cancel()
		log.Debug(fmt.Sprintf("[ParallelController] Timeout set to %v", pc.config.Timeout))
	}

	// Initialize rate limiter
	pc.limiter = concurrent.NewAdaptiveRateLimiter(pc.config.RequestsPerMinute)

	// Create priority queue and populate it
	pq := NewPriorityQueue()
	for _, task := range tasks {
		// Estimate size if not provided
		if task.EstimatedSize == 0 {
			task.EstimatedSize = pc.estimateSize(ctx, task, fetcher)
		}
		pq.Push(task)
	}
	pq.Close() // Close queue, no more items will be added

	// Create result aggregator
	aggregator := NewResultAggregator(len(tasks) * 10)
	aggregator.Start(ctx)

	// Track statistics
	startTime := time.Now()
	var completedSuites int32
	totalSuites := int32(len(tasks))

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
			return pc.suiteWorker(ctx, pq, fetcher, aggregator, &completedSuites, totalSuites, progressMonitor)
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

	// Verify completeness
	log.Debug(fmt.Sprintf("[ParallelController] Execution complete: got %d cases from %d suites (expected ~%d)",
		len(cases), completedSuites, totalExpected))
	
	if len(errors) > 0 {
		log.Debug(fmt.Sprintf("[ParallelController] Errors encountered: %d", len(errors)))
	}

	// Check for significant data loss (>10%)
	if totalExpected > 0 && len(cases) < int(float64(totalExpected)*0.9) {
		log.Debug(fmt.Sprintf("[ParallelController] WARNING: Significant data loss detected! Expected ~%d, got %d (%.1f%%)",
			totalExpected, len(cases), float64(len(cases))/float64(totalExpected)*100))
	}

	result := &ExecutionResult{
		Cases:         cases,
		Errors:        errors,
		Stats:         stats,
		Partial:       err != nil && len(cases) > 0,
		ExpectedCases: totalExpected,
	}

	if err != nil && len(cases) == 0 {
		return result, fmt.Errorf("parallel execution failed: %w", err)
	}

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
	progressMonitor *progress.Monitor,
) error {
	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Get next task from queue
		task, ok := pq.Pop()
		if !ok {
			// Queue is closed and empty
			return nil
		}

		log.Debug(fmt.Sprintf("Processing suite %d (estimated size: %d)", task.SuiteID, task.EstimatedSize))

		// Fetch all pages for this suite
		err := pc.fetchSuiteParallel(ctx, task, fetcher, aggregator)
		if err != nil {
			log.Debug(fmt.Sprintf("Error fetching suite %d: %v", task.SuiteID, err))
			aggregator.SubmitError(fmt.Errorf("suite %d: %w", task.SuiteID, err))
		}

		// Update progress
		completed := atomic.AddInt32(completedSuites, 1)
		if progressMonitor != nil {
			progressMonitor.Increment()
		}

		log.Debug(fmt.Sprintf("Completed suite %d (%d/%d)", task.SuiteID, completed, totalSuites))
	}
}

// fetchSuiteParallel fetches all pages from a suite in parallel
func (pc *ParallelController) fetchSuiteParallel(
	ctx context.Context,
	task SuiteTask,
	fetcher SuiteFetcher,
	aggregator *ResultAggregator,
) error {
	// Get total count to calculate pages
	totalCount, err := fetcher.GetTotalCases(ctx, task.ProjectID, task.SuiteID)
	if err != nil {
		return fmt.Errorf("failed to get total count for suite %d: %w", task.SuiteID, err)
	}

	if totalCount == 0 {
		log.Debug(fmt.Sprintf("[Suite %d] No cases to fetch", task.SuiteID))
		return nil // Nothing to fetch
	}

	// Calculate number of pages
	pageSize := pc.config.PageSize
	numPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	
	log.Debug(fmt.Sprintf("[Suite %d] Expecting %d cases, %d pages (pageSize=%d)", 
		task.SuiteID, totalCount, numPages, pageSize))

	// Create page requests
	pages := make([]PageRequest, numPages)
	for i := 0; i < numPages; i++ {
		pages[i] = PageRequest{
			SuiteTask: task,
			Offset:    i * pageSize,
			Limit:     pageSize,
			PageNum:   i + 1,
		}
	}

	// Fetch pages in parallel with limited concurrency
	maxPageWorkers := pc.config.MaxConcurrentPages
	if maxPageWorkers > numPages {
		maxPageWorkers = numPages
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxPageWorkers)

	var successCount int32
	var errorCount int32
	var mu sync.Mutex
	var firstError error

	for _, page := range pages {
		page := page // capture loop variable
		g.Go(func() error {
			result := pc.fetchPageWithRetry(ctx, page, fetcher)

			// Submit result to aggregator
			aggregator.Submit(result)

			if result.Error != nil {
				atomic.AddInt32(&errorCount, 1)
				mu.Lock()
				if firstError == nil {
					firstError = result.Error
				}
				mu.Unlock()
				return nil // Don't stop on individual page errors
			}

			atomic.AddInt32(&successCount, 1)
			return nil
		})
	}

	// Wait for all page fetches
	if err := g.Wait(); err != nil {
		return fmt.Errorf("suite %d: %w", task.SuiteID, err)
	}

	// Log completion status
	success := atomic.LoadInt32(&successCount)
	errors := atomic.LoadInt32(&errorCount)
	
	if errors > 0 {
		log.Debug(fmt.Sprintf("[Suite %d] Completed with errors: %d/%d pages succeeded, %d failed", 
			task.SuiteID, success, numPages, errors))
	} else {
		log.Debug(fmt.Sprintf("[Suite %d] All %d pages fetched successfully", task.SuiteID, numPages))
	}

	// Check if any pages succeeded
	if success == 0 && firstError != nil {
		return fmt.Errorf("suite %d: all %d page fetches failed: %w", task.SuiteID, numPages, firstError)
	}

	// Warn if significant portion failed (>20%)
	if float64(errors)/float64(numPages) > 0.2 {
		log.Debug(fmt.Sprintf("[Suite %d] WARNING: %.0f%% of pages failed (%d/%d)", 
			task.SuiteID, float64(errors)/float64(numPages)*100, errors, numPages))
	}

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

	// Apply rate limiting
	pc.limiter.Wait()

	// Fetch with retry
	retryConfig := concurrent.DefaultRetryConfig()
	err := concurrent.RetryWithContext(ctx, retryConfig, func() error {
		cases, err := fetcher.FetchPage(ctx, req)
		if err != nil {
			return err
		}
		result.Cases = cases
		return nil
	})

	result.Duration = time.Since(start)

	if err != nil {
		result.Error = fmt.Errorf("page %d (offset %d): %w", req.PageNum, req.Offset, err)
		// Record error for adaptive rate limiting
		pc.limiter.RecordResponseTime(result.Duration)
	} else {
		// Record success for adaptive rate limiting
		pc.limiter.RecordResponseTime(result.Duration)
	}

	return result
}

// estimateSize estimates the number of cases in a suite
func (pc *ParallelController) estimateSize(
	ctx context.Context,
	task SuiteTask,
	fetcher SuiteFetcher,
) int {
	// Try to get total count
	count, err := fetcher.GetTotalCases(ctx, task.ProjectID, task.SuiteID)
	if err == nil {
		return count
	}

	// Fallback: try fetching first page to estimate
	req := PageRequest{
		SuiteTask: task,
		Offset:    0,
		Limit:     1, // Just get count
	}

	cases, err := fetcher.FetchPage(ctx, req)
	if err != nil {
		return 100 // Default estimate
	}

	// If we got cases, use a conservative estimate
	if len(cases) > 0 {
		return len(cases) * 10 // Assume 10 pages
	}

	return 100 // Default
}

// GetStats returns statistics about the current execution
// Note: This is a placeholder as stats are returned in ExecutionResult
func (pc *ParallelController) GetStats() AggregationStats {
	return AggregationStats{}
}
