// Package concurrency provides generic parallel fetch strategies for API resources.
package concurrency

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

// FetchOption configures optional behavior for FetchParallel and FetchParallelBySuite.
type FetchOption func(*fetchOptions)

type fetchOptions struct {
	reporter       ProgressReporter
	continueOnErr  bool
	maxConcurrency int
}

// isCancellationError checks context cancellation across wrapped and string-only forms.
// Some network stacks return text errors that do not preserve sentinel wrapping.
func isCancellationError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "context canceled") || strings.Contains(msg, "deadline exceeded")
}

func defaultFetchOptions() *fetchOptions {
	return &fetchOptions{
		maxConcurrency: 0, // unlimited (bounded only by len(projectIDs))
	}
}

// WithReporter sets an optional progress reporter.
func WithReporter(r ProgressReporter) FetchOption {
	return func(o *fetchOptions) {
		o.reporter = r
	}
}

// WithContinueOnError allows partial results — errors are collected but don't stop other fetches.
func WithContinueOnError() FetchOption {
	return func(o *fetchOptions) {
		o.continueOnErr = true
	}
}

// WithMaxConcurrency limits the number of concurrent fetches.
// Default 0 = no limit (all projects fetched concurrently).
func WithMaxConcurrency(n int) FetchOption {
	return func(o *fetchOptions) {
		o.maxConcurrency = n
	}
}

// FetchParallel loads a resource from multiple projects in parallel.
// Uses errgroup for goroutine management.
//
// Type parameter T can be any type (e.g., data.Group, data.Label, ItemInfo).
// The fetchFn is called once per projectID; results are collected into a map.
//
// Example:
//
//	results, err := concurrency.FetchParallel(ctx, []int64{30, 34},
//	    func(pid int64) ([]ItemInfo, error) {
//	        return fetchGroupItems(cli, pid)
//	    },
//	    concurrency.WithReporter(task),
//	)
func FetchParallel[T any](
	ctx context.Context,
	projectIDs []int64,
	fetchFn func(projectID int64) ([]T, error),
	opts ...FetchOption,
) (map[int64][]T, error) {
	if len(projectIDs) == 0 {
		return make(map[int64][]T), nil
	}

	options := defaultFetchOptions()
	for _, opt := range opts {
		opt(options)
	}

	results := make(map[int64][]T, len(projectIDs))
	var mu sync.Mutex
	var collectedErrors []error
	var errMu sync.Mutex

	concurrency := len(projectIDs)
	if options.maxConcurrency > 0 && options.maxConcurrency < concurrency {
		concurrency = options.maxConcurrency
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrency)

	for _, pid := range projectIDs {
		pid := pid // capture loop variable
		g.Go(func() error {
			items, err := fetchFn(pid)
			if err != nil {
				if isCancellationError(err) || ctx.Err() != nil {
					return err
				}

				if options.reporter != nil {
					options.reporter.OnError()
				}

				if options.continueOnErr {
					errMu.Lock()
					collectedErrors = append(collectedErrors, fmt.Errorf("project %d: %w", pid, err))
					errMu.Unlock()
					return nil
				}
				return fmt.Errorf("project %d: %w", pid, err)
			}

			mu.Lock()
			results[pid] = items
			mu.Unlock()

			if options.reporter != nil {
				options.reporter.OnItemComplete()
				options.reporter.OnBatchReceived(len(items))
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return results, err
	}

	if len(collectedErrors) > 0 {
		return results, fmt.Errorf("%d errors during parallel fetch: %v", len(collectedErrors), collectedErrors[0])
	}

	return results, nil
}
