// Package ui provides a unified terminal display system for gotr.
//
// It replaces mpb progress bars and scattered fmt.Fprintf(os.Stderr, ...) calls
// with a consistent, live-updating display using ANSI escape codes.
//
// Primary metric shown is CASES received (not suites), with speed and elapsed time.
// This gives meaningful real-time feedback even when individual suites take minutes.
//
// Usage:
//
//	d := ui.New()
//	d.SetHeader("Параллельная загрузка данных")
//	t := d.AddTask("Проект 30 (10 сьютов)", 10)
//	// t implements parallel.ProgressReporter — pass as config.Reporter
//	t.OnCasesReceived(250)  // cases from page
//	t.OnPageFetched()       // page done
//	t.OnSuiteComplete()     // suite done
//	d.Finish()
//
// Static helpers for one-time messages:
//
//	ui.Info(os.Stderr, "Загрузка структуры проектов...")
//	ui.Success(os.Stderr, "Готово")
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ---------------------------------------------------------------------------
// Display — live-updating terminal area
// ---------------------------------------------------------------------------

// Display renders tracked tasks with real-time statistics,
// refreshing in-place at ~5 Hz using ANSI escape codes.
type Display struct {
	w         io.Writer
	mu        sync.Mutex
	header    string
	tasks     []*Task
	startTime time.Time
	lines     int   // lines rendered in last frame
	stopped   int32 // atomic flag
	done      chan struct{}
	quiet     bool
}

// DisplayOption configures a Display.
type DisplayOption func(*Display)

// WithWriter sets the output writer (default: os.Stderr).
func WithWriter(w io.Writer) DisplayOption {
	return func(d *Display) { d.w = w }
}

// WithQuiet disables all display output.
func WithQuiet(q bool) DisplayOption {
	return func(d *Display) { d.quiet = q }
}

// New creates a Display and starts the background refresh loop.
func New(opts ...DisplayOption) *Display {
	d := &Display{
		w:         os.Stderr,
		startTime: time.Now(),
		done:      make(chan struct{}),
	}
	for _, o := range opts {
		o(d)
	}
	if !d.quiet {
		go d.refreshLoop()
	}
	return d
}

// SetHeader sets the header line shown above tasks.
func (d *Display) SetHeader(h string) {
	d.mu.Lock()
	d.header = h
	d.mu.Unlock()
}

// AddTask creates a new tracked task and returns it.
// total is the expected number of suites.
// The returned *Task implements parallel.ProgressReporter.
func (d *Display) AddTask(name string, total int) *Task {
	t := &Task{
		Name:      name,
		Total:     int32(total),
		startTime: time.Now(),
	}
	d.mu.Lock()
	d.tasks = append(d.tasks, t)
	d.mu.Unlock()
	return t
}

// Finish stops the refresh loop and prints the final frame.
func (d *Display) Finish() {
	if atomic.CompareAndSwapInt32(&d.stopped, 0, 1) {
		close(d.done)
		d.render() // final frame
		fmt.Fprintln(d.w)
	}
}

// ---------------------------------------------------------------------------
// Task — tracked operation
// ---------------------------------------------------------------------------

// Task represents one tracked operation (e.g., loading a project).
// It is safe for concurrent use via atomic operations.
//
// Task implements parallel.ProgressReporter via structural typing:
//
//	OnSuiteComplete()
//	OnCasesReceived(count int)
//	OnPageFetched()
//	OnError()
type Task struct {
	Name      string
	Total     int32 // expected suites
	startTime time.Time

	completed atomic.Int32 // suites completed
	items     atomic.Int64 // accumulated cases
	pages     atomic.Int32 // pages fetched
	errors    atomic.Int32 // errors
	finished  atomic.Int32 // 0 or 1
}

// --- Methods matching parallel.ProgressReporter ---

// OnSuiteComplete marks one suite as completed.
func (t *Task) OnSuiteComplete() { t.completed.Add(1) }

// OnCasesReceived adds received cases count.
func (t *Task) OnCasesReceived(count int) { t.items.Add(int64(count)) }

// OnPageFetched marks one page as fetched.
func (t *Task) OnPageFetched() { t.pages.Add(1) }

// OnError records an error.
func (t *Task) OnError() { t.errors.Add(1) }

// --- Getters ---

// Completed returns the number of completed suites.
func (t *Task) Completed() int32 { return t.completed.Load() }

// Items returns the total number of cases received.
func (t *Task) Items() int64 { return t.items.Load() }

// Pages returns the number of pages fetched.
func (t *Task) Pages() int32 { return t.pages.Load() }

// Errors returns the number of errors.
func (t *Task) Errors() int32 { return t.errors.Load() }

// Finish marks the task as complete for rendering.
func (t *Task) Finish() { t.finished.Store(1) }

