// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package cleanup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Korrnals/gotr/internal/report"
)

// PDFRenderer is implemented by anything that can produce a PDF artifact
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

// WriteResult is the set of artifacts produced by Write.
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

// PredictPaths returns the absolute paths the writer would produce for
// rep when called with the given reportsDir and opts. The directory is
// not created. Useful for self-referencing the output paths inside the
// report itself (e.g. the "Artifacts" section).
func PredictPaths(reportsDir string, rep *Report, opts WriteOptions) []string {
	if rep == nil || reportsDir == "" {
		return nil
	}
	base := buildBaseFilename(rep)
	cls := report.ClassifyReportWithSubdir(base+".md", rep.Label, entitySubdir(rep.Filters.EntityTypes))
	dir := filepath.Join(reportsDir, cls.RelDir())
	var out []string
	if opts.Markdown {
		out = append(out, filepath.Join(dir, base+".md"))
	}
	if opts.JSON {
		out = append(out, filepath.Join(dir, base+".json"))
	}
	if opts.CSV {
		out = append(out, filepath.Join(dir, base+".csv"))
	}
	if opts.PDF && opts.PDFRenderer != nil {
		out = append(out, filepath.Join(dir, base+".pdf"))
	}
	return out
}

// Write persists the report under
//
//	<reportsDir>/cleanup-attachments/<entity-group>/<label>/<YYYY-MM-DD>/cleanup-<id>-<ts>.<ext>
//
// where <entity-group> is derived from rep.EntityTypes (single type → its
// name, multiple → "group-entity", empty → "all-entity"). This keeps
// reports from different scopes from colliding when the same label is
// reused.
//
// INDEX.md is refreshed via report.Reindex on success. Failures during PDF
// rendering are reported but do not roll back already-written formats.
func Write(ctx context.Context, reportsDir string, rep *Report, opts WriteOptions) (WriteResult, error) {
	_ = ctx
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
	cls := report.ClassifyReportWithSubdir(base+".md", rep.Label, entitySubdir(rep.Filters.EntityTypes))
	dir := filepath.Join(reportsDir, cls.RelDir())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return WriteResult{}, fmt.Errorf("cleanup-report: create dir: %w", err)
	}

	var res WriteResult
	if err := writeFormats(dir, base, rep, opts, &res); err != nil {
		return res, err
	}

	if err := report.Reindex(reportsDir); err != nil {
		// Index is a best-effort artifact; report it but don't fail.
		fmt.Fprintf(os.Stderr, "warning: cleanup-report: refresh INDEX.md: %v\n", err)
	}
	return res, nil
}

// writeFormats emits each requested format (Markdown, JSON, CSV, PDF)
// into dir using base as the filename stem, populating the matching
// fields of res. Formats are written in deterministic order; the first
// failure short-circuits and any already-written files are left in
// place for inspection.
func writeFormats(dir, base string, rep *Report, opts WriteOptions, res *WriteResult) error {
	// Markdown — human-readable rendition (always cheapest, written first).
	if opts.Markdown {
		path := filepath.Join(dir, base+".md")
		if err := os.WriteFile(path, []byte(RenderMarkdown(rep)), 0o644); err != nil {
			return fmt.Errorf("cleanup-report: write md: %w", err)
		}
		res.MarkdownPath = path
	}
	// JSON — machine-consumable rendition for audit pipelines.
	if opts.JSON {
		b, err := RenderJSON(rep)
		if err != nil {
			return err
		}
		path := filepath.Join(dir, base+".json")
		if err := os.WriteFile(path, b, 0o644); err != nil {
			return fmt.Errorf("cleanup-report: write json: %w", err)
		}
		res.JSONPath = path
	}
	// CSV — flat row-per-attachment view for spreadsheet review.
	if opts.CSV {
		b, err := RenderCSV(rep)
		if err != nil {
			return err
		}
		path := filepath.Join(dir, base+".csv")
		if err := os.WriteFile(path, b, 0o644); err != nil {
			return fmt.Errorf("cleanup-report: write csv: %w", err)
		}
		res.CSVPath = path
	}
	// PDF — rendered last via the supplied PDFRenderer (if any), since
	// the PDF generator pulls in fonts and is the slowest format.
	if opts.PDF && opts.PDFRenderer != nil {
		path := filepath.Join(dir, base+".pdf")
		if err := opts.PDFRenderer.Save(rep, path); err != nil {
			return fmt.Errorf("cleanup-report: write pdf: %w", err)
		}
		res.PDFPath = path
	}
	return nil
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

// entitySubdir returns a directory segment that buckets cleanup reports
// by the entity types they cover. Single-type runs get a clean name
// ("result", "run", "plan", …). Multi-type or "all" runs get
// "group-entity" so they don't collide with single-type reports and
// are easy to spot.
func entitySubdir(types []string) string {
	if len(types) == 0 {
		return "all-entity"
	}
	// Normalize: lowercase, sort, dedupe.
	seen := make(map[string]struct{}, len(types))
	uniq := make([]string, 0, len(types))
	for _, t := range types {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		uniq = append(uniq, t)
	}
	if len(uniq) == 0 {
		return "all-entity"
	}
	if len(uniq) == 1 {
		return uniq[0]
	}
	sort.Strings(uniq)
	return "group-entity"
}
