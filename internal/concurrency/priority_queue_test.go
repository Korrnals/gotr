package concurrency

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPriorityQueue_Basic(t *testing.T) {
	pq := NewPriorityQueue()

	// Push tasks with different estimated sizes
	tasks := []SuiteTask{
		{SuiteID: 1, ProjectID: 10, EstimatedSize: 50},   // Low priority
		{SuiteID: 2, ProjectID: 10, EstimatedSize: 500},  // Medium priority
		{SuiteID: 3, ProjectID: 10, EstimatedSize: 1500}, // High priority
		{SuiteID: 4, ProjectID: 10, EstimatedSize: 100},  // Medium priority
	}

	for _, task := range tasks {
		pq.Push(task)
	}

	assert.Equal(t, 4, pq.Len())

	// Pop should return in priority order: High > Medium > Low
	// Suite 3 (1500, High) should come first
	task, ok := pq.Peek()
	assert.True(t, ok)
	assert.Equal(t, int64(3), task.SuiteID)
	assert.Equal(t, PriorityHigh, task.GetPriority())

	// Now pop it
	task, ok = pq.Pop()
	assert.True(t, ok)
	assert.Equal(t, int64(3), task.SuiteID)

	// Next two should be Medium priority (Suite 2 and 4 in any order)
	var mediumIDs []int64
	for i := 0; i < 2; i++ {
		task, ok = pq.Pop()
		assert.True(t, ok)
		assert.Equal(t, PriorityMedium, task.GetPriority())
		mediumIDs = append(mediumIDs, task.SuiteID)
	}
	assert.ElementsMatch(t, []int64{2, 4}, mediumIDs)

	// Last should be Low priority (Suite 1)
	task, ok = pq.Pop()
	assert.True(t, ok)
	assert.Equal(t, int64(1), task.SuiteID)
	assert.Equal(t, PriorityLow, task.GetPriority())

	assert.Equal(t, 0, pq.Len())
}

func TestPriorityQueue_TryPop(t *testing.T) {
	pq := NewPriorityQueue()

	// TryPop on empty queue should return false
	_, ok := pq.TryPop()
	assert.False(t, ok)

	// Push and TryPop
	pq.Push(SuiteTask{SuiteID: 1, EstimatedSize: 100})

	task, ok := pq.TryPop()
	assert.True(t, ok)
	assert.Equal(t, int64(1), task.SuiteID)

	// Queue should be empty again
	assert.True(t, pq.IsEmpty())
}

func TestPriorityQueue_Close(t *testing.T) {
	pq := NewPriorityQueue()

	// Close empty queue
	pq.Close()
	assert.True(t, pq.IsClosed())

	// Pop on closed empty queue should return false
	_, ok := pq.Pop()
	assert.False(t, ok)
}

func TestPriorityQueue_CloseWithItems(t *testing.T) {
	pq := NewPriorityQueue()

	// Push items
	pq.Push(SuiteTask{SuiteID: 1, EstimatedSize: 100})
	pq.Push(SuiteTask{SuiteID: 2, EstimatedSize: 200})

	// Close
	pq.Close()

	// Can still Pop remaining items
	task, ok := pq.Pop()
	assert.True(t, ok)
	assert.Equal(t, int64(2), task.SuiteID)

	task, ok = pq.Pop()
	assert.True(t, ok)
	assert.Equal(t, int64(1), task.SuiteID)

	// Now empty, should return false
	_, ok = pq.Pop()
	assert.False(t, ok)
}

func TestPriorityQueue_PopBlocking(t *testing.T) {
	pq := NewPriorityQueue()

	// Pop should block until item is available
	done := make(chan SuiteTask, 1)

	go func() {
		task, _ := pq.Pop()
		done <- task
	}()

	// Wait a bit to ensure goroutine is blocked
	time.Sleep(50 * time.Millisecond)

	// Push item
	pq.Push(SuiteTask{SuiteID: 42, EstimatedSize: 100})

	// Should receive the item
	select {
	case task := <-done:
		assert.Equal(t, int64(42), task.SuiteID)
	case <-time.After(time.Second):
		t.Fatal("Pop did not unblock after Push")
	}
}

