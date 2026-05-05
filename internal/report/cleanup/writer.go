// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package cleanup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Korrnals/gotr/internal/report"
)

// PDFRenderer is implemented by anything that can produce a PDF artefact
// for a Report. It exists so the writer can defer to the existing
// internal/report/pdf package without an import cycle.
type PDFRenderer interface {
	Save(r *Report, outputPath string) error
}

// WriteOptions controls which formats are written.
type WriteOptions struct {
	Markdown bool
	JSON     bool
	CSV      bool
	PDF      bool

	// PDF is rendered via this renderer when non-nil.
	PDFRenderer PDFRenderer
}

// AllFormats returns a WriteOptions that selects every format. The PDF
// renderer must still be supplied to actually write the .pdf file.
func AllFormats() WriteOptions {
	return WriteOptions{Markdown: true, JSON: true, CSV: true, PDF: true}
}

// WriteResult is the set of artefacts produced by Write.
type WriteResult struct {
	MarkdownPath string
	JSONPath     string
	CSVPath      string
	PDFPath      string
}

// Files returns all non-empty paths from the result, in deterministic
// order (md, json, csv, pdf).
func (r WriteResult) Files() []string {
	var out []string
	for _, p := range []string{r.MarkdownPath, r.JSONPath, r.CSVPath, r.PDFPath} {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Write persists the report under
//
//	<reportsDir>/cleanup-attachments/<label>/<YYYY-MM>/cleanup-<id>-<ts>.<ext>
//
// using the existing report.ClassifyReportWithLabel hierarchy. INDEX.md
// is refreshed via report.Reindex on success. Failures during PDF
// rendering are reported but do not roll back already-written formats.
func Write(ctx context.Context, reportsDir string, rep *Report, opts WriteOptions) (WriteResult, error) {
	if rep == nil {
		return WriteResult{}, fmt.Errorf("cleanup-report: nil report")
	}
	if reportsDir == "" {
		return WriteResult{}, fmt.Errorf("cleanup-report: empty reports dir")
	}
	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		return WriteResult{}, fmt.Errorf("cleanup-report: create reports dir: %w", err)
	}

	base := buildBaseFilename(rep)
	cls := report.ClassifyReportWithLabel(base+".md", rep.Label)
	dir := filepath.Join(reportsDir, cls.RelDir())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return WriteResult{}, fmt.Errorf("cleanup-report: create dir: %w", err)
	}

	var res WriteResult

	if opts.Markdown {
		path := filepath.Join(dir, base+".md")
		if err := os.WriteFile(path, []byte(RenderMarkdown(rep)), 0o644); err != nil {
			return res, fmt.Errorf("cleanup-report: write md: %w", err)
		}
		res.MarkdownPath = path
	}

	if opts.JSON {
		b, err := RenderJSON(rep)
		if err != nil {
			return res, err
		}
		path := filepath.Join(dir, base+".json")
		if err := os.WriteFile(path, b, 0o644); err != nil {
			return res, fmt.Errorf("cleanup-report: write json: %w", err)
		}
		res.JSONPath = path
	}

	if opts.CSV {
		b, err := RenderCSV(rep)
		if err != nil {
			return res, err
		}
		path := filepath.Join(dir, base+".csv")
		if err := os.WriteFile(path, b, 0o644); err != nil {
			return res, fmt.Errorf("cleanup-report: write csv: %w", err)
		}
		res.CSVPath = path
	}

	if opts.PDF && opts.PDFRenderer != nil {
		path := filepath.Join(dir, base+".pdf")
		if err := opts.PDFRenderer.Save(rep, path); err != nil {
			return res, fmt.Errorf("cleanup-report: write pdf: %w", err)
		}
		res.PDFPath = path
	}

	if err := report.Reindex(reportsDir); err != nil {
		// Index is a best-effort artefact; report it but don't fail.
		fmt.Fprintf(os.Stderr, "warning: cleanup-report: refresh INDEX.md: %v\n", err)
	}
	_ = ctx

	return res, nil
}

// buildBaseFilename returns the timestamped basename (no extension)
// shared by every emitted format.
//
//	cleanup-<sanitized-snapshot-id>-<RFC3339-compact>
//
// When SnapshotID is empty (dry-run / --no-snapshot) the literal
// "no_snapshot" segment is used so the file still classifies cleanly
// under the cleanup-attachments hierarchy.
func buildBaseFilename(rep *Report) string {
	id := sanitizeFilenameSegment(rep.SnapshotID)
	if id == "" {
		id = "no_snapshot"
	}
	return fmt.Sprintf("cleanup-attachments-%s-%s",
		rep.Timestamp.UTC().Format("20060102T150405Z"),
		id,
	)
}

var unsafeFilenameChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func sanitizeFilenameSegment(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return unsafeFilenameChars.ReplaceAllString(s, "_")
}
