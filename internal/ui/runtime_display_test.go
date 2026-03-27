package ui

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOpenEditor_WithEditorEnv(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "note.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	prev := os.Getenv("EDITOR")
	t.Cleanup(func() { _ = os.Setenv("EDITOR", prev) })
	if err := os.Setenv("EDITOR", "true"); err != nil {
		t.Fatalf("Setenv() error: %v", err)
	}

	if err := OpenEditor(file); err != nil {
		t.Fatalf("OpenEditor() error: %v", err)
	}
}

func TestDisplayTaskAndMessages(t *testing.T) {
	buf := &bytes.Buffer{}

	SetMessageQuiet(false)
	Info(buf, "hello")
	Infof(buf, "%s", "world")
	Success(buf, "ok")
	Successf(buf, "%d", 1)
	Warning(buf, "warn")
	Warningf(buf, "%s", "warnf")
	Error(buf, "err")
	Phase(buf, "phase")
	Stat(buf, "*", "k", 1)
	Section(buf, "sec")
	Cancelled(buf)
	Preview(buf, "title", []PreviewField{{Label: "A", Value: 1}})

	if got := buf.String(); !strings.Contains(got, "PREVIEW") || !strings.Contains(got, "Cancelled") {
		t.Fatalf("unexpected message output: %s", got)
	}

	SetMessageQuiet(true)
	quietBuf := &bytes.Buffer{}
	Info(quietBuf, "hidden")
	Success(quietBuf, "hidden")
	Warning(quietBuf, "hidden")
	Phase(quietBuf, "hidden")
	Stat(quietBuf, "*", "hidden", 1)
	Section(quietBuf, "hidden")
	Cancelled(quietBuf)
	if quietBuf.Len() != 0 {
		t.Fatalf("expected no output in quiet mode, got: %s", quietBuf.String())
	}

	task := (&Task{Name: "t", Total: 10, startTime: time.Now()})
	task.OnItemComplete()
	task.OnBatchReceived(5)
	task.OnPageFetched()
	task.OnError()
	task.Finish()

	if task.Completed() != 1 || task.Items() != 5 || task.Pages() != 1 || task.Errors() != 1 || !task.IsFinished() {
		t.Fatalf("unexpected task counters: completed=%d items=%d pages=%d errors=%d finished=%v", task.Completed(), task.Items(), task.Pages(), task.Errors(), task.IsFinished())
	}
	if task.Elapsed() < 0 {
		t.Fatalf("elapsed should be non-negative")
	}
}

func TestRunWithStatusAndOperation(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := context.Background()

	_, err := RunWithStatus[int](ctx, StatusConfig{Title: "quiet", Writer: buf, Quiet: true}, func(context.Context) (int, error) {
		return 7, nil
	})
	if err != nil {
		t.Fatalf("RunWithStatus quiet error: %v", err)
	}

	op := NewOperation(StatusConfig{Title: "op", Writer: buf, Quiet: true})
	do := op.(*displayOperation)
	do.Phase("phase")
	do.Info("info %d", 1)
	h := do.AddTask("task", 3).(*displayTaskHandle)
	h.Increment()
	h.Add(2)
	h.Page()
	h.Error(nil)
	h.OnItemComplete()
	h.OnBatchReceived(3)
	h.OnPageFetched()
	h.OnError()
	h.Finish()
	if h.Errors() != 2 {
		t.Fatalf("unexpected errors count: %d", h.Errors())
	}
	if h.Elapsed() < 0 {
		t.Fatalf("elapsed should be non-negative")
	}
	if err := do.Finish(); err != nil {
		t.Fatalf("operation Finish() error: %v", err)
	}
}
