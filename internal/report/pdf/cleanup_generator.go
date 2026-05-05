// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package pdf

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	cleanupreport "github.com/Korrnals/gotr/internal/report/cleanup"
)

// CleanupGenerator renders cleanup.Report values to PDF using the same
// fonts and layout primitives as the migration generator.
type CleanupGenerator struct{}

// NewCleanupGenerator returns a fresh CleanupGenerator.
func NewCleanupGenerator() *CleanupGenerator {
	return &CleanupGenerator{}
}

// Render produces a PDF document for the given cleanup report.
func (g *CleanupGenerator) Render(r *cleanupreport.Report) ([]byte, error) {
	if r == nil {
		return nil, fmt.Errorf("pdf: nil cleanup report")
	}
	pdf := fpdf.New(pageOrientation, pageUnits, pageSize, "")
	pdf.SetMargins(marginLeft, marginTop, marginRight)
	pdf.SetAutoPageBreak(true, marginTop)

	pdf.AddUTF8FontFromBytes(FontFamily, "", notoSansRegular)
	pdf.AddUTF8FontFromBytes(FontFamily, "B", notoSansBold)

	pdf.AddPage()

	writeCleanupTitle(pdf, r)
	writeCleanupRun(pdf, r)
	writeCleanupFilters(pdf, r)
	writeCleanupChunking(pdf, r)
	writeCleanupSummary(pdf, r)
	writeCleanupProjects(pdf, r)
	writeCleanupEntityBreakdown(pdf, r)
	writeCleanupItems(pdf, r)
	writeCleanupFailures(pdf, r)
	writeCleanupSnapshotArtifacts(pdf, r)
	writeCleanupFilesOnDisk(pdf, r)
	writeCleanupReferences(pdf, r)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf: output: %w", err)
	}
	return buf.Bytes(), nil
}

// Save renders the cleanup report and writes it to outputPath.
func (g *CleanupGenerator) Save(r *cleanupreport.Report, outputPath string) error {
	data, err := g.Render(r)
	if err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("pdf: write %s: %w", outputPath, err)
	}
	return nil
}

func writeCleanupTitle(pdf *fpdf.Fpdf, r *cleanupreport.Report) {
	title := "Attachments Cleanup Report"
	if r.DryRun {
		title += "  (DRY-RUN)"
	}
	pdf.SetFont(FontFamily, "B", 20)
	pdf.CellFormat(contentWidth, 10, title, "", 1, "L", false, 0, "")

	pdf.SetFont(FontFamily, "", 10)
	subtitle := fmt.Sprintf("Snapshot: %s   |   %s",
		nonEmpty(r.SnapshotID, "—"),
		r.Timestamp.UTC().Format(time.RFC3339),
	)
	pdf.CellFormat(contentWidth, lineHeight, subtitle, "", 1, "L", false, 0, "")
	pdf.Ln(4)
}

func writeCleanupRun(pdf *fpdf.Fpdf, r *cleanupreport.Report) {
	sectionHeader(pdf, "Run")
	rows := [][2]string{
		{"Report ID", r.ID},
	}
	if r.RunID != "" {
		rows = append(rows, [2]string{"Run ID", r.RunID})
	}
	rows = append(rows,
		[2]string{"Server", nonEmpty(r.Server, "—")},
		[2]string{"gotr version", nonEmpty(r.GotrVer, "—")},
		[2]string{"Label", nonEmpty(r.Label, "—")},
		[2]string{"User", nonEmpty(r.User, "—")},
		[2]string{"Dry-run", boolStr(r.DryRun)},
	)
	if len(r.CLIArgs) > 0 {
		rows = append(rows, [2]string{"CLI", "gotr " + strings.Join(r.CLIArgs, " ")})
	}
	renderKeyValue(pdf, rows)
}

