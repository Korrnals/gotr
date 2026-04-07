package ui

import (
	"bytes"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFmtDuration(t *testing.T) {
	assert.Equal(t, "500ms", fmtDuration(500*time.Millisecond))
	assert.Equal(t, "2s", fmtDuration(2*time.Second))
	assert.Equal(t, "1m05s", fmtDuration(65*time.Second))
}

func TestFmtCount(t *testing.T) {
	assert.Equal(t, "999", fmtCount(999))
	assert.Equal(t, "1.0K", fmtCount(1000))
	assert.Equal(t, "1.5K", fmtCount(1500))
	assert.Equal(t, "1.0M", fmtCount(1_000_000))
}

func TestDisplay_RenderAndFinish(t *testing.T) {
	var buf bytes.Buffer
	d := New(WithWriter(&buf))
	t.Cleanup(func() {
		if atomic.LoadInt32(&d.stopped) == 0 {
			d.Finish()
		}
	})

	d.SetHeader("Coverage wave")
	task := d.AddTask("project 30", 2)
	task.OnBatchReceived(1500)
	task.OnPageFetched()
	task.OnItemComplete()
	task.OnError()
	task.startTime = time.Now().Add(-2 * time.Second)

	d.render()
	first := buf.String()
	assert.Contains(t, first, "Coverage wave")
	assert.Contains(t, first, "project 30")
	assert.Contains(t, first, "cases")
	assert.Contains(t, first, "err")

	d.render()
	second := buf.String()
	assert.Contains(t, second, "\033[")

	d.Finish()
	assert.Equal(t, int32(1), atomic.LoadInt32(&d.stopped))
}

func TestDisplay_RefreshLoopStopsOnDone(t *testing.T) {
	var buf bytes.Buffer
	d := &Display{w: &buf, done: make(chan struct{})}
	finished := make(chan struct{})

	go func() {
		d.refreshLoop()
		close(finished)
	}()

	time.Sleep(250 * time.Millisecond)
	close(d.done)

	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("refreshLoop did not stop after done close")
	}
}

func TestDisplay_QuietNoOutput(t *testing.T) {
	var buf bytes.Buffer
	d := New(WithWriter(&buf), WithQuiet(true))
	d.SetHeader("hidden")
	_ = d.AddTask("task", 1)
	d.Finish()
	assert.Equal(t, "", buf.String())
}

func TestFormattedHelpers_RespectMessageQuiet(t *testing.T) {
	var buf bytes.Buffer
	SetMessageQuiet(false)
	Infof(&buf, "info %d", 1)
	Successf(&buf, "ok %s", "x")
	Warningf(&buf, "warn %s", "y")
	assert.Contains(t, buf.String(), "info 1")
	assert.Contains(t, buf.String(), "ok x")
	assert.Contains(t, buf.String(), "warn y")

	before := buf.String()
	SetMessageQuiet(true)
	Infof(&buf, "suppressed")
	Successf(&buf, "suppressed")
	Warningf(&buf, "suppressed")
	assert.Equal(t, before, buf.String())

	SetMessageQuiet(false)
	Phase(&buf, "phase")
	assert.True(t, strings.Contains(buf.String(), "phase"))
}

// TestDisplay_QuietWithErrors verifies quiet mode error handling.
// Covers: Error message functions respect quiet mode.
func TestDisplay_QuietWithErrors(t *testing.T) {
	var buf bytes.Buffer
	SetMessageQuiet(true)

	Error(&buf, "this should always show")
	assert.Contains(t, buf.String(), "this should always show")

	before := buf.String()
	Info(&buf, "suppressed")
	assert.Equal(t, before, buf.String())

	SetMessageQuiet(false)
}

// TestDisplay_MultipleTasksConcurrent tests concurrent task updates.
// Covers: thread-safe updates to multiple tasks.
func TestDisplay_MultipleTasksConcurrent(t *testing.T) {
	var buf bytes.Buffer
	d := New(WithWriter(&buf))
	t.Cleanup(func() {
		if atomic.LoadInt32(&d.stopped) == 0 {
			d.Finish()
		}
	})

	d.SetHeader("Concurrent test")
	task1 := d.AddTask("task 1", 10)
	task2 := d.AddTask("task 2", 5)
	task3 := d.AddTask("task 3", 8)

	// Simulate concurrent updates
	done := make(chan struct{})
	for i := 0; i < 3; i++ {
		go func(t *Task) {
			for j := 0; j < 100; j++ {
				t.OnBatchReceived(10)
				t.OnItemComplete()
				t.OnPageFetched()
			}
			done <- struct{}{}
		}([]*Task{task1, task2, task3}[i])
	}

	for i := 0; i < 3; i++ {
		<-done
	}

	d.render()
	output := buf.String()
	assert.Contains(t, output, "task 1")
	assert.Contains(t, output, "task 2")
	assert.Contains(t, output, "task 3")
}

