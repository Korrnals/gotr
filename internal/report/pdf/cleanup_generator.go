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
	writeCleanupSummary(pdf, r)
	writeCleanupProjects(pdf, r)
	writeCleanupItems(pdf, r)
	writeCleanupFailures(pdf, r)
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
		{"Server", nonEmpty(r.Server, "—")},
		{"gotr version", nonEmpty(r.GotrVer, "—")},
		{"Label", nonEmpty(r.Label, "—")},
		{"User", nonEmpty(r.User, "—")},
		{"Dry-run", boolStr(r.DryRun)},
	}
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
	)
	switch {
	case n >= gb:
		return fmt.Sprintf("%.2f GB", float64(n)/float64(gb))
	case n >= mb:
		return fmt.Sprintf("%.2f MB", float64(n)/float64(mb))
	case n >= kb:
		return fmt.Sprintf("%.2f KB", float64(n)/float64(kb))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