func writeCleanupFilters(pdf *fpdf.Fpdf, r *cleanupreport.Report) {
	sectionHeader(pdf, "Filters")
	scope := "—"
	switch {
	case r.Filters.AllProjects:
		scope = "all projects"
	case len(r.Filters.ProjectIDs) > 0:
		ids := make([]string, len(r.Filters.ProjectIDs))
		for i, id := range r.Filters.ProjectIDs {
			ids[i] = fmt.Sprintf("%d", id)
		}
		scope = strings.Join(ids, ", ")
	}
	rows := [][2]string{
		{"Scope", scope},
		{"Older than", nonEmpty(r.Filters.OlderThan, "—")},
		{"Entity types", joinOrDash(r.Filters.EntityTypes)},
		{"Scan strategy", nonEmpty(r.Filters.ScanStrategy, "auto")},
	}
	if r.Filters.CutoffUnix > 0 {
		rows = append(rows, [2]string{"Cutoff (UTC)", time.Unix(r.Filters.CutoffUnix, 0).UTC().Format(time.RFC3339)})
	}
	if r.Filters.Limit > 0 {
		rows = append(rows, [2]string{"Limit", fmt.Sprintf("%d", r.Filters.Limit)})
	}
	renderKeyValue(pdf, rows)
}

func writeCleanupSummary(pdf *fpdf.Fpdf, r *cleanupreport.Report) {
	sectionHeader(pdf, "Summary")
	rows := [][2]string{
		{"Total selected", fmt.Sprintf("%d", r.Summary.TotalSelected)},
		{"Backed up", fmt.Sprintf("%d (%s)", r.Summary.BackedUp, humanBytes(r.Summary.BackupBytes))},
		{"Deleted", fmt.Sprintf("%d", r.Summary.Deleted)},
		{"Failed", fmt.Sprintf("%d", r.Summary.Failed)},
		{"Freed on server", humanBytes(r.Summary.FreedBytes)},
	}
	renderKeyValue(pdf, rows)
}

func writeCleanupProjects(pdf *fpdf.Fpdf, r *cleanupreport.Report) {
	if len(r.Projects) == 0 {
		return
	}
	sectionHeader(pdf, "Per-project breakdown")

	cols := []float64{20, 70, 20, 30, 40}
	headers := []string{"Project", "Name", "Count", "Bytes", "Oldest"}
	pdfTableHeader(pdf, cols, headers)
	pdf.SetFont(FontFamily, "", 9)
	for _, p := range r.Projects {
		oldest := "—"
		if p.OldestUnix > 0 {
			oldest = time.Unix(p.OldestUnix, 0).UTC().Format("2006-01-02")
		}
		row := []string{
			fmt.Sprintf("%d", p.ProjectID),
			truncatePDF(p.ProjectName, 50),
			fmt.Sprintf("%d", p.Count),
			humanBytes(p.TotalBytes),
			oldest,
		}
		pdfTableRow(pdf, cols, row)
	}
	pdf.Ln(2)
}

func writeCleanupItems(pdf *fpdf.Fpdf, r *cleanupreport.Report) {
	if !cleanupHasItems(r) {
		return
	}
	sectionHeader(pdf, "Deleted attachments")

	cols := []float64{18, 22, 56, 22, 30, 32}
	headers := []string{"Project", "Att. ID", "Name", "Size", "Parent", "Created"}
	pdfTableHeader(pdf, cols, headers)
	pdf.SetFont(FontFamily, "", 9)

	for _, p := range r.Projects {
		for _, it := range p.Items {
			created := "—"
			if it.CreatedUnix > 0 {
				created = time.Unix(it.CreatedUnix, 0).UTC().Format("2006-01-02")
			}
			parent := "—"
			if it.ParentKind != "" {
				parent = it.ParentKind
				if it.ParentID != "" {
					parent = parent + ":" + it.ParentID
				}
			}
			row := []string{
				fmt.Sprintf("%d", p.ProjectID),
				fmt.Sprintf("%d", it.AttachmentID),
				truncatePDF(it.Name, 38),
				humanBytes(it.Size),
				truncatePDF(parent, 22),
				created,
			}
			pdfTableRow(pdf, cols, row)
		}
	}
	pdf.Ln(2)
}

