package pdf

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	"github.com/Korrnals/gotr/internal/paths"
	"github.com/Korrnals/gotr/internal/report"
)

// Layout constants (A4 portrait, 15mm margins).
const (
	pageOrientation = "P"
	pageUnits       = "mm"
	pageSize        = "A4"
	marginLeft      = 15.0
	marginTop       = 15.0
	marginRight     = 15.0
	contentWidth    = 210.0 - marginLeft - marginRight // A4 width 210mm
	lineHeight      = 5.5
)

// Generator renders MigrationReport values to PDF.
type Generator struct{}

// NewGenerator returns a fresh Generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// Render produces a PDF document as a byte slice.
// The caller is responsible for persisting the result.
func (g *Generator) Render(r *report.MigrationReport) ([]byte, error) {
	if r == nil {
		return nil, fmt.Errorf("pdf: nil migration report")
	}
	pdf := fpdf.New(pageOrientation, pageUnits, pageSize, "")
	pdf.SetMargins(marginLeft, marginTop, marginRight)
	pdf.SetAutoPageBreak(true, marginTop)

	pdf.AddUTF8FontFromBytes(FontFamily, "", notoSansRegular)
	pdf.AddUTF8FontFromBytes(FontFamily, "B", notoSansBold)

	pdf.AddPage()

	writeTitle(pdf, r)
	writeConfiguration(pdf, r)
	writeSummary(pdf, r)
	writeSkipped(pdf, r)
	writeRollback(pdf, r)
	writePerformance(pdf, r)
	writeReferences(pdf, r)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf: output: %w", err)
	}
	return buf.Bytes(), nil
}

// Save renders the report and writes it to outputPath.
func (g *Generator) Save(r *report.MigrationReport, outputPath string) error {
	data, err := g.Render(r)
	if err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("pdf: write %s: %w", outputPath, err)
	}
	return nil
}

// --- sections ---

func writeTitle(pdf *fpdf.Fpdf, r *report.MigrationReport) {
	pdf.SetFont(FontFamily, "B", 20)
	pdf.CellFormat(contentWidth, 10, "Migration Report", "", 1, "L", false, 0, "")

	pdf.SetFont(FontFamily, "", 10)
	status := strings.ToUpper(r.Status)
	if status == "" {
		status = "UNKNOWN"
	}
	subtitle := fmt.Sprintf("Snapshot: %s   |   %s   |   Status: %s",
		nonEmpty(r.SnapshotID, "n/a"),
		r.Timestamp.Format(time.RFC3339),
		status,
	)
	pdf.CellFormat(contentWidth, lineHeight, subtitle, "", 1, "L", false, 0, "")
	pdf.Ln(4)
}

func writeConfiguration(pdf *fpdf.Fpdf, r *report.MigrationReport) {
	sectionHeader(pdf, "Configuration")
	rows := [][2]string{
		{"Source project", fmt.Sprintf("%d", r.SourceProjectID)},
		{"Target project", fmt.Sprintf("%d", r.TargetProjectID)},
		{"Migration type", nonEmpty(r.MigrationType, "n/a")},
		{"User", nonEmpty(r.User, "n/a")},
		{"Report ID", nonEmpty(r.ID, "n/a")},
	}
	renderKeyValue(pdf, rows)
}