// TestDisplay_LargeOutput tests display with large number of cases.
// Covers: performance paths with large numbers.
func TestDisplay_LargeOutput(t *testing.T) {
	var buf bytes.Buffer
	d := New(WithWriter(&buf))
	t.Cleanup(func() {
		if atomic.LoadInt32(&d.stopped) == 0 {
			d.Finish()
		}
	})

	d.SetHeader("Large output test")
	task := d.AddTask("heavy project", 1000)

	// Simulate large data ingestion
	task.OnBatchReceived(1_000_000)
	task.OnPageFetched()
	for i := 0; i < 100; i++ {
		task.OnItemComplete()
	}

	d.render()
	output := buf.String()
	assert.Contains(t, output, "heavy project")
	assert.Contains(t, output, "M")   // Should format as millions
	assert.Contains(t, output, "cases")
}

// TestDisplay_TaskWithZeroTotal tests task tracking with zero total suites.
// Covers: edge case when total is 0.
func TestDisplay_TaskWithZeroTotal(t *testing.T) {
	var buf bytes.Buffer
	d := New(WithWriter(&buf))
	t.Cleanup(func() {
		if atomic.LoadInt32(&d.stopped) == 0 {
			d.Finish()
		}
	})

	task := d.AddTask("zero total", 0)
	task.OnBatchReceived(100)
	d.render()

	output := buf.String()
	assert.Contains(t, output, "zero total")
	assert.Contains(t, output, "0/0")
}

// TestDisplay_ErrorAccumulation tests error tracking.
// Covers: OnError and Errors() getter.
func TestDisplay_ErrorAccumulation(t *testing.T) {
	var buf bytes.Buffer
	d := New(WithWriter(&buf))
	t.Cleanup(func() {
		if atomic.LoadInt32(&d.stopped) == 0 {
			d.Finish()
		}
	})

	task := d.AddTask("error test", 5)
	for i := 0; i < 3; i++ {
		task.OnError()
	}

	assert.Equal(t, int32(3), task.Errors())

	d.render()
	output := buf.String()
	assert.Contains(t, output, "err")
}

// TestDisplay_TaskFinishFlag tests task completion marking.
// Covers: Finish(), IsFinished() methods on Task.
func TestDisplay_TaskFinishFlag(t *testing.T) {
	var buf bytes.Buffer
	d := New(WithWriter(&buf))
	t.Cleanup(func() {
		if atomic.LoadInt32(&d.stopped) == 0 {
			d.Finish()
		}
	})

	task := d.AddTask("finish test", 5)
	assert.False(t, task.IsFinished())

	task.Finish()
	assert.True(t, task.IsFinished())

	d.render()
	output := buf.String()
	assert.Contains(t, output, "✅") // Finished tasks show checkmark
}

// TestDisplay_SpeedCalculation tests speed calculation after elapsed time.
// Covers: speed showing only after 1 second.
func TestDisplay_SpeedCalculation(t *testing.T) {
	var buf bytes.Buffer
	d := New(WithWriter(&buf))
	t.Cleanup(func() {
		if atomic.LoadInt32(&d.stopped) == 0 {
			d.Finish()
		}
	})

	task := d.AddTask("speed test", 1)
	task.OnBatchReceived(1000)

	// Set start time to 2 seconds ago
	task.startTime = time.Now().Add(-2 * time.Second)

	d.render()
	output := buf.String()
	// After 2 seconds with 1000 cases, should show speed (roughly 500/s)
	assert.Contains(t, output, "/s")
}

// TestDisplay_PageFetchTracking tests page count accumulation.
// Covers: OnPageFetched() and Pages() getter.
func TestDisplay_PageFetchTracking(t *testing.T) {
	var buf bytes.Buffer
	d := New(WithWriter(&buf))
	t.Cleanup(func() {
		if atomic.LoadInt32(&d.stopped) == 0 {
			d.Finish()
		}
	})

	task := d.AddTask("page tracking", 10)
	for i := 0; i < 5; i++ {
		task.OnPageFetched()
	}

	assert.Equal(t, int32(5), task.Pages())

	d.render()
	output := buf.String()
	assert.Contains(t, output, "5 pages")
}

// TestDisplay_ItemCompleted tests suite completion tracking.
// Covers: OnItemComplete() and Completed() getter.
func TestDisplay_ItemCompleted(t *testing.T) {
	var buf bytes.Buffer
	d := New(WithWriter(&buf))
	t.Cleanup(func() {
		if atomic.LoadInt32(&d.stopped) == 0 {
			d.Finish()
		}
	})

	task := d.AddTask("completion test", 10)
	for i := 0; i < 7; i++ {
		task.OnItemComplete()
	}

	assert.Equal(t, int32(7), task.Completed())

	d.render()
	output := buf.String()
	assert.Contains(t, output, "7/10")
}

// TestDisplay_DoubleFinish tests idempotent Finish() call.
// Covers: second Finish() does nothing (compared via atomic).
func TestDisplay_DoubleFinish(t *testing.T) {
	var buf bytes.Buffer
	d := New(WithWriter(&buf))

	d.Finish()
	state1 := atomic.LoadInt32(&d.stopped)
	assert.Equal(t, int32(1), state1)

	// Second Finish should not change anything
	d.Finish()
	state2 := atomic.LoadInt32(&d.stopped)
	assert.Equal(t, state1, state2)
}

