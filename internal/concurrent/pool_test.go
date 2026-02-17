package concurrent

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewWorkerPool(t *testing.T) {
	pool := NewWorkerPool()
	assert.NotNil(t, pool)
	assert.NotNil(t, pool.limiter)
	assert.Equal(t, 5, pool.maxWorkers)
}

func TestNewWorkerPool_WithOptions(t *testing.T) {
	pool := NewWorkerPool(
		WithMaxWorkers(10),
		WithRateLimit(300),
	)
	assert.NotNil(t, pool)
	assert.Equal(t, 10, pool.maxWorkers)
	assert.NotNil(t, pool.limiter)
}

func TestWorkerPool_Submit(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(2))
	
	var counter int32
	for i := 0; i < 5; i++ {
		pool.Submit(func() error {
			atomic.AddInt32(&counter, 1)
			return nil
		})
	}

	err := pool.Wait()
	assert.NoError(t, err)
	assert.Equal(t, int32(5), counter)
}

func TestWorkerPool_SubmitWithError(t *testing.T) {
	pool := NewWorkerPool(WithMaxWorkers(2))
	
	expectedErr := errors.New("test error")
	pool.Submit(func() error {
		return expectedErr
	})

	err := pool.Wait()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test error")
}

func TestParallelMap(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	
	results, err := ParallelMap(items, 3, func(item, index int) (int, error) {
		return item * 2, nil
	})

	assert.NoError(t, err)
	assert.Len(t, results, 5)
	
	for i, result := range results {
		assert.NoError(t, result.Error)
		assert.Equal(t, i, result.Index)
		assert.Equal(t, items[i]*2, result.Data)
	}
}

func TestParallelMap_EmptySlice(t *testing.T) {
	results, err := ParallelMap([]int{}, 3, func(item, index int) (int, error) {
		return item, nil
	})

	assert.NoError(t, err)
	assert.Nil(t, results)
}

func TestParallelForEach(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	var counter int32

	err := ParallelForEach(items, 3, func(item, index int) error {
		atomic.AddInt32(&counter, 1)
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, int32(5), counter)
}

func TestParallelForEach_WithError(t *testing.T) {
	items := []int{1, 2, 3}
	expectedErr := errors.New("test error")

	err := ParallelForEach(items, 3, func(item, index int) error {
		if item == 2 {
			return expectedErr
		}
		return nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test error")
}

func TestBatchProcessor_Process(t *testing.T) {
	bp := NewBatchProcessor[int](
		WithBatchSize[int](3),
		WithDelayBetweenBatches[int](10*time.Millisecond),
	)

	items := []int{1, 2, 3, 4, 5, 6, 7}
	var batches [][]int

	err := bp.Process(items, func(batch []int) error {
		batches = append(batches, batch)
		return nil
	})

	assert.NoError(t, err)
	assert.Len(t, batches, 3) // 3 items per batch: [1,2,3], [4,5,6], [7]
	assert.Len(t, batches[0], 3)
	assert.Len(t, batches[1], 3)
	assert.Len(t, batches[2], 1)
}

func TestBatchProcessor_ProcessWithRetry(t *testing.T) {
	bp := NewBatchProcessor[int](
		WithBatchSize[int](2),
		WithRetryPolicy[int](2, 10*time.Millisecond),
	)

	items := []int{1, 2}
	attempts := 0

	err := bp.Process(items, func(batch []int) error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary error")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, attempts)
}

func TestBatchProcessor_ProcessError(t *testing.T) {
	bp := NewBatchProcessor[int](
		WithBatchSize[int](2),
		WithRetryPolicy[int](1, 10*time.Millisecond),
	)

	items := []int{1, 2}

	err := bp.Process(items, func(batch []int) error {
		return errors.New("permanent error")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permanent error")
}

func TestBatchProcessor_EmptySlice(t *testing.T) {
	bp := NewBatchProcessor[int]()

	called := false
	err := bp.Process([]int{}, func(batch []int) error {
		called = true
		return nil
	})

	assert.NoError(t, err)
	assert.False(t, called)
}
