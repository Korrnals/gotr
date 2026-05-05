// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package cleanup

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakePDFRenderer struct {
	called bool
	fail   bool
}

func (f *fakePDFRenderer) Save(_ *Report, outputPath string) error {
	f.called = true
	if f.fail {
		return errors.New("synthetic pdf failure")
	}
	return os.WriteFile(outputPath, []byte("%PDF-1.4 stub"), 0o644)
}

func TestWrite_AllFormats(t *testing.T) {
	dir := t.TempDir()
	rep := fixedReport(t)
	pdf := &fakePDFRenderer{}

	out, err := Write(context.Background(), dir, rep, WriteOptions{
		Markdown: true, JSON: true, CSV: true, PDF: true, PDFRenderer: pdf,
	})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !pdf.called {
		t.Fatal("PDFRenderer.Save was not called")
	}

	for _, p := range out.Files() {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected file at %s: %v", p, err)
		}
		if !strings.Contains(p, filepath.Join("cleanup-attachments", "audit-2026-05", "2026-05")) {
			t.Errorf("file outside expected hierarchy: %s", p)
		}
	}

	if filepath.Ext(out.MarkdownPath) != ".md" ||
		filepath.Ext(out.JSONPath) != ".json" ||
		filepath.Ext(out.CSVPath) != ".csv" ||
		filepath.Ext(out.PDFPath) != ".pdf" {
		t.Errorf("unexpected extensions: %+v", out)
	}

	// INDEX.md is regenerated.
	idx := filepath.Join(dir, "INDEX.md")
	b, err := os.ReadFile(idx)
	if err != nil {
		t.Fatalf("read INDEX.md: %v", err)
	}
	if !strings.Contains(string(b), "cleanup-attachments") {
		t.Errorf("INDEX.md missing cleanup-attachments category:\n%s", b)
	}
}

func TestWrite_NoSnapshotFallback(t *testing.T) {
	dir := t.TempDir()
	rep := fixedReport(t)
	rep.SnapshotID = ""
	rep.DryRun = true

	out, err := Write(context.Background(), dir, rep, WriteOptions{Markdown: true})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !strings.Contains(out.MarkdownPath, "no_snapshot") {
		t.Errorf("dry-run report should embed no_snapshot in filename: %s", out.MarkdownPath)
	}
}

func TestWrite_RejectsNil(t *testing.T) {
	if _, err := Write(context.Background(), t.TempDir(), nil, AllFormats()); err == nil {
		t.Error("expected error for nil report")
	}
	if _, err := Write(context.Background(), "", fixedReport(t), AllFormats()); err == nil {
		t.Error("expected error for empty reports dir")
	}
}