func writeCleanupFailures(pdf *fpdf.Fpdf, r *cleanupreport.Report) {
	if len(r.Failures) == 0 {
		return
	}
	sectionHeader(pdf, "Failures")

	cols := []float64{30, 30, 120}
	pdfTableHeader(pdf, cols, []string{"Att. ID", "Project", "Error"})
	pdf.SetFont(FontFamily, "", 9)
	for _, f := range r.Failures {
		pdfTableRow(pdf, cols, []string{
			fmt.Sprintf("%d", f.AttachmentID),
			fmt.Sprintf("%d", f.ProjectID),
			truncatePDF(f.Error, 90),
		})
	}
	pdf.Ln(2)
}

func writeCleanupReferences(pdf *fpdf.Fpdf, r *cleanupreport.Report) {
	sectionHeader(pdf, "References")
	pdf.SetFont(FontFamily, "", 10)
	if r.SnapshotID != "" && !r.DryRun {
		pdf.CellFormat(contentWidth, lineHeight, "Snapshot: ~/.gotr/snaps/cleanup-attachments/"+r.SnapshotID, "", 1, "L", false, 0, "")
		pdf.CellFormat(contentWidth, lineHeight, "Rollback: gotr snap rollback "+r.SnapshotID, "", 1, "L", false, 0, "")
	} else if r.DryRun {
		pdf.CellFormat(contentWidth, lineHeight, "No snapshot was taken (dry-run).", "", 1, "L", false, 0, "")
	}
	pdf.CellFormat(contentWidth, lineHeight, "Report directory: ~/.gotr/reports/cleanup-attachments", "", 1, "L", false, 0, "")
}

// --- table primitives ---

func pdfTableHeader(pdf *fpdf.Fpdf, cols []float64, headers []string) {
	pdf.SetFont(FontFamily, "B", 9)
	pdf.SetFillColor(230, 230, 230)
	for i, h := range headers {
		pdf.CellFormat(cols[i], lineHeight+1, h, "1", 0, "L", true, 0, "")
	}
	pdf.Ln(-1)
}

func pdfTableRow(pdf *fpdf.Fpdf, cols []float64, row []string) {
	for i, v := range row {
		pdf.CellFormat(cols[i], lineHeight, v, "1", 0, "L", false, 0, "")
	}
	pdf.Ln(-1)
}

func cleanupHasItems(r *cleanupreport.Report) bool {
	for _, p := range r.Projects {
		if len(p.Items) > 0 {
			return true
		}
	}
	return false
}

func truncatePDF(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return s[:maxLen]
	}
	return s[:maxLen-1] + "…"
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func joinOrDash(xs []string) string {
	if len(xs) == 0 {
		return "—"
	}
	return strings.Join(xs, ", ")
}

func humanBytes(n int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
		tb = 1024 * gb
	)
	switch {
	case n >= tb:
		return fmt.Sprintf("%.2f TB (%d B)", float64(n)/float64(tb), n)
	case n >= gb:
		return fmt.Sprintf("%.2f GB (%d B)", float64(n)/float64(gb), n)
	case n >= mb:
		return fmt.Sprintf("%.2f MB (%d B)", float64(n)/float64(mb), n)
	case n >= kb:
		return fmt.Sprintf("%.2f KB (%d B)", float64(n)/float64(kb), n)
	default:
		return fmt.Sprintf("%d B", n)
	}
}

// writeCleanupChunking renders the chunked-execution & concurrency
// profile of the run. No-op when no ChunkingInfo is attached.
func writeCleanupChunking(pdf *fpdf.Fpdf, r *cleanupreport.Report) {
	if r.Chunking == nil {
		return
	}
	c := r.Chunking
	sectionHeader(pdf, "Execution & concurrency")
	rows := [][2]string{}
	if c.ChunkSize > 0 {
		rows = append(rows, [2]string{"Chunk size", fmt.Sprintf("%d", c.ChunkSize)})
	}
	if c.ChunksTotal > 0 {
		rows = append(rows, [2]string{"Chunks", fmt.Sprintf("%d / %d", c.ChunksCompleted, c.ChunksTotal)})
	}
	if c.ScanTimeoutPerProject != "" {
		rows = append(rows, [2]string{"Scan timeout / project", c.ScanTimeoutPerProject})
	}
	if c.DeleteConcurrency > 0 {
		rows = append(rows, [2]string{"Delete concurrency", fmt.Sprintf("%d", c.DeleteConcurrency)})
	}
	if c.BackupConcurrency > 0 {
		rows = append(rows, [2]string{"Backup concurrency", fmt.Sprintf("%d", c.BackupConcurrency)})
	}
	if c.ResumedFrom != "" {
		rows = append(rows, [2]string{"Resumed from", c.ResumedFrom})
	}
	rows = append(rows,
		[2]string{"Reference scan", boolStr(!c.SkipReferences)},
		[2]string{"Compress binaries", boolStr(c.Compress)},
	)
	renderKeyValue(pdf, rows)
}