func TestPriorityQueue_Peek(t *testing.T) {
	pq := NewPriorityQueue()

	// Peek on empty queue
	_, ok := pq.Peek()
	assert.False(t, ok)

	// Push items
	pq.Push(SuiteTask{SuiteID: 1, EstimatedSize: 100})
	pq.Push(SuiteTask{SuiteID: 2, EstimatedSize: 200})

	// Peek should return highest priority without removing
	task, ok := pq.Peek()
	assert.True(t, ok)
	assert.Equal(t, int64(2), task.SuiteID)
	assert.Equal(t, 2, pq.Len()) // Length unchanged
}

func TestPriorityQueue_GetStats(t *testing.T) {
	pq := NewPriorityQueue()

	// Empty queue stats
	stats := pq.GetStats()
	assert.Equal(t, 0, stats.Total)

	// Add items with different priorities
	pq.Push(SuiteTask{SuiteID: 1, EstimatedSize: 50})   // Low
	pq.Push(SuiteTask{SuiteID: 2, EstimatedSize: 500})  // Medium
	pq.Push(SuiteTask{SuiteID: 3, EstimatedSize: 1500}) // High
	pq.Push(SuiteTask{SuiteID: 4, EstimatedSize: 100})  // Medium
	pq.Push(SuiteTask{SuiteID: 5, EstimatedSize: 2000}) // High

	stats = pq.GetStats()
	assert.Equal(t, 5, stats.Total)
	assert.Equal(t, 2, stats.HighPriority)
	assert.Equal(t, 2, stats.MediumPriority)
	assert.Equal(t, 1, stats.LowPriority)
}

func TestPriorityQueue_Concurrent(t *testing.T) {
	pq := NewPriorityQueue()
	numProducers := 5
	numItemsPerProducer := 100

	var wg sync.WaitGroup

	// Multiple producers
	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()
			for j := 0; j < numItemsPerProducer; j++ {
				// Mix of priorities
				size := (producerID*numItemsPerProducer + j) % 2000
				pq.Push(SuiteTask{
					SuiteID:       int64(producerID*1000000 + j),
					EstimatedSize: size,
				})
			}
		}(i)
	}

	// Wait for all pushes to complete
	wg.Wait()

	expectedTotal := numProducers * numItemsPerProducer
	assert.Equal(t, expectedTotal, pq.Len())

	// Pop all items - should be in priority order
	var prevPriority Priority = PriorityHigh + 1
	for i := 0; i < expectedTotal; i++ {
		task, ok := pq.Pop()
		assert.True(t, ok, "Failed to pop item %d", i)
		currentPriority := task.GetPriority()
		assert.True(t, currentPriority <= prevPriority,
			"Priority order violated: got %d after %d", currentPriority, prevPriority)
		prevPriority = currentPriority
	}

	assert.True(t, pq.IsEmpty())
}

func TestSuiteTask_GetPriority(t *testing.T) {
	tests := []struct {
		name          string
		estimatedSize int
		want          Priority
	}{
		{"large suite", 1500, PriorityHigh},
		{"exactly 1000", 1000, PriorityHigh},
		{"medium suite", 500, PriorityMedium},
		{"exactly 100", 100, PriorityMedium},
		{"small suite", 50, PriorityLow},
		{"zero", 0, PriorityLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := SuiteTask{EstimatedSize: tt.estimatedSize}
			got := task.GetPriority()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPriorityQueue_PushWithPriority(t *testing.T) {
	pq := NewPriorityQueue()

	// Push with explicit priority (overriding natural priority)
	pq.PushWithPriority(SuiteTask{SuiteID: 1, EstimatedSize: 50}, PriorityHigh)
	pq.PushWithPriority(SuiteTask{SuiteID: 2, EstimatedSize: 1500}, PriorityLow)

	task, _ := pq.Pop()
	assert.Equal(t, int64(1), task.SuiteID) // High priority despite small size

	task, _ = pq.Pop()
	assert.Equal(t, int64(2), task.SuiteID) // Low priority despite large size
}
