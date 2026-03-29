package ui

import (
	"bytes"
	"context"
	"errors"
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

func TestRunWithStatus_NonQuietSuccess(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := context.Background()

	v, err := RunWithStatus[int](ctx, StatusConfig{Title: "sync", Writer: buf, Quiet: false}, func(context.Context) (int, error) {
		time.Sleep(20 * time.Millisecond)
		return 42, nil
	})
	if err != nil {
		t.Fatalf("RunWithStatus() error: %v", err)
	}
	if v != 42 {
		t.Fatalf("unexpected value: %d", v)
	}
	out := buf.String()
	if !strings.Contains(out, "sync completed") {
		t.Fatalf("expected completion message, got: %s", out)
	}
}

func TestRunWithStatus_NonQuietError(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := context.Background()

	_, err := RunWithStatus[int](ctx, StatusConfig{Title: "import", Writer: buf, Quiet: false}, func(context.Context) (int, error) {
		time.Sleep(20 * time.Millisecond)
		return 0, errors.New("boom")
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	out := buf.String()
	if !strings.Contains(out, "import failed") {
		t.Fatalf("expected failure message, got: %s", out)
	}
}

func TestRunWithStatus_EmptyTitleBypassesSpinner(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := context.Background()

	v, err := RunWithStatus[int](ctx, StatusConfig{Title: "", Writer: buf, Quiet: false}, func(context.Context) (int, error) {
		return 5, nil
	})
	if err != nil {
		t.Fatalf("RunWithStatus() error: %v", err)
	}
	if v != 5 {
		t.Fatalf("unexpected value: %d", v)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no status output when title is empty, got: %s", buf.String())
	}
}

func TestNewOperation_DefaultWriterAndHeader(t *testing.T) {
	op := NewOperation(StatusConfig{Title: "header", Writer: nil, Quiet: true})
	do, ok := op.(*displayOperation)
	if !ok {
		t.Fatalf("unexpected operation type")
	}
	if do.writer == nil {
		t.Fatalf("writer must be defaulted")
	}
	if do.display == nil {
		t.Fatalf("display must be initialized")
	}
}

func TestRunWithStatus_DefaultWriterWhenNil(t *testing.T) {
	ctx := context.Background()
	v, err := RunWithStatus[int](ctx, StatusConfig{Title: "", Writer: nil, Quiet: true}, func(context.Context) (int, error) {
		return 9, nil
	})
	if err != nil {
		t.Fatalf("RunWithStatus() error: %v", err)
	}
	if v != 9 {
		t.Fatalf("unexpected value: %d", v)
	}
}

func TestRunWithStatus_NonQuietSpinnerTick(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := context.Background()

	_, err := RunWithStatus[int](ctx, StatusConfig{Title: "long", Writer: buf, Quiet: false}, func(context.Context) (int, error) {
		time.Sleep(260 * time.Millisecond)
		return 1, nil
	})
	if err != nil {
		t.Fatalf("RunWithStatus() error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "long") {
		t.Fatalf("expected title in output, got: %s", out)
	}
}
