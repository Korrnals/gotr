// Package concurrency provides priority queue implementation for suite tasks.
package concurrency

import (
	"container/heap"
	"sync"
)

// PriorityItem represents an item in the priority queue
type PriorityItem struct {
	Task     SuiteTask
	Priority Priority
	// Index is used by heap.Interface for efficient removal
	Index int
}

// priorityHeap implements heap.Interface
type priorityHeap []PriorityItem

func (h priorityHeap) Len() int { return len(h) }

// Less defines priority order: higher priority values come first
func (h priorityHeap) Less(i, j int) bool {
	if h[i].Priority != h[j].Priority {
		return h[i].Priority > h[j].Priority
	}

	if h[i].Task.EstimatedSize != h[j].Task.EstimatedSize {
		return h[i].Task.EstimatedSize > h[j].Task.EstimatedSize
	}

	return h[i].Task.SuiteID < h[j].Task.SuiteID
}

func (h priorityHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}

func (h *priorityHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(PriorityItem)
	item.Index = n
	*h = append(*h, item)
}

func (h *priorityHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	item.Index = -1 // for safety
	*h = old[:n-1]
	return item
}

// PriorityQueue is a thread-safe priority queue for SuiteTasks
type PriorityQueue struct {
	heap  priorityHeap
	mu    sync.RWMutex
	cond  *sync.Cond
	closed bool
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{
		heap: make(priorityHeap, 0),
	}
	pq.cond = sync.NewCond(&pq.mu)
	heap.Init(&pq.heap)
	return pq
}

// Push adds a task to the queue with its natural priority
func (pq *PriorityQueue) Push(task SuiteTask) {
	pq.PushWithPriority(task, task.GetPriority())
}

// PushWithPriority adds a task with a specific priority
func (pq *PriorityQueue) PushWithPriority(task SuiteTask, priority Priority) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if pq.closed {
		return
	}

	item := PriorityItem{
		Task:     task,
		Priority: priority,
	}
	heap.Push(&pq.heap, item)
	pq.cond.Signal() // Wake up one waiting Pop
}

// Pop removes and returns the highest priority task.
// Blocks if the queue is empty until a task is available or the queue is closed.
// Returns (task, true) on success, (_, false) if queue is closed and empty.
func (pq *PriorityQueue) Pop() (SuiteTask, bool) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	for len(pq.heap) == 0 && !pq.closed {
		pq.cond.Wait()
	}

	if len(pq.heap) == 0 {
		return SuiteTask{}, false
	}

	item := heap.Pop(&pq.heap).(PriorityItem)
	return item.Task, true
}

// TryPop attempts to pop without blocking.
// Returns (task, true) on success, (_, false) if queue is empty.
func (pq *PriorityQueue) TryPop() (SuiteTask, bool) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.heap) == 0 {
		return SuiteTask{}, false
	}

	item := heap.Pop(&pq.heap).(PriorityItem)
	return item.Task, true
}

// Len returns the current number of items in the queue
func (pq *PriorityQueue) Len() int {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return len(pq.heap)
}

// IsEmpty returns true if the queue is empty
func (pq *PriorityQueue) IsEmpty() bool {
	return pq.Len() == 0
}

// Close closes the queue, unblocking any waiting Pop operations.
// After Close, Pop will return (_, false) when empty.
func (pq *PriorityQueue) Close() {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	pq.closed = true
	pq.cond.Broadcast() // Wake all waiters
}

// IsClosed returns true if the queue has been closed
func (pq *PriorityQueue) IsClosed() bool {
	pq.mu.RLock()
	defer pq.mu.RUnlock()
	return pq.closed
}

// Peek returns the highest priority task without removing it
func (pq *PriorityQueue) Peek() (SuiteTask, bool) {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	if len(pq.heap) == 0 {
		return SuiteTask{}, false
	}

	return pq.heap[0].Task, true
}

// GetAll returns all tasks in the queue (for debugging/metrics)
// Does not remove items from the queue
func (pq *PriorityQueue) GetAll() []SuiteTask {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	tasks := make([]SuiteTask, len(pq.heap))
	for i, item := range pq.heap {
		tasks[i] = item.Task
	}
	return tasks
}

// GetStats returns statistics about the queue contents
func (pq *PriorityQueue) GetStats() QueueStats {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	stats := QueueStats{
		Total: len(pq.heap),
	}

	for _, item := range pq.heap {
		switch item.Priority {
		case PriorityHigh:
			stats.HighPriority++
		case PriorityMedium:
			stats.MediumPriority++
		case PriorityLow:
			stats.LowPriority++
		}
	}

	return stats
}

// QueueStats contains statistics about queue contents
type QueueStats struct {
	Total          int
	HighPriority   int
	MediumPriority int
	LowPriority    int
}
