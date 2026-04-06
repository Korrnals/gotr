package output

import (
	"errors"
	"os"
	"testing"
)

// DryRunPrinter methods write to os.Stderr, which is hard to capture in tests.
// We redirect stderr to /dev/null to exercise the code paths for coverage.

func withDevNull(fn func()) {
	orig := os.Stderr
	devnull, err := os.Open(os.DevNull)
	if err == nil {
		os.Stderr = devnull
		defer func() {
			os.Stderr = orig
			devnull.Close()
		}()
	}
	fn()
}

func TestDryRunPrinter_PrintBatch(t *testing.T) {
	p := NewDryRunPrinter("test-cmd")
	withDevNull(func() {
		// more than 10 items to hit truncation branch
		items := make([]string, 15)
		for i := range items {
			items[i] = "item"
		}
		p.PrintBatch("sync", items)
		// fewer than 10 items
		p.PrintBatch("sync", items[:3])
	})
}

func TestDryRunPrinter_PrintOperation(t *testing.T) {
	p := NewDryRunPrinter("test-cmd")
	withDevNull(func() {
		// nil body path
		p.PrintOperation("create", "POST", "/api/v2/add_case/1", nil)

		// marshal success path
		p.PrintOperation("update", "POST", "/api/v2/update_case/1", map[string]any{"title": "Case"})

		// marshal error path
		p.PrintOperation("bad", "POST", "/api/v2/add_case/1", map[string]any{"fn": func() {}})
	})
}

func TestDryRunPrinter_PrintSimple(t *testing.T) {
	p := NewDryRunPrinter("test-cmd")
	withDevNull(func() {
		p.PrintSimple("delete", "delete run 10")
	})
}

func TestDryRunPrinter_FormatBodyForDisplay(t *testing.T) {
	// nil body
	got := FormatBodyForDisplay(nil)
	if got != "(no body)" {
		t.Errorf("expected '(no body)', got %q", got)
	}
	// valid body
	type S struct {
		X int `json:"x"`
	}
	out := FormatBodyForDisplay(S{X: 42})
	if out == "" {
		t.Error("expected non-empty JSON output")
	}
	// unmarshalable body (func is not JSON-serializable)
	bad := FormatBodyForDisplay(func() {})
	if bad == "" {
		t.Error("expected error string, got empty")
	}
}

func TestDryRunPrinter_PrintSummary(t *testing.T) {
	p := NewDryRunPrinter("cmd")
	withDevNull(func() {
		p.PrintSummary([]string{"create case", "update run"})
		p.PrintSummary(nil)
	})
}

func TestDryRunPrinter_PrintValidationError(t *testing.T) {
	p := NewDryRunPrinter("cmd")
	withDevNull(func() {
		p.PrintValidationError(errors.New("validation failed"))
	})
}
