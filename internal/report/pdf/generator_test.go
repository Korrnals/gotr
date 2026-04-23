package pdf

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Korrnals/gotr/internal/report"
)

const pdfMagic = "%PDF-"

func buildSampleReport() *report.MigrationReport {
	r := report.NewMigrationReport("snap-20260424T010000Z-test", 30, 34, "sync_full", "tester")
	r.AddResourceStats("cases", 1684, 1684, 0, 0, 0)
	r.AddResourceStats("sections", 12, 12, 0, 0, 0)
	r.AddResourceStats("shared_steps", 0, 0, 0, 0, 0)
	r.AddSkipped("cases", []report.SkipReason{
		{ID: 1, Reason: "duplicate", Detail: "same title"},
		{ID: 2, Reason: "duplicate", Detail: "same title"},
		{ID: 3, Reason: "unsupported", Detail: "шаг содержит вложение"}, // exercise UTF-8
	})
	r.SetRollbackInfo("snap-20260424T010000Z-test", true, []string{"case", "section", "shared_step", "suite"})
	r.SetPerformance(22*time.Minute+time.Second, 1684, 0)
	r.MarkSuccess()
	return r
}

func TestGenerator_Render_ProducesPDFBytes(t *testing.T) {
	r := buildSampleReport()
	data, err := NewGenerator().Render(r)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if len(data) < 1024 {
		t.Fatalf("rendered PDF looks too small: %d bytes", len(data))
	}
	if !bytes.HasPrefix(data, []byte(pdfMagic)) {
		t.Fatalf("output does not start with %q: %q", pdfMagic, data[:min(16, len(data))])
	}
}

func TestGenerator_Render_NilReport(t *testing.T) {
	if _, err := NewGenerator().Render(nil); err == nil {
		t.Fatal("expected error for nil report, got nil")
	}
}

func TestGenerator_Render_EmptyReport(t *testing.T) {
	r := report.NewMigrationReport("", 0, 0, "", "")
	data, err := NewGenerator().Render(r)
	if err != nil {
		t.Fatalf("Render empty failed: %v", err)
	}
	if !bytes.HasPrefix(data, []byte(pdfMagic)) {
		t.Fatalf("output does not start with %q", pdfMagic)
	}
}

func TestGenerator_Save_WritesFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "report.pdf")
	if err := NewGenerator().Save(buildSampleReport(), outPath); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() < 1024 {
		t.Fatalf("written PDF too small: %d bytes", info.Size())
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.HasPrefix(data, []byte(pdfMagic)) {
		t.Fatal("file does not start with PDF magic")
	}
}

func TestGenerator_Save_WriteError(t *testing.T) {
	// Target path inside a non-existent directory triggers a write error.
	err := NewGenerator().Save(buildSampleReport(), filepath.Join(t.TempDir(), "nope", "x.pdf"))
	if err == nil {
		t.Fatal("expected write error, got nil")
	}
}

func TestNonEmpty(t *testing.T) {
	if got := nonEmpty("", "fallback"); got != "fallback" {
		t.Fatalf("nonEmpty(empty) = %q, want fallback", got)
	}
	if got := nonEmpty("value", "fallback"); got != "value" {
		t.Fatalf("nonEmpty(value) = %q, want value", got)
	}
}
