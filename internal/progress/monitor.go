// Package progress provides universal progress monitoring for long-running operations.
// It uses channel-based communication to decouple business logic from UI updates.
package progress

import (
	"context"
	"sync/atomic"
)

// Monitor provides a thread-safe way to track progress of operations.
// It can be passed to any method that supports progress reporting.
type Monitor struct {
	// ProgressChan is the channel to send progress updates to.
	// Each Increment() call sends 1 to this channel.
	// The receiver is responsible for aggregating these updates.
	ProgressChan chan<- int

	// Total is the expected total number of items to process.
	// Used for calculating percentage, optional for Monitor itself.
	Total int

	// completed tracks the number of completed items (atomic).
	completed int64
}

// NewMonitor creates a new progress monitor with the given channel and total.
// If channel is nil, the monitor becomes a no-op (safe to use).
//
// Example:
//
//	progressChan := make(chan int, 100)
//	monitor := progress.NewMonitor(progressChan, totalSuites)
//	cases, err := client.GetCasesParallelWithMonitor(pid, suiteIDs, 5, monitor)
func NewMonitor(ch chan<- int, total int) *Monitor {
	return &Monitor{
		ProgressChan: ch,
		Total:        total,
		completed:    0,
	}
}

// Increment increments the progress counter and sends update to channel.
// This method is thread-safe and non-blocking (drops update if channel is full).
func (m *Monitor) Increment() {
	if m == nil {
		return
	}
	atomic.AddInt64(&m.completed, 1)
	if m.ProgressChan != nil {
		select {
		case m.ProgressChan <- 1:
		default: // Don't block if channel is full
		}
	}
}

// IncrementBy increments by a specific amount.
func (m *Monitor) IncrementBy(n int) {
	if m == nil || n <= 0 {
		return
	}
	atomic.AddInt64(&m.completed, int64(n))
	if m.ProgressChan != nil {
		select {
		case m.ProgressChan <- n:
		default:
		}
	}
}

// Completed returns the number of completed items.
func (m *Monitor) Completed() int64 {
	if m == nil {
		return 0
	}
	return atomic.LoadInt64(&m.completed)
}

// Percentage returns the completion percentage (0-100).
func (m *Monitor) Percentage() float64 {
	if m == nil || m.Total <= 0 {
		return 0
	}
	completed := float64(atomic.LoadInt64(&m.completed))
	return (completed / float64(m.Total)) * 100
}

// Close closes the progress channel.
// Should be called when operation is complete.
func (m *Monitor) Close() {
	if m != nil && m.ProgressChan != nil {
		close(m.ProgressChan)
	}
}

// MonitorContext wraps a Monitor with context for cancellation support.
type MonitorContext struct {
	context.Context
	*Monitor
}

// WithMonitor creates a new context with a progress monitor.
// This allows methods to accept a standard context.Context while still
// supporting progress updates.
//
// Example:
//
//	ctx, monitor := progress.WithMonitor(context.Background(), progressChan, total)
//	cases, err := client.GetCasesParallelCtx(ctx, pid, suiteIDs, 5)
func WithMonitor(parent context.Context, ch chan<- int, total int) (context.Context, *Monitor) {
	monitor := NewMonitor(ch, total)
	return &MonitorContext{
		Context: parent,
		Monitor: monitor,
	}, monitor
}

// FromContext extracts a Monitor from a context, if present.
// Returns nil if no monitor is found.
func FromContext(ctx context.Context) *Monitor {
	if mc, ok := ctx.(*MonitorContext); ok {
		return mc.Monitor
	}
	return nil
}
