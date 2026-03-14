// Package concurrency provides per-suite parallel fetch strategy.
package concurrency

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

// FetchParallelBySuite loads resources across all suites of a project in parallel.
// Each suite is fetched concurrently; results are combined into a single slice.
//
// Type parameter T can be any type (e.g., data.Section).
// The fetchPerSuite function is called once per suiteID.
//
// Example:
//
//	sections, err := concurrency.FetchParallelBySuite(ctx, suiteIDs,
//	    func(suiteID int64) ([]data.Section, error) {
//	        return cli.GetSections(projectID, suiteID)
//	    },
//	    concurrency.WithReporter(task),
//	    concurrency.WithContinueOnError(),
//	)
func FetchParallelBySuite[T any](
	ctx context.Context,
	suiteIDs []int64,
	fetchPerSuite func(suiteID int64) ([]T, error),
	opts ...FetchOption,
) ([]T, error) {
	if len(suiteIDs) == 0 {
		return nil, nil
	}

	options := defaultFetchOptions()
	for _, opt := range opts {
		opt(options)
	}

	var allItems []T
	var mu sync.Mutex
	var collectedErrors []error
	var errMu sync.Mutex

	concLimit := len(suiteIDs)
	if options.maxConcurrency > 0 && options.maxConcurrency < concLimit {
		concLimit = options.maxConcurrency
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concLimit)

	for _, sid := range suiteIDs {
		sid := sid // capture
		g.Go(func() error {
			items, err := fetchPerSuite(sid)
			if err != nil {
				if isCancellationError(err) || ctx.Err() != nil {
					return err
				}

				if options.reporter != nil {
					options.reporter.OnError()
				}

				if options.continueOnErr {
					errMu.Lock()
					collectedErrors = append(collectedErrors, fmt.Errorf("suite %d: %w", sid, err))
					errMu.Unlock()
					return nil
				}
				return fmt.Errorf("suite %d: %w", sid, err)
			}

			if len(items) > 0 {
				mu.Lock()
				allItems = append(allItems, items...)
				mu.Unlock()
			}

			if options.reporter != nil {
				options.reporter.OnItemComplete()
				options.reporter.OnBatchReceived(len(items))
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return allItems, err
	}

	if len(collectedErrors) > 0 {
		return allItems, fmt.Errorf("%d errors during parallel suite fetch: %v", len(collectedErrors), collectedErrors[0])
	}

	return allItems, nil
}
