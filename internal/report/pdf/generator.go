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
	}
	if r.SourceProjectName != "" {
		rows = append(rows, [2]string{"Source project name", r.SourceProjectName})
	}
	rows = append(rows, [2]string{"Target project", fmt.Sprintf("%d", r.TargetProjectID)})
	if r.TargetProjectName != "" {
		rows = append(rows, [2]string{"Target project name", r.TargetProjectName})
	}
	rows = append(rows,
		[2]string{"Migration type", nonEmpty(r.MigrationType, "n/a")},
		[2]string{"User", nonEmpty(r.User, "n/a")},
		[2]string{"Report ID", nonEmpty(r.ID, "n/a")},
	)
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
	kvRows := make([]kvRow, len(rows))
	for i, r := range rows {
		kvRows[i] = kvRow{Key: r[0], Value: r[1], Kind: detectKVKind(r[1])}
	}
	renderKVTyped(pdf, kvRows)
}

// kvRow is a single key→value row used by the styled renderer.
type kvRow struct {
	Key   string
	Value string
	// Kind drives the value styling. Recognized values:
	//   ""    — plain text (default)
	//   "cmd" — shell command (italic, faint yellow background)
	//   "path"— filesystem path (italic, faint grey background)
	//   "url" — URL (blue text)
	// Any other value is treated as plain text.
	Kind string
}

// renderKVTyped renders a styled key/value block:
//   - zebra-striped row backgrounds for readability,
//   - key column rendered in bold with a slightly darker accent fill,
//   - value column rendered with style hints based on Kind,
//   - long values are wrapped onto multiple lines and the row height
//     grows to fit them.
func renderKVTyped(pdf *fpdf.Fpdf, rows []kvRow) {
	const (
		keyWidth = 58.0
		lineH    = 6.0
		padX     = 2.0
	)
	valWidth := contentWidth - keyWidth

	// Disable auto-page-break so rows don't get split mid-cell.
	prevAuto, prevMargin := pdf.GetAutoPageBreak()
	pdf.SetAutoPageBreak(false, prevMargin)
	defer pdf.SetAutoPageBreak(prevAuto, prevMargin)

	for i, row := range rows {
		// Pre-compute wrapped value lines and row height.
		pdf.SetFont(FontFamily, "", 10)
		valLines := pdf.SplitLines([]byte(row.Value), valWidth-padX*2)
		if len(valLines) == 0 {
			valLines = [][]byte{[]byte("")}
		}
		rowH := float64(len(valLines)) * lineH

		// Manual page-break to keep the whole row intact.
		_, pageH := pdf.GetPageSize()
		_, _, _, bottomMargin := pdf.GetMargins()
		if pdf.GetY()+rowH > pageH-bottomMargin {
			pdf.AddPage()
		}

		startX := pdf.GetX()
		startY := pdf.GetY()

		// Zebra stripe for the entire row (very light grey on every
		// other row).
		if i%2 == 1 {
			pdf.SetFillColor(247, 247, 250)
			pdf.Rect(startX, startY, contentWidth, rowH, "F")
		}

		// Key column accent fill (slightly darker than the zebra).
		pdf.SetFillColor(232, 234, 240)
		pdf.Rect(startX, startY, keyWidth, rowH, "F")

		// Key text — vertically centered when the row is multi-line.
		pdf.SetXY(startX+1.5, startY)
		pdf.SetFont(FontFamily, "B", 10)
		pdf.SetTextColor(60, 60, 80)
		pdf.CellFormat(keyWidth-1.5, rowH, row.Key, "", 0, "L", false, 0, "")

		// Value column background per Kind.
		switch row.Kind {
		case "cmd":
			pdf.SetFillColor(255, 248, 220)
			pdf.Rect(startX+keyWidth, startY, valWidth, rowH, "F")
			pdf.SetTextColor(120, 70, 0)
		case "path":
			pdf.SetFillColor(240, 240, 240)
			pdf.Rect(startX+keyWidth, startY, valWidth, rowH, "F")
			pdf.SetTextColor(70, 70, 70)
		case "url":
			pdf.SetTextColor(20, 70, 180)
		default:
			pdf.SetTextColor(20, 20, 20)
		}
		pdf.SetFont(FontFamily, "", 10)

		// Render each wrapped line of the value.
		for li, ln := range valLines {
			pdf.SetXY(startX+keyWidth+padX, startY+float64(li)*lineH)
			pdf.CellFormat(valWidth-padX*2, lineH, string(ln), "", 0, "L", false, 0, "")
		}

		// Reset state for next row.
		pdf.SetTextColor(0, 0, 0)
		pdf.SetXY(startX, startY+rowH)
	}
	pdf.Ln(2)
}

// detectKVKind heuristically classifies a KV value to pick a style.
// Used by the legacy [][2]string callers.
func detectKVKind(v string) string {
	if v == "" || v == "—" || v == "n/a" {
		return ""
	}
	if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
		return "url"
	}
	if strings.HasPrefix(v, "gotr ") || strings.HasPrefix(v, "$ ") {
		return "cmd"
	}
	if strings.HasPrefix(v, "/") || strings.HasPrefix(v, "~/") || strings.HasPrefix(v, "./") {
		return "path"
	}
	return ""
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