func writeSummary(pdf *fpdf.Fpdf, r *report.MigrationReport) {
	sectionHeader(pdf, "Summary")
	if len(r.Summary) == 0 {
		pdf.SetFont(FontFamily, "", 10)
		pdf.CellFormat(contentWidth, lineHeight, "No resource statistics recorded.", "", 1, "L", false, 0, "")
		pdf.Ln(2)
		return
	}

	headers := []string{"Resource", "Source", "Created", "Updated", "Skipped", "Failed"}
	widths := []float64{54, 25, 25, 25, 25, 26}

	pdf.SetFont(FontFamily, "B", 10)
	for i, h := range headers {
		pdf.CellFormat(widths[i], 7, h, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont(FontFamily, "", 10)
	keys := sortedKeys(r.Summary)
	var totalSrc, totalCreated, totalUpdated, totalSkipped, totalFailed int64
	for _, k := range keys {
		s := r.Summary[k]
		values := []string{
			k,
			fmt.Sprintf("%d", s.SourceCount),
			fmt.Sprintf("%d", s.Created),
			fmt.Sprintf("%d", s.Updated),
			fmt.Sprintf("%d", s.Skipped),
			fmt.Sprintf("%d", s.Failed),
		}
		aligns := []string{"L", "R", "R", "R", "R", "R"}
		for i, v := range values {
			pdf.CellFormat(widths[i], 6.5, v, "1", 0, aligns[i], false, 0, "")
		}
		pdf.Ln(-1)
		totalSrc += s.SourceCount
		totalCreated += s.Created
		totalUpdated += s.Updated
		totalSkipped += s.Skipped
		totalFailed += s.Failed
	}

	pdf.SetFont(FontFamily, "B", 10)
	totals := []string{
		"TOTAL",
		fmt.Sprintf("%d", totalSrc),
		fmt.Sprintf("%d", totalCreated),
		fmt.Sprintf("%d", totalUpdated),
		fmt.Sprintf("%d", totalSkipped),
		fmt.Sprintf("%d", totalFailed),
	}
	aligns := []string{"L", "R", "R", "R", "R", "R"}
	for i, v := range totals {
		pdf.CellFormat(widths[i], 7, v, "1", 0, aligns[i], false, 0, "")
	}
	pdf.Ln(-1)
	pdf.Ln(3)
}

func writeSkipped(pdf *fpdf.Fpdf, r *report.MigrationReport) {
	if len(r.Skipped) == 0 {
		return
	}
	sectionHeader(pdf, "Skipped Resources")

	keys := make([]string, 0, len(r.Skipped))
	for k := range r.Skipped {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pdf.SetFont(FontFamily, "", 10)
	for _, k := range keys {
		reasons := r.Skipped[k]
		if len(reasons) == 0 {
			continue
		}
		pdf.SetFont(FontFamily, "B", 11)
		pdf.CellFormat(contentWidth, lineHeight+1, fmt.Sprintf("%s — %d entries", k, len(reasons)), "", 1, "L", false, 0, "")

		groups := make(map[string]int)
		for _, reason := range reasons {
			key := reason.Reason
			if key == "" {
				key = "unspecified"
			}
			groups[key]++
		}
		grpKeys := make([]string, 0, len(groups))
		for g := range groups {
			grpKeys = append(grpKeys, g)
		}
		sort.Strings(grpKeys)

		pdf.SetFont(FontFamily, "", 10)
		for _, g := range grpKeys {
			pdf.CellFormat(contentWidth, lineHeight, fmt.Sprintf("  - %s: %d", g, groups[g]), "", 1, "L", false, 0, "")
		}
		pdf.Ln(1)
	}
}

func writeRollback(pdf *fpdf.Fpdf, r *report.MigrationReport) {
	if r.Rollback.SnapshotID == "" {
		return
	}
	sectionHeader(pdf, "Rollback")
	rows := [][2]string{
		{"Snapshot ID", r.Rollback.SnapshotID},
		{"Enabled", fmt.Sprintf("%v", r.Rollback.Enabled)},
		{"Deletion order", strings.Join(r.Rollback.DeleteOrder, " -> ")},
		{"Command", nonEmpty(r.Rollback.Command, "n/a")},
	}
	renderKeyValue(pdf, rows)
}

func writePerformance(pdf *fpdf.Fpdf, r *report.MigrationReport) {
	if r.Duration <= 0 && r.Performance.EntitiesPerSec == 0 {
		return
	}
	sectionHeader(pdf, "Performance")
	rows := [][2]string{
		{"Total time", fmt.Sprintf("%.1fs", r.Duration.Seconds())},
		{"Entities/sec", fmt.Sprintf("%.1f", r.Performance.EntitiesPerSec)},
	}
	if r.Performance.PeakMemoryMB > 0 {
		rows = append(rows, [2]string{"Peak memory", fmt.Sprintf("%d MB", r.Performance.PeakMemoryMB)})
	}
	renderKeyValue(pdf, rows)
}

func writeReferences(pdf *fpdf.Fpdf, r *report.MigrationReport) {
	sectionHeader(pdf, "References")

	pdf.SetFont(FontFamily, "", 10)
	if r.SnapshotID != "" {
		snapPath := "~/.gotr/snaps/" + r.SnapshotID
		// Best-effort: if an absolute path was somehow stored, collapse it to ~.
		if strings.HasPrefix(r.SnapshotID, "/") {
			snapPath = paths.RelToHome(r.SnapshotID)
		}
		pdf.CellFormat(contentWidth, lineHeight, "Snapshot: "+snapPath, "", 1, "L", false, 0, "")
	}
	pdf.CellFormat(contentWidth, lineHeight, "Report directory: ~/.gotr/reports", "", 1, "L", false, 0, "")
	pdf.CellFormat(contentWidth, lineHeight, "Rollback: gotr snap rollback "+nonEmpty(r.SnapshotID, "<snapshot-id>"), "", 1, "L", false, 0, "")
}

// --- helpers ---

func sectionHeader(pdf *fpdf.Fpdf, title string) {
	pdf.SetFont(FontFamily, "B", 14)
	pdf.CellFormat(contentWidth, 8, title, "", 1, "L", false, 0, "")
	pdf.SetDrawColor(180, 180, 180)
	y := pdf.GetY()
	pdf.Line(marginLeft, y, marginLeft+contentWidth, y)
	pdf.Ln(2)
}

func renderKeyValue(pdf *fpdf.Fpdf, rows [][2]string) {
	keyWidth := 55.0
	valWidth := contentWidth - keyWidth
	pdf.SetFont(FontFamily, "", 10)
	for _, row := range rows {
		pdf.SetFont(FontFamily, "B", 10)
		pdf.CellFormat(keyWidth, lineHeight, row[0], "", 0, "L", false, 0, "")
		pdf.SetFont(FontFamily, "", 10)
		pdf.CellFormat(valWidth, lineHeight, row[1], "", 1, "L", false, 0, "")
	}
	pdf.Ln(2)
}

func sortedKeys(m map[string]*report.ResourceStats) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func nonEmpty(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
