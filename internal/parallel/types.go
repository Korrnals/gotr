// Package parallel provides recursive parallelization for API requests.
// It enables concurrent fetching across multiple suites with parallel pagination
// within each suite, using priority queues and adaptive rate limiting.
package parallel

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
)

// PriorityThresholds define suite size thresholds for priority assignment
var PriorityThresholds = struct {
	High   int // Suites larger than this get High priority
	Medium int // Suites larger than this get Medium priority
}{
	High:   1000,
	Medium: 100,
}

// SuiteTask represents a unit of work for fetching cases from a suite
type SuiteTask struct {
	SuiteID       int64
	ProjectID     int64
	EstimatedSize int // Estimated number of cases (for prioritization)
}

// GetPriority returns the priority based on estimated size
func (st SuiteTask) GetPriority() Priority {
	switch {
	case st.EstimatedSize >= PriorityThresholds.High:
		return PriorityHigh
	case st.EstimatedSize >= PriorityThresholds.Medium:
		return PriorityMedium
	default:
		return PriorityLow
	}
}

// PageRequest represents a request to fetch a specific page of cases
type PageRequest struct {
	SuiteTask
	Offset int
	Limit  int
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
}

// IsSuccess returns true if the page was fetched successfully
func (pr PageResult) IsSuccess() bool {
	return pr.Error == nil
}

// ExecutionResult contains the final result of parallel execution
type ExecutionResult struct {
	Cases         []data.Case
	Errors        []error
	Stats         AggregationStats
	Partial       bool // true if execution was interrupted
	ExpectedCases int  // estimated total cases expected (for verification)
}

// Fetcher is the function type for fetching cases
type Fetcher func(ctx context.Context, projectID int64, suiteID int64, offset int, limit int) ([]data.Case, error)

// SuiteFetcher is the interface for fetching suite data
type SuiteFetcher interface {
	// FetchPage fetches a single page of cases
	FetchPage(ctx context.Context, req PageRequest) ([]data.Case, error)
	// GetTotalCases returns the total number of cases in a suite
	GetTotalCases(ctx context.Context, projectID int64, suiteID int64) (int, error)
}

// ControllerConfig configures the ParallelController
type ControllerConfig struct {
	// MaxConcurrentSuites limits parallel suite processing (default: 5)
	MaxConcurrentSuites int
	// MaxConcurrentPages limits parallel page fetching per suite (default: 3)
	MaxConcurrentPages int
	// RequestsPerMinute is the API rate limit (default: 150)
	RequestsPerMinute int
	// Timeout is the total operation timeout (default: 5m)
	Timeout time.Duration
	// PageSize is the number of items per page request (default: 250)
	PageSize int
	// PriorityThresholds override default priority thresholds
	PriorityThresholds *struct {
		High   int
		Medium int
	}
}

// DefaultControllerConfig returns a default configuration
func DefaultControllerConfig() *ControllerConfig {
	return &ControllerConfig{
		MaxConcurrentSuites: 5,
		MaxConcurrentPages:  3,
		RequestsPerMinute:   150,
		Timeout:             5 * time.Minute,
		PageSize:            250,
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
	if c.RequestsPerMinute <= 0 {
		c.RequestsPerMinute = 150
	}
	if c.Timeout <= 0 {
		c.Timeout = 5 * time.Minute
	}
	if c.PageSize <= 0 {
		c.PageSize = 250
	}
	return nil
}
