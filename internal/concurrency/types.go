// Package concurrency provides recursive parallelization for API requests.
// It enables concurrent fetching across multiple suites with parallel pagination
// within each suite, using priority queues and adaptive rate limiting.
package concurrency

import (
	"context"
	"time"

	"github.com/Korrnals/gotr/internal/models/data"
)

// Priority represents task priority level
type Priority int

const (
	// PriorityLow for small suites (< 100 cases)
	PriorityLow Priority = 1
	// PriorityMedium for medium suites (100-1000 cases)
	PriorityMedium Priority = 2
	// PriorityHigh for large suites (> 1000 cases)
	PriorityHigh Priority = 3

	// priorityThresholdHigh is a read-only threshold for high-priority suites.
	priorityThresholdHigh = 1000
	// priorityThresholdMedium is a read-only threshold for medium-priority suites.
	priorityThresholdMedium = 100
)

// SuiteTask represents a unit of work for fetching cases from a suite
type SuiteTask struct {
	SuiteID       int64
	ProjectID     int64
	EstimatedSize int // Estimated number of cases (for prioritization)
}

// GetPriority returns the priority based on estimated size
func (st SuiteTask) GetPriority() Priority {
	switch {
	case st.EstimatedSize >= priorityThresholdHigh:
		return PriorityHigh
	case st.EstimatedSize >= priorityThresholdMedium:
		return PriorityMedium
	default:
		return PriorityLow
	}
}

// PageRequest represents a request to fetch a specific page of cases
type PageRequest struct {
	SuiteTask
	Offset  int
	Limit   int
	PageNum int // For debugging/metrics
}

// PageResult represents the result of fetching a page
type PageResult struct {
	SuiteID   int64
	Offset    int
	Limit     int
	Cases     []data.Case
	Error     error
	Duration  time.Duration
	FetchedAt time.Time
	TotalSize int64 // Total number of cases for the suite (from API "size" field, -1 = unknown)
}

// FailedPage describes a page that could not be fetched even after recovery attempts.
type FailedPage struct {
	ProjectID int64  `json:"project_id" yaml:"project_id"`
	SuiteID   int64  `json:"suite_id" yaml:"suite_id"`
	Offset    int    `json:"offset" yaml:"offset"`
	Limit     int    `json:"limit" yaml:"limit"`
	PageNum   int    `json:"page_num" yaml:"page_num"`
	Error     string `json:"error" yaml:"error"`
}

// IsSuccess returns true if the page was fetched successfully
func (pr PageResult) IsSuccess() bool {
	return pr.Error == nil
}

// ExecutionResult contains the final result of parallel execution
type ExecutionResult struct {
	Cases       []data.Case
	Errors      []error
	FailedPages []FailedPage
	Stats       AggregationStats
	Partial     bool // true if execution was interrupted
}

// Fetcher is the function type for fetching cases
type Fetcher func(ctx context.Context, projectID int64, suiteID int64, offset int, limit int) ([]data.Case, error)

// SuiteFetcher is the interface for fetching suite data
type SuiteFetcher interface {
	// FetchPageCtx fetches a single page of cases.
	// Returns (cases, totalSize, error). totalSize is the total count from API ("size" field); -1 if unknown.
	FetchPageCtx(ctx context.Context, req PageRequest) ([]data.Case, int64, error)
}

// ProgressReporter receives progress updates from concurrent operations.
// All methods must be thread-safe.
//
// Used by FetchParallel[T] and FetchParallelBySuite[T] strategies.
// The ui.Task type implements this interface.
type ProgressReporter interface {
	OnItemComplete()       // one unit of work completed (project, suite, etc.)
	OnBatchReceived(n int) // a batch of N items received
	OnError()              // an error occurred
}

// PaginatedProgressReporter extends ProgressReporter for paginated strategies.
// Used by ParallelController (heavy strategy for cases).
//
// The ui.Task type implements this interface.
type PaginatedProgressReporter interface {
	ProgressReporter
	OnPageFetched() // one page of paginated results fetched
}

// ControllerConfig configures the ParallelController
type ControllerConfig struct {
	// MaxConcurrentSuites limits parallel suite processing (default: 5)
	MaxConcurrentSuites int
	// MaxConcurrentPages limits parallel page fetching per suite (default: 3)
	MaxConcurrentPages int
	// RequestsPerMinute is the API rate limit (default: 180).
	// Set to 0 to disable rate limiting (recommended for TestRail Server).
	RequestsPerMinute int
	// Timeout is the total operation timeout (default: 5m)
	Timeout time.Duration
	// PageSize is the number of items per page request (default: 250)
	PageSize int
	// Reporter receives fine-grained progress updates (optional).
	// Typically a *ui.Task that implements PaginatedProgressReporter.
	Reporter PaginatedProgressReporter
	// MaxRetries is the number of retries per page request (default: 5).
	// Set to 0 to disable retries.
	MaxRetries int
	// MaxConsecutiveErrorWaves: stop streaming after N consecutive waves
	// where ALL pages errored (no data). Default: 3.
	MaxConsecutiveErrorWaves int
}

// DefaultControllerConfig returns a default configuration
func DefaultControllerConfig() *ControllerConfig {
	return &ControllerConfig{
		MaxConcurrentSuites:      5,
		MaxConcurrentPages:       3,
		RequestsPerMinute:        180,
		Timeout:                  5 * time.Minute,
		PageSize:                 250,
		MaxRetries:               5,
		MaxConsecutiveErrorWaves: 3,
	}
}

// WithMaxConcurrentSuites sets the max concurrent suites
func (c *ControllerConfig) WithMaxConcurrentSuites(n int) *ControllerConfig {
	c.MaxConcurrentSuites = n
	return c
}

// WithMaxConcurrentPages sets the max concurrent pages
func (c *ControllerConfig) WithMaxConcurrentPages(n int) *ControllerConfig {
	c.MaxConcurrentPages = n
	return c
}

// WithTimeout sets the timeout
func (c *ControllerConfig) WithTimeout(d time.Duration) *ControllerConfig {
	c.Timeout = d
	return c
}

// Validate validates the configuration
func (c *ControllerConfig) Validate() error {
	if c.MaxConcurrentSuites <= 0 {
		c.MaxConcurrentSuites = 5
	}
	if c.MaxConcurrentPages <= 0 {
		c.MaxConcurrentPages = 3
	}
	if c.RequestsPerMinute < 0 {
		c.RequestsPerMinute = 180
	}
	// RequestsPerMinute == 0 is valid — means unlimited (no rate limiting)
	if c.Timeout <= 0 {
		c.Timeout = 5 * time.Minute
	}
	if c.PageSize <= 0 {
		c.PageSize = 250
	}
	if c.MaxConsecutiveErrorWaves <= 0 {
		c.MaxConsecutiveErrorWaves = 3
	}
	return nil
}
