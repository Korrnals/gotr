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

	// Apply timeout if configured
	if pc.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, pc.config.Timeout)
		defer cancel()
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

	result := &ExecutionResult{
		Cases:   cases,
		Errors:  errors,
		Stats:   stats,
		Partial: err != nil && len(cases) > 0,
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
		return fmt.Errorf("failed to get total count: %w", err)
	}

	if totalCount == 0 {
		return nil // Nothing to fetch
	}

	// Calculate number of pages
	pageSize := pc.config.PageSize
	numPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

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
	var mu sync.Mutex
	var firstError error

	for _, page := range pages {
		page := page // capture loop variable
		g.Go(func() error {
			result := pc.fetchPageWithRetry(ctx, page, fetcher)

			// Submit result to aggregator
			aggregator.Submit(result)

			if result.Error != nil {
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
		return err
	}

	// Check if any pages succeeded
	if successCount == 0 && firstError != nil {
		return fmt.Errorf("all page fetches failed: %w", firstError)
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
