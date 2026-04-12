package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Korrnals/gotr/internal/concurrency"
)

// Operation provides an intent-level progress runtime facade for commands.
type Operation interface {
	Phase(title string)
	Info(format string, args ...any)
	AddTask(name string, total int) TaskHandle
	Finish()
}

// TaskHandle bridges intent-level task control with concurrency reporters.
type TaskHandle interface {
	concurrency.PaginatedProgressReporter
	Increment()
	Add(n int)
	Page()
	Error(err error)
	Errors() int32
	Finish()
	Elapsed() time.Duration
}

// StatusConfig configures status operations and simple status runs.
type StatusConfig struct {
	Title  string
	Writer io.Writer
	Quiet  bool
}

// displayOperation adapts Display to the Operation facade.
type displayOperation struct {
	display *Display
	writer  io.Writer
	quiet   bool
}

// displayTaskHandle adapts Task to the TaskHandle facade.
type displayTaskHandle struct {
	task *Task
}

// NewOperation creates a progress runtime backed by Display.
func NewOperation(cfg StatusConfig) Operation {
	writer := cfg.Writer
	if writer == nil {
		writer = os.Stderr
	}

	display := New(WithWriter(writer), WithQuiet(cfg.Quiet))
	if cfg.Title != "" {
		display.SetHeader(cfg.Title)
	}

	return &displayOperation{
		display: display,
		writer:  writer,
		quiet:   cfg.Quiet,
	}
}

// RunWithStatus executes a simple status-wrapped operation.
func RunWithStatus[T any](ctx context.Context, cfg StatusConfig, fn func(context.Context) (T, error)) (T, error) {
	writer := cfg.Writer
	if writer == nil {
		writer = os.Stderr
	}
	if !cfg.Quiet && cfg.Title != "" {
		stop := make(chan struct{})
		done := make(chan struct{})
		start := time.Now()
		go func() {
			defer close(done)
			ticker := time.NewTicker(200 * time.Millisecond)
			defer ticker.Stop()
			frames := []string{"|", "/", "-", `\\`}
			frame := 0
			// Render first frame immediately (no 200ms delay)
			fmt.Fprintf(writer, "\r\033[2K📥 %s %s ⏱ %s", cfg.Title, frames[0], fmtDuration(time.Since(start)))
			frame = 1
			for {
				select {
				case <-stop:
					return
				case <-ticker.C:
					fmt.Fprintf(writer, "\r\033[2K📥 %s %s ⏱ %s", cfg.Title, frames[frame%len(frames)], fmtDuration(time.Since(start)))
					frame++
				}
			}
		}()

		value, err := fn(ctx)
		close(stop)
		<-done // wait for spinner goroutine to exit before writing
		fmt.Fprintf(writer, "\r\033[2K")
		if err != nil {
			fmt.Fprintf(writer, "⚠️  %s failed after %s\n", cfg.Title, fmtDuration(time.Since(start)))
		} else {
			fmt.Fprintf(writer, "✅ %s completed in %s\n", cfg.Title, fmtDuration(time.Since(start)))
		}
		return value, err
	}
	return fn(ctx)
}

// Phase emits a phase transition message when output is enabled.
func (o *displayOperation) Phase(title string) {
	if o.quiet {
		return
	}
	Phase(o.writer, title)
}

// Info emits an informational message when output is enabled.
func (o *displayOperation) Info(format string, args ...any) {
	if o.quiet {
		return
	}
	Infof(o.writer, format, args...)
}

// AddTask creates a tracked task handle within the active operation.
func (o *displayOperation) AddTask(name string, total int) TaskHandle {
	return &displayTaskHandle{task: o.display.AddTask(name, total)}
}

// Finish finalizes the operation and renders the final display frame.
func (o *displayOperation) Finish() {
	o.display.Finish()
}

// Increment marks one logical item as completed.
func (h *displayTaskHandle) Increment() { h.task.OnItemComplete() }

// Add records a received batch size.
func (h *displayTaskHandle) Add(n int) { h.task.OnBatchReceived(n) }

// Page marks one fetched page.
func (h *displayTaskHandle) Page() { h.task.OnPageFetched() }

// Error records an error event for the task.
func (h *displayTaskHandle) Error(err error) { h.task.OnError() }

// Errors returns the number of recorded error events.
func (h *displayTaskHandle) Errors() int32 { return h.task.Errors() }

// Finish marks the task as completed.
func (h *displayTaskHandle) Finish() { h.task.Finish() }

// Elapsed returns the task execution duration.
func (h *displayTaskHandle) Elapsed() time.Duration { return h.task.Elapsed() }

// OnItemComplete implements concurrency.ProgressReporter.
func (h *displayTaskHandle) OnItemComplete() { h.Increment() }

// OnBatchReceived implements concurrency.ProgressReporter.
func (h *displayTaskHandle) OnBatchReceived(n int) { h.Add(n) }

// OnPageFetched implements concurrency.PaginatedProgressReporter.
func (h *displayTaskHandle) OnPageFetched() { h.Page() }

// OnError implements concurrency.ProgressReporter.
func (h *displayTaskHandle) OnError() { h.task.OnError() }
