// Package concurrent provides utilities for concurrent API requests with rate limiting.
package concurrent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// WorkerPool manages a pool of workers for concurrent operations.
type WorkerPool struct {
	limiter     *RateLimiter
	maxWorkers  int
	errGroup    *errgroup.Group
	errGroupCtx context.Context
}

// PoolOption configures the WorkerPool.
type PoolOption func(*WorkerPool)

// WithMaxWorkers sets the maximum number of concurrent workers.
func WithMaxWorkers(n int) PoolOption {
	return func(p *WorkerPool) {
		p.maxWorkers = n
	}
}

// WithRateLimit sets the rate limit (requests per minute).
func WithRateLimit(requestsPerMinute int) PoolOption {
	return func(p *WorkerPool) {
		p.limiter = NewRateLimiter(requestsPerMinute)
	}
}

// NewWorkerPool creates a new worker pool with the given options.
func NewWorkerPool(opts ...PoolOption) *WorkerPool {
	pool := &WorkerPool{
		maxWorkers: 5, // Default: 5 concurrent workers
		limiter:    NewRateLimiter(150), // Default: 150 req/min (TestRail limit)
	}

	for _, opt := range opts {
		opt(pool)
	}

	pool.errGroup, pool.errGroupCtx = errgroup.WithContext(context.Background())
	pool.errGroup.SetLimit(pool.maxWorkers)

	return pool
}

// Submit submits a task to the pool.
func (p *WorkerPool) Submit(task func() error) {
	p.errGroup.Go(func() error {
		// Wait for rate limiter
		if p.limiter != nil {
			p.limiter.Wait()
		}
		return task()
	})
}

// Wait waits for all tasks to complete and returns the first error.
func (p *WorkerPool) Wait() error {
	return p.errGroup.Wait()
}

// Context returns the context for the pool.
func (p *WorkerPool) Context() context.Context {
	return p.errGroupCtx
}

// Result represents a result from a concurrent operation.
type Result[T any] struct {
	Data  T
	Error error
	Index int
}

// ParallelMap executes a function in parallel for each item in the slice.
func ParallelMap[T any, R any](items []T, maxWorkers int, fn func(T, int) (R, error)) ([]Result[R], error) {
	if len(items) == 0 {
		return nil, nil
	}

	if maxWorkers <= 0 {
		maxWorkers = 5
	}

	results := make([]Result[R], len(items))
	var mu sync.Mutex

	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(maxWorkers)

	for i, item := range items {
		i, item := i, item // Capture loop variables
		g.Go(func() error {
			data, err := fn(item, i)
			mu.Lock()
			results[i] = Result[R]{Data: data, Error: err, Index: i}
			mu.Unlock()
			return nil // We don't stop on individual errors
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

// ParallelForEach executes a function in parallel for each item (no return value).
func ParallelForEach[T any](items []T, maxWorkers int, fn func(T, int) error) error {
	if len(items) == 0 {
		return nil
	}

	if maxWorkers <= 0 {
		maxWorkers = 5
	}

	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(maxWorkers)

	for i, item := range items {
		i, item := i, item // Capture loop variables
		g.Go(func() error {
			return fn(item, i)
		})
	}

	return g.Wait()
}

// BatchProcessor processes items in batches with rate limiting.
type BatchProcessor[T any] struct {
	batchSize     int
	delayBetween  time.Duration
	maxRetries    int
	retryDelay    time.Duration
}

// BatchOption configures the BatchProcessor.
type BatchOption[T any] func(*BatchProcessor[T])

// WithBatchSize sets the batch size.
func WithBatchSize[T any](size int) BatchOption[T] {
	return func(bp *BatchProcessor[T]) {
		bp.batchSize = size
	}
}

// WithDelayBetweenBatches sets the delay between batches.
func WithDelayBetweenBatches[T any](delay time.Duration) BatchOption[T] {
	return func(bp *BatchProcessor[T]) {
		bp.delayBetween = delay
	}
}

// WithRetryPolicy sets the retry policy.
func WithRetryPolicy[T any](maxRetries int, retryDelay time.Duration) BatchOption[T] {
	return func(bp *BatchProcessor[T]) {
		bp.maxRetries = maxRetries
		bp.retryDelay = retryDelay
	}
}

// NewBatchProcessor creates a new batch processor.
func NewBatchProcessor[T any](opts ...BatchOption[T]) *BatchProcessor[T] {
	bp := &BatchProcessor[T]{
		batchSize:    10,
		delayBetween: 100 * time.Millisecond,
		maxRetries:   3,
		retryDelay:   time.Second,
	}

	for _, opt := range opts {
		opt(bp)
	}

	return bp
}

// Process processes items in batches.
func (bp *BatchProcessor[T]) Process(items []T, processor func([]T) error) error {
	if len(items) == 0 {
		return nil
	}

	for i := 0; i < len(items); i += bp.batchSize {
		end := i + bp.batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]

		var err error
		for attempt := 0; attempt <= bp.maxRetries; attempt++ {
			if attempt > 0 {
				time.Sleep(bp.retryDelay)
			}

			err = processor(batch)
			if err == nil {
				break
			}
		}

		if err != nil {
			return fmt.Errorf("batch %d-%d failed after %d attempts: %w", i, end, bp.maxRetries+1, err)
		}

		if end < len(items) && bp.delayBetween > 0 {
			time.Sleep(bp.delayBetween)
		}
	}

	return nil
}