// writeCleanupEntityBreakdown renders a per-project × entity-type
// matrix using the canonical column order.
func writeCleanupEntityBreakdown(pdf *fpdf.Fpdf, r *cleanupreport.Report) {
	if len(r.EntityBreakdown) == 0 {
		return
	}
	sectionHeader(pdf, "Per-project × entity-type breakdown")
	colsKinds := []string{"case", "run", "plan", "plan_entry", "result", "test"}
	widths := []float64{45, 14, 14, 14, 22, 16, 14, 16, 25}
	headers := []string{"Project", "case", "run", "plan", "p_entry", "result", "test", "Total", "Bytes"}
	pdfTableHeader(pdf, widths, headers)
	pdf.SetFont(FontFamily, "", 9)
	totals := make(map[string]int, len(colsKinds))
	var grandTotal int
	var grandBytes int64
	for _, row := range r.EntityBreakdown {
		name := truncatePDF(fmt.Sprintf("%s (%d)", row.ProjectName, row.ProjectID), 28)
		cells := []string{name}
		for _, k := range colsKinds {
			n := row.Counts[k]
			totals[k] += n
			cells = append(cells, fmt.Sprintf("%d", n))
		}
		cells = append(cells, fmt.Sprintf("%d", row.Total), humanBytes(row.Bytes))
		pdfTableRow(pdf, widths, cells)
		grandTotal += row.Total
		grandBytes += row.Bytes
	}
	footer := []string{"TOTAL"}
	for _, k := range colsKinds {
		footer = append(footer, fmt.Sprintf("%d", totals[k]))
	}
	footer = append(footer, fmt.Sprintf("%d", grandTotal), humanBytes(grandBytes))
	pdf.SetFont(FontFamily, "B", 9)
	pdfTableRow(pdf, widths, footer)
	pdf.Ln(2)
}

// writeCleanupSnapshotArtifacts renders the snapshot-directory file
// inventory + counters (mapping/integrity/references). No-op on
// dry-run or when no snapshot was produced.
func writeCleanupSnapshotArtifacts(pdf *fpdf.Fpdf, r *cleanupreport.Report) {
	if r.Snapshot == nil || r.SnapshotID == "" || r.DryRun {
		return
	}
	sectionHeader(pdf, "Snapshot artifacts")
	pdf.SetFont(FontFamily, "", 10)
	pdf.MultiCell(contentWidth, lineHeight, "Files written to the snapshot directory and their role:", "", "L", false)
	pdf.Ln(1)
	writeCleanupSnapshotFileTable(pdf, r.Snapshot)
	writeCleanupSnapshotCounters(pdf, r.Snapshot)
}

// writeCleanupSnapshotFileTable renders the file→role→path table.
func writeCleanupSnapshotFileTable(pdf *fpdf.Fpdf, s *cleanupreport.SnapshotInfo) {
	cols := []float64{36, 78, 66}
	pdfTableHeader(pdf, cols, []string{"File", "Role", "Path"})
	pdf.SetFont(FontFamily, "", 8)
	rows := [][3]string{}
	if s.MetaPath != "" {
		rows = append(rows, [3]string{"meta.json", "Snapshot metadata", s.MetaPath})
	}
	if s.MappingPath != "" {
		rows = append(rows, [3]string{
			"attachments.json",
			fmt.Sprintf("v3.6 mapping schema=%d, sha256 per file", s.MappingSchemaVersion),
			s.MappingPath,
		})
	}
	if s.ReferencesPath != "" {
		rows = append(rows, [3]string{"references.json", "Markdown URL refs in entity bodies", s.ReferencesPath})
	}
	if s.IntegrityPath != "" {
		rows = append(rows, [3]string{"integrity.json", "Per-file sha256 + Merkle root", s.IntegrityPath})
	}
	if s.FilesDir != "" {
		rows = append(rows, [3]string{"files/", "Backed-up attachment binaries", s.FilesDir})
	}
	for _, row := range rows {
		pdfTableRow(pdf, cols, []string{
			truncatePDF(row[0], 24),
			truncatePDF(row[1], 52),
			truncatePDF(row[2], 44),
		})
	}
	pdf.Ln(2)
}

