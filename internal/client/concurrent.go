package client

import (
	"context"
	"fmt"
	"sync"

	"github.com/Korrnals/gotr/internal/concurrent"
	"github.com/Korrnals/gotr/internal/models/data"
)

// defaultWorkers is the default number of parallel workers.
const defaultWorkers = 5

// GetCasesParallel fetches cases from multiple suites in parallel.
// Uses WorkerPool for concurrency limits and RateLimiter
// to respect API rate limits (150 req/min).
//
// Parameters:
//   - projectID: project ID
//   - suiteIDs: list of suite IDs to load
//   - workers: number of parallel workers (0 = defaultWorkers)
//   - monitor: optional progress monitor (may be nil)
//
// Returns:
//   - map[suiteID] => list of cases
//   - error if at least one request failed
//
// Example without progress:
//
//	cases, err := client.GetCasesParallel(30, []int64{1, 2, 3}, 5, nil)
//	if err != nil {
//	    log.Printf("Some suites failed: %v", err)
//	}
//
// Example with progress bar:
//
//	progressChan := make(chan int, 100)
//	monitor := progress.NewMonitor(progressChan, len(suiteIDs))
//	go func() {
//	    for range progressChan {
//	        bar.Add(1)
//	    }
//	}()
//	cases, err := client.GetCasesParallel(30, suiteIDs, 5, monitor)
func (c *HTTPClient) GetCasesParallel(ctx context.Context, projectID int64, suiteIDs []int64, workers int, monitor ProgressMonitor) (map[int64]data.GetCasesResponse, error) {
	if len(suiteIDs) == 0 {
		return make(map[int64]data.GetCasesResponse), nil
	}

	if workers <= 0 {
		workers = defaultWorkers
	}

	// Results
	results := make(map[int64]data.GetCasesResponse, len(suiteIDs))
	var mu sync.Mutex

	// Errors
	var errs []error
	var errMu sync.Mutex

	// Worker pool with concurrency limit, rate limiter, and progress monitor
	opts := []concurrent.PoolOption{
		concurrent.WithMaxWorkers(workers),
		concurrent.WithRateLimit(180),
	}
	if monitor != nil {
		opts = append(opts, concurrent.WithProgressMonitor(monitor))
	}
	pool := concurrent.NewWorkerPool(opts...)

	// Submit tasks
	for _, suiteID := range suiteIDs {
		sid := suiteID // capture loop variable
		pool.Submit(func() error {
			// Fetch cases (without internal progress, pool handles that)
			cases, err := c.GetCases(ctx, projectID, sid, 0)
			if err != nil {
				errMu.Lock()
				errs = append(errs, fmt.Errorf("suite %d: %w", sid, err))
				errMu.Unlock()
				return err
			}

			// Store result
			mu.Lock()
			results[sid] = cases
			mu.Unlock()

			return nil
		})
	}

	// Wait for all tasks to complete
	if err := pool.Wait(); err != nil {
		return results, fmt.Errorf("parallel execution failed: %w", err)
	}

	return results, nil
}

// GetSuitesParallel fetches suites from multiple projects in parallel.
// Useful for compare-all commands that need suites from two projects.
//
// Parameters:
//   - projectIDs: list of project IDs
//   - workers: number of parallel workers (0 = defaultWorkers)
//   - monitor: optional progress monitor (may be nil)
//
// Returns:
//   - map[projectID] => list of suites
//   - error if at least one request failed
func (c *HTTPClient) GetSuitesParallel(ctx context.Context, projectIDs []int64, workers int, monitor ProgressMonitor) (map[int64]data.GetSuitesResponse, error) {
	if len(projectIDs) == 0 {
		return make(map[int64]data.GetSuitesResponse), nil
	}

	if workers <= 0 {
		workers = defaultWorkers
	}

	// Results
	results := make(map[int64]data.GetSuitesResponse, len(projectIDs))
	var mu sync.Mutex

	// Errors
	var errs []error
	var errMu sync.Mutex

	// Worker pool with optional progress monitor
	opts := []concurrent.PoolOption{
		concurrent.WithMaxWorkers(workers),
		concurrent.WithRateLimit(180),
	}
	if monitor != nil {
		opts = append(opts, concurrent.WithProgressMonitor(monitor))
	}
	pool := concurrent.NewWorkerPool(opts...)

	// Submit tasks
	for _, projectID := range projectIDs {
		pid := projectID // capture loop variable
		pool.Submit(func() error {
			// Fetch suites
			suites, err := c.GetSuites(ctx, pid)
			if err != nil {
				errMu.Lock()
				errs = append(errs, fmt.Errorf("project %d: %w", pid, err))
				errMu.Unlock()
				return err
			}

			// Store result
			mu.Lock()
			results[pid] = suites
			mu.Unlock()

			return nil
		})
	}

	// Wait for completion
	if err := pool.Wait(); err != nil {
		return results, fmt.Errorf("parallel execution failed: %w", err)
	}

	return results, nil
}

// GetCasesForSuitesParallel fetches all cases for a list of suites within one project.
// Merges results into a flat list of cases.
//
// Parameters:
//   - projectID: project ID
//   - suiteIDs: list of suite IDs
//   - workers: number of parallel workers
//   - monitor: optional progress monitor (may be nil)
//
// Returns:
//   - flat list of all cases across all suites
//   - error if at least one request failed
func (c *HTTPClient) GetCasesForSuitesParallel(ctx context.Context, projectID int64, suiteIDs []int64, workers int, monitor ProgressMonitor) (data.GetCasesResponse, error) {
	if len(suiteIDs) == 0 {
		return data.GetCasesResponse{}, nil
	}

	// Fetch cases in parallel
	results, err := c.GetCasesParallel(ctx, projectID, suiteIDs, workers, monitor)
	if err != nil && len(results) == 0 {
		return nil, err
	}

	// Merge results into a flat list
	var allCases data.GetCasesResponse
	for _, suiteCases := range results {
		allCases = append(allCases, suiteCases...)
	}

	return allCases, err
}