// IsFinished returns whether the task is marked done.
func (t *Task) IsFinished() bool { return t.finished.Load() != 0 }

// Elapsed returns duration since task start.
func (t *Task) Elapsed() time.Duration { return time.Since(t.startTime) }

// ---------------------------------------------------------------------------
// Rendering
// ---------------------------------------------------------------------------

func (d *Display) refreshLoop() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-d.done:
			return
		case <-ticker.C:
			d.render()
		}
	}
}

func (d *Display) render() {
	if d.quiet {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	var buf strings.Builder

	// Erase previous frame
	if d.lines > 0 {
		fmt.Fprintf(&buf, "\033[%dA", d.lines)
		for i := 0; i < d.lines; i++ {
			buf.WriteString("\033[2K\n")
		}
		fmt.Fprintf(&buf, "\033[%dA", d.lines)
	}

	lineCount := 0

	// Header with elapsed timer
	if d.header != "" {
		elapsed := time.Since(d.startTime)
		fmt.Fprintf(&buf, "📥 %s  ⏱ %s\n", d.header, fmtDuration(elapsed))
		lineCount++
	}

	// Legend line
	if len(d.tasks) > 0 {
		buf.WriteString("   ⓘ  проект: кейсов (стр.) | сьюты | скорость | время\n")
		lineCount++
	}

	// Tasks — primary metric is CASES, not suites
	for _, t := range d.tasks {
		suitesCompleted := t.completed.Load()
		suitesTotal := t.Total
		cases := t.items.Load()
		pages := t.pages.Load()
		errs := t.errors.Load()
		elapsed := time.Since(t.startTime)
		done := t.finished.Load() != 0

		var sb strings.Builder

		// Status icon
		if done {
			sb.WriteString("   ✅ ")
		} else {
			sb.WriteString("   ⏳ ")
		}

		// Name + cases (primary metric)
		sb.WriteString(t.Name)
		sb.WriteString(": ")
		sb.WriteString(fmtCount(cases))
		sb.WriteString(" кейсов")

		// Pages
		if pages > 0 {
			sb.WriteString(fmt.Sprintf(" (%d стр.)", pages))
		}

		// Suites progress
		sb.WriteString(fmt.Sprintf(" | %d/%d сьютов", suitesCompleted, suitesTotal))

		// Speed (cases/sec) — show after 1s to avoid jitter
		if secs := elapsed.Seconds(); secs > 1.0 && cases > 0 {
			speed := float64(cases) / secs
			sb.WriteString(fmt.Sprintf(" | %s/с", fmtCount(int64(speed))))
		}

		// Elapsed
		sb.WriteString(fmt.Sprintf(" | %s", fmtDuration(elapsed)))

		// Errors (at end — conditional, so main fields always align)
		if errs > 0 {
			sb.WriteString(fmt.Sprintf(" | ⚠ %d ош.", errs))
		}

		fmt.Fprintln(&buf, sb.String())
		lineCount++
	}

	d.lines = lineCount
	fmt.Fprint(d.w, buf.String())
}

func fmtDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dмс", d.Milliseconds())
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if m > 0 {
		return fmt.Sprintf("%dм%02dс", m, s)
	}
	return fmt.Sprintf("%dс", s)
}

func fmtCount(n int64) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

// ---------------------------------------------------------------------------
// Static message helpers — one-time lines written to a writer.
// Use these for phase transitions, summaries, and status outside the live area.
// ---------------------------------------------------------------------------

// Info prints an informational message.
func Info(w io.Writer, msg string) { fmt.Fprintf(w, "ℹ️  %s\n", msg) }

// Infof prints a formatted informational message.
func Infof(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "ℹ️  %s\n", fmt.Sprintf(format, args...))
}

// Success prints a success message.
func Success(w io.Writer, msg string) { fmt.Fprintf(w, "✅ %s\n", msg) }

// Successf prints a formatted success message.
func Successf(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "✅ %s\n", fmt.Sprintf(format, args...))
}

// Warning prints a warning message.
func Warning(w io.Writer, msg string) { fmt.Fprintf(w, "⚠️  %s\n", msg) }

// Warningf prints a formatted warning message.
func Warningf(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "⚠️  %s\n", fmt.Sprintf(format, args...))
}

// Error prints an error message.
func Error(w io.Writer, msg string) { fmt.Fprintf(w, "❌ %s\n", msg) }

// Phase prints a phase transition message.
func Phase(w io.Writer, msg string) { fmt.Fprintf(w, "🔄 %s\n", msg) }

// Stat prints a statistics line with label and value.
func Stat(w io.Writer, icon, label string, value interface{}) {
	fmt.Fprintf(w, "   %s %s: %v\n", icon, label, value)
}

// Section prints a section header.
func Section(w io.Writer, msg string) { fmt.Fprintf(w, "\n📊 %s\n", msg) }
