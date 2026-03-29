package ui

import (
	"bytes"
	"context"
	"testing"
)

// TestRunWithStatus_Success tests RunWithStatus with successful operation
func TestRunWithStatus_Success(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := context.Background()

	result, err := RunWithStatus[string](ctx, StatusConfig{
		Title:  "Loading",
		Writer: buf,
		Quiet:  false,
	}, func(ctx context.Context) (string, error) {
		return "done", nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "done" {
		t.Fatalf("expected 'done', got %q", result)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Loading")) {
		t.Fatalf("expected 'Loading' in output, got %q", buf.String())
	}
}

// TestRunWithStatus_Quiet tests RunWithStatus in quiet mode
func TestRunWithStatus_Quiet(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := context.Background()

	result, err := RunWithStatus[int](ctx, StatusConfig{
		Title:  "Silent",
		Writer: buf,
		Quiet:  true,
	}, func(ctx context.Context) (int, error) {
		return 42, nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != 42 {
		t.Fatalf("expected 42, got %d", result)
	}
	if buf.Len() != 0 {
		t.Fatalf("quiet mode should produce no output, got %q", buf.String())
	}
}

// TestRunWithStatus_NoTitle tests RunWithStatus without title
func TestRunWithStatus_NoTitle(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := context.Background()

	result, err := RunWithStatus[bool](ctx, StatusConfig{
		Title:  "",
		Writer: buf,
		Quiet:  false,
	}, func(ctx context.Context) (bool, error) {
		return true, nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result {
		t.Fatalf("expected true, got false")
	}
}

// TestDisplayOperation_Phase tests Phase method
func TestDisplayOperation_Phase(t *testing.T) {
	buf := &bytes.Buffer{}
	op := NewOperation(StatusConfig{Title: "Test", Writer: buf, Quiet: false})
	dop := op.(*displayOperation)

	dop.Phase("Starting")
	if buf.Len() == 0 {
		t.Fatalf("Phase should produce output")
	}
}

// TestDisplayOperation_Info tests Info method
func TestDisplayOperation_Info(t *testing.T) {
	buf := &bytes.Buffer{}
	op := NewOperation(StatusConfig{Title: "Test", Writer: buf, Quiet: false})
	dop := op.(*displayOperation)

	dop.Info("Progress: %d", 50)
	if buf.Len() == 0 {
		t.Fatalf("Info should produce output")
	}
}
