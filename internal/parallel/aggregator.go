// Package parallel provides result aggregation for parallel operations.
package parallel

import (
	"context"
	stderrors "errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Korrnals/gotr/internal/log"
	"github.com/Korrnals/gotr/internal/models/data"
)

// ResultAggregator collects and deduplicates results from parallel fetchers
type ResultAggregator struct {
	// Configuration
	bufferSize int

	// Channels
	resultCh chan PageResult
	errCh    chan error
	doneCh   chan struct{}

	// Storage
	cases    []data.Case
	casesMu  sync.RWMutex
	seenIDs  map[int64]struct{}
	seenMu   sync.RWMutex

	// Error handling
	errors   []error
	errorsMu sync.RWMutex

	// Statistics
	totalCases  int64
	totalPages  int64
	failedPages int64
	startTime   time.Time
	endTime     time.Time

	// State
	started  bool
	stopped  bool
	mu       sync.Mutex
}

// NewResultAggregator creates a new result aggregator
func NewResultAggregator(bufferSize int) *ResultAggregator {
	if bufferSize <= 0 {
		bufferSize = 1000
	}

	return &ResultAggregator{
		bufferSize: bufferSize,
		seenIDs:    make(map[int64]struct{}),
		errors:     make([]error, 0),
	}
}

// Start begins the aggregation process
// This should be called once before submitting any results
func (ra *ResultAggregator) StartCtx(ctx context.Context) {
	ra.mu.Lock()
	defer ra.mu.Unlock()

	if ra.started {
		return
	}

	ra.started = true
	ra.startTime = time.Now()
	ra.resultCh = make(chan PageResult, ra.bufferSize)
	ra.errCh = make(chan error, 100)
	ra.doneCh = make(chan struct{})

	// Start the aggregation goroutine
	go ra.aggregateCtx(ctx)
}

// aggregate runs in a separate goroutine and processes incoming results
func (ra *ResultAggregator) aggregateCtx(ctx context.Context) {
	defer close(ra.doneCh)

	resultCh := ra.resultCh
	errCh := ra.errCh
	ctxDone := ctx.Done()

	for {
		if resultCh == nil && errCh == nil {
			return
		}

		select {
		case <-ctxDone:
			log.Debug("ResultAggregator: context cancelled, draining channels")
			ctxDone = nil

		case result, ok := <-resultCh:
			if !ok {
				resultCh = nil
				continue
			}
			ra.processResult(result)

		case err, ok := <-errCh:
			if !ok {
				errCh = nil
				continue
			}
			ra.processError(err)
		}
	}
}

// processResult handles a single PageResult
func (ra *ResultAggregator) processResult(result PageResult) {
	atomic.AddInt64(&ra.totalPages, 1)

	if result.Error != nil {
		atomic.AddInt64(&ra.failedPages, 1)
		ra.addError(result.Error)
		return
	}

	// Add cases with deduplication
	ra.addCases(result.Cases)
}

// addCases adds cases to the result slice with deduplication
func (ra *ResultAggregator) addCases(cases []data.Case) {
	if len(cases) == 0 {
		return
	}

	ra.casesMu.Lock()
	defer ra.casesMu.Unlock()

	ra.seenMu.Lock()
	defer ra.seenMu.Unlock()

	for _, c := range cases {
		// Check for duplicates using case ID
		if _, exists := ra.seenIDs[c.ID]; exists {
			continue
		}

		ra.seenIDs[c.ID] = struct{}{}
		ra.cases = append(ra.cases, c)
		atomic.AddInt64(&ra.totalCases, 1)
	}
}

// addError adds an error to the error list
func (ra *ResultAggregator) addError(err error) {
	if err == nil {
		return
	}

	ra.errorsMu.Lock()
	defer ra.errorsMu.Unlock()

	ra.errors = append(ra.errors, err)
}

// processError handles an error from the error channel
func (ra *ResultAggregator) processError(err error) {
	ra.addError(err)
}

// Submit submits a result for aggregation
// Safe to call from multiple goroutines
func (ra *ResultAggregator) Submit(result PageResult) {
	ra.mu.Lock()
	if !ra.started || ra.stopped {
		ra.mu.Unlock()
		return
	}
	ra.mu.Unlock()

	select {
	case ra.resultCh <- result:
		// Success
	default:
		// Channel full, log warning and block
		log.Debug("ResultAggregator: result channel full, blocking")
		ra.resultCh <- result
	}
}