// TestDisplay_HeaderUpdate tests header changes.
// Covers: SetHeader() updates and renders.
func TestDisplay_HeaderUpdate(t *testing.T) {
	var buf bytes.Buffer
	d := New(WithWriter(&buf))
	t.Cleanup(func() {
		if atomic.LoadInt32(&d.stopped) == 0 {
			d.Finish()
		}
	})

	d.SetHeader("First header")
	d.render()
	first := buf.String()

	buf.Reset()
	d.SetHeader("Second header")
	d.render()
	second := buf.String()

	assert.Contains(t, first, "First header")
	assert.Contains(t, second, "Second header")
}

// TestDisplay_RenderWithoutTasks tests rendering with no tasks.
// Covers: render() with empty task list.
func TestDisplay_RenderWithoutTasks(t *testing.T) {
	var buf bytes.Buffer
	d := New(WithWriter(&buf))
	t.Cleanup(func() {
		if atomic.LoadInt32(&d.stopped) == 0 {
			d.Finish()
		}
	})

	d.SetHeader("No tasks")
	d.render()

	output := buf.String()
	assert.Contains(t, output, "No tasks")
	// Should not have task legend when no tasks
	assert.NotContains(t, output, "project:")
}

// TestDisplay_ElapsedTiming tests elapsed time calculation.
// Covers: Elapsed() duration getter for tasks.
func TestDisplay_ElapsedTiming(t *testing.T) {
	var buf bytes.Buffer
	d := New(WithWriter(&buf))
	t.Cleanup(func() {
		if atomic.LoadInt32(&d.stopped) == 0 {
			d.Finish()
		}
	})

	task := d.AddTask("timing test", 1)
	task.startTime = time.Now().Add(-5 * time.Second)

	elapsed := task.Elapsed()
	assert.True(t, elapsed >= 5*time.Second)
	assert.True(t, elapsed < 6*time.Second)
}

// TestMessage_AllVariants tests all message helper variants.
// Covers: Info, Infof, Success, Successf, Warning, Warningf, Error, Phase, Stat, Section.
func TestMessage_AllVariants(t *testing.T) {
	var buf bytes.Buffer
	SetMessageQuiet(false)

	Info(&buf, "info")
	Infof(&buf, "info %s", "formatted")
	Success(&buf, "success")
	Successf(&buf, "success %d", 1)
	Warning(&buf, "warning")
	Warningf(&buf, "warning %s", "formatted")
	Error(&buf, "error")
	Phase(&buf, "phase")
	Stat(&buf, "📊", "label", "value")
	Section(&buf, "section")

	output := buf.String()
	assert.Contains(t, output, "info")
	assert.Contains(t, output, "formatted")
	assert.Contains(t, output, "success")
	assert.Contains(t, output, "warning")
	assert.Contains(t, output, "error")
	assert.Contains(t, output, "phase")
	assert.Contains(t, output, "label")
	assert.Contains(t, output, "section")

	SetMessageQuiet(false)
}

// TestFmtDuration_EdgeCases tests duration formatting edge cases.
// Covers: boundary cases for milliseconds, seconds, minutes.
func TestFmtDuration_EdgeCases(t *testing.T) {
	assert.Equal(t, "0ms", fmtDuration(0))
	assert.Equal(t, "1ms", fmtDuration(time.Millisecond))
	assert.Equal(t, "999ms", fmtDuration(999*time.Millisecond))
	assert.Equal(t, "1s", fmtDuration(time.Second))
	assert.Equal(t, "59s", fmtDuration(59*time.Second))
	assert.Equal(t, "1m00s", fmtDuration(60*time.Second))
	assert.Equal(t, "10m00s", fmtDuration(10*time.Minute))
}

// TestFmtCount_EdgeCases tests count formatting edge cases.
// Covers: boundary cases for K and M formatting.
func TestFmtCount_EdgeCases(t *testing.T) {
	assert.Equal(t, "0", fmtCount(0))
	assert.Equal(t, "1", fmtCount(1))
	assert.Equal(t, "999", fmtCount(999))
	assert.Equal(t, "1.0K", fmtCount(1000))
	assert.Equal(t, "10.0K", fmtCount(10000))
	assert.Equal(t, "999.9K", fmtCount(999900))
	assert.Equal(t, "1.0M", fmtCount(1000000))
	assert.Equal(t, "10.0M", fmtCount(10000000))
}

// TestDisplay_QuietRender verifies render() returns early in quiet mode.
// Covers: display.go render() quiet branch.
func TestDisplay_QuietRender(t *testing.T) {
	var buf bytes.Buffer
	d := New(WithWriter(&buf), WithQuiet(true))
	d.SetHeader("hidden header")
	_ = d.AddTask("hidden task", 5)
	// Directly invoke render to cover the quiet early-return.
	d.render()
	assert.Equal(t, "", buf.String())
}
