// Package parallel provides recursive parallelization for TestRail API requests.
//
// It enables concurrent fetching of test cases across multiple test suites,
// with parallel pagination within each suite. The package uses priority queues
// to prioritize large suites, adaptive rate limiting to respect API limits,
// and provides comprehensive error handling and progress tracking.
//
// Main Components:
//
//   - ParallelController: Orchestrates parallel fetching across multiple suites
//   - PriorityQueue: Thread-safe priority queue for suite task scheduling
//   - ResultAggregator: Collects and deduplicates results from parallel workers
//
// Usage Example:
//
//	controller := parallel.NewController(&parallel.ControllerConfig{
//	    MaxConcurrentSuites: 5,
//	    MaxConcurrentPages:  3,
//	    Timeout:             5 * time.Minute,
//	})
//
//	tasks := []parallel.SuiteTask{
//	    {SuiteID: 1, ProjectID: 10},
//	    {SuiteID: 2, ProjectID: 10},
//	}
//
//	result, err := controller.Execute(ctx, tasks, fetcher, progressMonitor)
//
// Architecture:
//
// The package implements a two-level parallelization strategy:
//
// Level 1 - Suite Parallelism:
//   - Multiple suites are processed concurrently (up to MaxConcurrentSuites)
//   - Priority queue ensures large suites (>1000 cases) are processed first
//
// Level 2 - Page Parallelism:
//   - Within each suite, pages are fetched in parallel (up to MaxConcurrentPages)
//   - Adaptive rate limiting prevents API throttling
//
// Error Handling:
//   - Individual page failures don't stop the entire operation
//   - Partial results are returned with error information
//   - Timeout support with graceful shutdown
//
// Performance:
//   - Default configuration targets 60-80% improvement over sequential fetching
//   - Target: < 5 minutes for 36,000+ cases (from ~12 minutes sequential)
package parallel