// writeCleanupSnapshotCounters renders the key/value counter block.
func writeCleanupSnapshotCounters(pdf *fpdf.Fpdf, s *cleanupreport.SnapshotInfo) {
	rows := [][2]string{}
	if s.MappingTotal > 0 {
		rows = append(rows,
			[2]string{"Mapping entries", fmt.Sprintf("%d", s.MappingTotal)},
			[2]string{"Restorable entries", fmt.Sprintf("%d / %d", s.MappingRestorable, s.MappingTotal)},
		)
	}
	if s.FilesCount > 0 {
		rows = append(rows, [2]string{"Files in files/", fmt.Sprintf("%d (%s)", s.FilesCount, humanBytes(s.FilesBytes))})
	}
	if s.ReferencesSkipped {
		rows = append(rows, [2]string{"Reference scan", "skipped (--skip-references)"})
	} else {
		rows = append(rows,
			[2]string{"Entities with references", fmt.Sprintf("%d", s.EntitiesScanned)},
			[2]string{"Markdown URL refs indexed", fmt.Sprintf("%d", s.RefsIndexed)},
		)
	}
	if s.IntegrityRoot != "" {
		rows = append(rows, [2]string{"Integrity Merkle root", truncatePDF(s.IntegrityRoot, 64)})
	}
	if s.IntegrityFiles > 0 {
		rows = append(rows, [2]string{"Integrity files covered", fmt.Sprintf("%d", s.IntegrityFiles)})
	}
	if len(rows) > 0 {
		renderKeyValue(pdf, rows)
	}
}

// writeCleanupFilesOnDisk renders the artifact-paths inventory: audit
// reports, snapshot dir, and checkpoint cache used by --resume.
func writeCleanupFilesOnDisk(pdf *fpdf.Fpdf, r *cleanupreport.Report) {
	if r.Artifacts == nil {
		return
	}
	a := r.Artifacts
	if len(a.ReportPaths) == 0 && a.SnapshotPath == "" && a.CheckpointDir == "" {
		return
	}
	sectionHeader(pdf, "Files on disk")
	pdf.SetFont(FontFamily, "", 9)
	if len(a.ReportPaths) > 0 {
		pdf.SetFont(FontFamily, "B", 9)
		pdf.CellFormat(contentWidth, lineHeight, "Audit reports:", "", 1, "L", false, 0, "")
		pdf.SetFont(FontFamily, "", 9)
		for _, p := range a.ReportPaths {
			pdf.CellFormat(contentWidth, lineHeight, "  · "+truncatePDF(p, 130), "", 1, "L", false, 0, "")
		}
	}
	if a.SnapshotPath != "" {
		pdf.SetFont(FontFamily, "B", 9)
		pdf.CellFormat(contentWidth, lineHeight, "Snapshot directory:", "", 1, "L", false, 0, "")
		pdf.SetFont(FontFamily, "", 9)
		pdf.CellFormat(contentWidth, lineHeight, "  "+truncatePDF(a.SnapshotPath, 130), "", 1, "L", false, 0, "")
	}
	if a.CheckpointDir != "" {
		pdf.SetFont(FontFamily, "B", 9)
		pdf.CellFormat(contentWidth, lineHeight, "Checkpoint cache (used by --resume):", "", 1, "L", false, 0, "")
		pdf.SetFont(FontFamily, "", 9)
		pdf.CellFormat(contentWidth, lineHeight, "  "+truncatePDF(a.CheckpointDir, 130), "", 1, "L", false, 0, "")
	}
	pdf.Ln(2)
}