// SubmitError submits an error for aggregation
func (ra *ResultAggregator) SubmitError(err error) {
	if err == nil {
		return
	}

	ra.mu.Lock()
	if !ra.started || ra.stopped {
		ra.mu.Unlock()
		return
	}
	ra.mu.Unlock()

	select {
	case ra.errCh <- err:
		// Success
	default:
		// Channel full, log and drop
		log.Debug("ResultAggregator: error channel full, dropping error")
	}
}

// Stop stops the aggregator and returns the final results
func (ra *ResultAggregator) Stop() ([]data.Case, []error) {
	ra.mu.Lock()
	defer ra.mu.Unlock()

	if !ra.started || ra.stopped {
		return ra.getCases(), ra.getErrors()
	}

	ra.stopped = true
	ra.endTime = time.Now()

	// Close channels to signal completion
	close(ra.resultCh)
	close(ra.errCh)

	// Wait for aggregate goroutine to finish
	<-ra.doneCh

	return ra.getCases(), ra.getErrors()
}

// getCases returns the collected cases (internal, should be called with lock held or after Stop)
func (ra *ResultAggregator) getCases() []data.Case {
	ra.casesMu.RLock()
	defer ra.casesMu.RUnlock()

	// Return a copy to avoid external modification
	result := make([]data.Case, len(ra.cases))
	copy(result, ra.cases)
	return result
}

// getErrors returns the collected errors (internal)
func (ra *ResultAggregator) getErrors() []error {
	ra.errorsMu.RLock()
	defer ra.errorsMu.RUnlock()

	result := make([]error, len(ra.errors))
	copy(result, ra.errors)
	return result
}

// GetResults returns the current results without stopping
func (ra *ResultAggregator) GetResults() ([]data.Case, []error) {
	return ra.getCases(), ra.getErrors()
}

// Stats returns current aggregation statistics
func (ra *ResultAggregator) Stats() AggregationStats {
	return AggregationStats{
		TotalCases:   int(atomic.LoadInt64(&ra.totalCases)),
		TotalPages:   int(atomic.LoadInt64(&ra.totalPages)),
		FailedPages:  int(atomic.LoadInt64(&ra.failedPages)),
		ErrorCount:   len(ra.getErrors()),
		Duration:     ra.getDuration(),
		IsRunning:    ra.IsRunning(),
	}
}

// getDuration calculates the duration since start
func (ra *ResultAggregator) getDuration() time.Duration {
	ra.mu.Lock()
	defer ra.mu.Unlock()

	if !ra.started {
		return 0
	}

	if ra.stopped {
		return ra.endTime.Sub(ra.startTime)
	}

	return time.Since(ra.startTime)
}

// IsRunning returns true if the aggregator is running
func (ra *ResultAggregator) IsRunning() bool {
	ra.mu.Lock()
	defer ra.mu.Unlock()
	return ra.started && !ra.stopped
}

// Wait waits for all results to be processed (blocks until Stop is called)
func (ra *ResultAggregator) Wait() {
	<-ra.doneCh
}

// AggregationStats contains statistics about the aggregation
type AggregationStats struct {
	TotalCases      int
	TotalPages      int
	FailedPages     int
	ErrorCount      int
	Duration        time.Duration
	IsRunning       bool
	TotalSuites     int
	CompletedSuites int
	StartTime       time.Time
	EndTime         time.Time
	ExpectedCases   int64 // sum of API totalSize across all suites (-1 per suite if unknown)
	SuitesWithTotal int   // how many suites reported totalSize
	SuitesVerified  int   // suites with all pages fetched, exhaustion confirmed, no permanent errors
	SuiteResults    []SuiteResultInfo // per-suite fetch details
}

// SuiteResultInfo tracks per-suite fetch results for integrity verification.
type SuiteResultInfo struct {
	SuiteID      int
	CasesFetched int
	Verified     bool
}

// HasErrors returns true if any errors were encountered
func (s AggregationStats) HasErrors() bool {
	return s.ErrorCount > 0 || s.FailedPages > 0
}

// ErrorRate returns the percentage of failed pages
func (s AggregationStats) ErrorRate() float64 {
	if s.TotalPages == 0 {
		return 0
	}
	return float64(s.FailedPages) / float64(s.TotalPages) * 100
}

// CombinedError returns a single error combining all errors, or nil if no errors
func CombinedError(errors []error) error {
	if len(errors) == 0 {
		return nil
	}

	if len(errors) == 1 {
		return errors[0]
	}

	// Combine errors
	msg := errors[0].Error()
	for i := 1; i < len(errors); i++ {
		msg += "; " + errors[i].Error()
	}
	return stderrors.New(msg)
}
