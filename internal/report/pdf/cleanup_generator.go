// Copyright (c) 2026 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package pdf

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

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
	withLandscape(pdf, func() {
		writeCleanupEntityBreakdown(pdf, r)
		writeCleanupItems(pdf, r)
	})
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
	pdf.SetTextColor(30, 40, 90)
	pdf.CellFormat(contentWidth, 10, title, "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)

	// Subtitle with two badges: Snapshot ID and timestamp.
	pdf.SetFont(FontFamily, "B", 9)
	pdf.SetTextColor(80, 80, 100)
	pdf.CellFormat(20, lineHeight, "Snapshot:", "", 0, "L", false, 0, "")
	pdf.SetFont(FontFamily, "", 9)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(45, lineHeight, nonEmpty(r.SnapshotID, "\u2014"), "", 0, "L", false, 0, "")
	pdf.SetFont(FontFamily, "B", 9)
	pdf.SetTextColor(80, 80, 100)
	pdf.CellFormat(22, lineHeight, "Timestamp:", "", 0, "L", false, 0, "")
	pdf.SetFont(FontFamily, "", 9)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(contentWidth-87, lineHeight, r.Timestamp.UTC().Format(time.RFC3339), "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
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
	if r.Summary.BackupSkipped > 0 {
		rows = append(rows, [2]string{"Skipped (ghost)", fmt.Sprintf("%d", r.Summary.BackupSkipped)})
	}
	renderKeyValue(pdf, rows)
}

func writeCleanupProjects(pdf *fpdf.Fpdf, r *cleanupreport.Report) {
	if len(r.Projects) == 0 {
		return
	}
	sectionHeader(pdf, "Per-project breakdown")

	cols := []float64{20, 70, 20, 30, 40}
	headers := []string{"Project", "Name", "Count", "Total Size", "Oldest"}
	pdfTableHeader(pdf, cols, headers)
	pdf.SetFont(FontFamily, "", 9)
	for _, p := range r.Projects {
		oldest := "—"
		if p.OldestUnix > 0 {
			oldest = time.Unix(p.OldestUnix, 0).UTC().Format("2006-01-02")
		}
		row := []string{
			fmt.Sprintf("%d", p.ProjectID),
			p.ProjectName,
			fmt.Sprintf("%d", p.Count),
			humanBytes(p.TotalBytes),
			oldest,
		}
		pdfTableRowMulti(pdf, cols, row)
	}
	pdf.Ln(2)
}

func writeCleanupItems(pdf *fpdf.Fpdf, r *cleanupreport.Report) {
	if !cleanupHasItems(r) {
		return
	}
	sectionHeader(pdf, "Deleted attachments")

	// Total width: 267mm (landscape A4 minus margins).
	cols := []float64{18, 45, 22, 70, 25, 28, 35, 24}
	headers := []string{"Proj.", "Project Name", "Att. ID", "Name", "Size", "Parent", "Parent Name", "Created"}
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
			parentName := "—"
			if it.ParentName != "" {
				parentName = it.ParentName
			}
			row := []string{
				fmt.Sprintf("%d", p.ProjectID),
				p.ProjectName,
				fmt.Sprintf("%d", it.AttachmentID),
				it.Name,
				humanBytes(it.Size),
				parent,
				parentName,
				created,
			}
			pdfTableRowMulti(pdf, cols, row)
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
		pdfTableRowMulti(pdf, cols, []string{
			fmt.Sprintf("%d", f.AttachmentID),
			fmt.Sprintf("%d", f.ProjectID),
			f.Error,
		})
	}
	pdf.Ln(2)
}

func writeCleanupReferences(pdf *fpdf.Fpdf, r *cleanupreport.Report) {
	sectionHeader(pdf, "References")
	rows := []kvRow{}
	if r.SnapshotID != "" && !r.DryRun {
		rows = append(rows,
			kvRow{Key: "Snapshot", Value: "~/.gotr/snaps/cleanup-attachments/" + r.SnapshotID, Kind: "path"},
			kvRow{Key: "Rollback", Value: "gotr snap rollback " + r.SnapshotID, Kind: "cmd"},
		)
	} else if r.DryRun {
		rows = append(rows, kvRow{Key: "Snapshot", Value: "— (dry-run, no snapshot taken)"})
	}
	rows = append(rows, kvRow{Key: "Report directory", Value: "~/.gotr/reports/cleanup-attachments", Kind: "path"})
	renderKVTyped(pdf, rows)
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

// withLandscape adds a landscape A4 page, runs fn, then switches back
// to a fresh portrait page so subsequent sections lay out normally.
// Effective content width during fn is `landscapeContentWidth`.
//
// fpdf swaps Wd/Ht when orientation="L", so we pass the portrait
// dimensions and let it rotate.
func withLandscape(pdf *fpdf.Fpdf, fn func()) {
	pdf.AddPageFormat("L", fpdf.SizeType{Wd: 210, Ht: 297})
	fn()
	pdf.AddPageFormat("P", fpdf.SizeType{Wd: 210, Ht: 297})
}

// cell. Each cell's text is split into lines that fit within its
// column width using the current font metrics; explicit "\n" in the
// text is also honored. All cells share the same total height — the
// max of the per-cell line counts — so vertical borders stay aligned.
//
// Manually handles page breaks: if the row doesn't fit on the
// remaining page space, a new page is started before rendering. This
// avoids fpdf's auto-page-break splitting a multi-line row mid-cell
// (which would corrupt the table layout).
func pdfTableRowMulti(pdf *fpdf.Fpdf, cols []float64, row []string) {
	cells := make([][]string, len(row))
	maxLines := 1
	for i, v := range row {
		lines := wrapCellLines(pdf, v, cols[i])
		cells[i] = lines
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}
	rowH := lineHeight * float64(maxLines)

	// Manual page-break: if the row would overflow, start a new page
	// before drawing anything. Preserve the current orientation
	// (portrait vs landscape) so a landscape table doesn't suddenly
	// switch to portrait mid-table and get clipped.
	pageW, pageH := pdf.GetPageSize()
	_, _, _, bottomMargin := pdf.GetMargins()
	pageBreakTrigger := pageH - bottomMargin
	if pdf.GetY()+rowH > pageBreakTrigger {
		if pageW > pageH {
			pdf.AddPageFormat("L", fpdf.SizeType{Wd: 210, Ht: 297})
		} else {
			pdf.AddPage()
		}
	}

	// Disable auto page break while drawing the row so SetXY-driven
	// per-line rendering can't accidentally trigger one mid-row.
	autoBreak, autoMargin := pdf.GetAutoPageBreak()
	pdf.SetAutoPageBreak(false, autoMargin)
	defer pdf.SetAutoPageBreak(autoBreak, autoMargin)

	startX := pdf.GetX()
	startY := pdf.GetY()
	for i, lines := range cells {
		x := startX
		for k := 0; k < i; k++ {
			x += cols[k]
		}
		// Border rectangle for the cell.
		pdf.Rect(x, startY, cols[i], rowH, "D")
		for li := 0; li < maxLines; li++ {
			text := ""
			if li < len(lines) {
				text = lines[li]
			}
			pdf.SetXY(x, startY+float64(li)*lineHeight)
			pdf.CellFormat(cols[i], lineHeight, text, "", 0, "L", false, 0, "")
		}
	}
	pdf.SetXY(startX, startY+rowH)
}

// wrapCellLines splits text so each line fits within the given column
// width (minus a small horizontal padding) at the current font.
// Honors explicit "\n" in the input. Falls back to a single line when
// pdf is nil (defensive — never expected at runtime).
//
// The wrapping is rune-aware (not byte-aware): we never split a UTF-8
// multi-byte sequence in the middle. fpdf's own SplitLines indexes the
// input by bytes and breaks Cyrillic/CJK strings at sub-character byte
// boundaries; passing the resulting fragments back to CellFormat causes
// utf8toutf16 to read past end of string and panic with "index out of
// range". This implementation measures widths via pdf.GetStringWidth on
// rune boundaries so the produced lines are always valid UTF-8.
func wrapCellLines(pdf *fpdf.Fpdf, text string, colWidth float64) []string {
	// Defensive sanitization: replace invalid UTF-8 sequences (e.g.
	// truncated names from upstream) with U+FFFD so fpdf never receives
	// a malformed string.
	text = strings.ToValidUTF8(text, "\uFFFD")
	if text == "" {
		return []string{""}
	}
	if pdf == nil {
		return strings.Split(text, "\n")
	}
	const pad = 2.0 // mm of horizontal padding inside the cell border
	avail := colWidth - pad
	if avail <= 0 {
		avail = colWidth
	}
	var out []string
	for _, paragraph := range strings.Split(text, "\n") {
		if paragraph == "" {
			out = append(out, "")
			continue
		}
		out = append(out, wrapParagraphRunes(pdf, paragraph, avail)...)
	}
	if len(out) == 0 {
		return []string{""}
	}
	return out
}

// wrapParagraphRunes performs rune-safe greedy word wrapping of a single
// paragraph (no embedded "\n"). Words longer than avail are broken at
// rune boundaries.
func wrapParagraphRunes(pdf *fpdf.Fpdf, paragraph string, avail float64) []string {
	if paragraph == "" {
		return []string{""}
	}
	if pdf.GetStringWidth(paragraph) <= avail {
		return []string{paragraph}
	}
	var lines []string
	var current strings.Builder
	currentWidth := 0.0
	flush := func() {
		lines = append(lines, current.String())
		current.Reset()
		currentWidth = 0
	}
	// Tokenise into words preserving single spaces between them; this
	// avoids losing whitespace when a wrap occurs at a word boundary.
	tokens := splitWithSpaces(paragraph)
	for _, tok := range tokens {
		w := pdf.GetStringWidth(tok)
		if currentWidth+w <= avail {
			current.WriteString(tok)
			currentWidth += w
			continue
		}
		// Token doesn't fit on current line.
		if current.Len() > 0 {
			// Drop a trailing space before wrapping for a cleaner break.
			line := strings.TrimRight(current.String(), " ")
			lines = append(lines, line)
			current.Reset()
			currentWidth = 0
			if tok == " " {
				continue
			}
		}
		if w <= avail {
			current.WriteString(tok)
			currentWidth = w
			continue
		}
		// Token alone is wider than avail — break by runes.
		for _, line := range breakRunes(pdf, tok, avail) {
			lw := pdf.GetStringWidth(line)
			if currentWidth+lw <= avail && current.Len() > 0 {
				current.WriteString(line)
				currentWidth += lw
				continue
			}
			if current.Len() > 0 {
				flush()
			}
			lines = append(lines, line)
		}
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

// splitWithSpaces returns the input split into alternating word/space
// tokens (single spaces), preserving the original spacing structure
// closely enough for cell wrapping.
func splitWithSpaces(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	var b strings.Builder
	inSpace := false
	for _, r := range s {
		if r == ' ' || r == '\t' {
			if !inSpace {
				if b.Len() > 0 {
					out = append(out, b.String())
					b.Reset()
				}
				inSpace = true
			}
			b.WriteRune(' ')
			continue
		}
		if inSpace {
			out = append(out, b.String())
			b.Reset()
			inSpace = false
		}
		b.WriteRune(r)
	}
	if b.Len() > 0 {
		out = append(out, b.String())
	}
	return out
}

// breakRunes splits a single token into chunks each fitting within
// avail, breaking at rune boundaries only. Always produces at least
// one chunk; an oversized single rune is emitted on its own line.
func breakRunes(pdf *fpdf.Fpdf, tok string, avail float64) []string {
	var lines []string
	var current strings.Builder
	currentWidth := 0.0
	for _, r := range tok {
		if !utf8.ValidRune(r) {
			r = '\uFFFD'
		}
		rw := pdf.GetStringWidth(string(r))
		if currentWidth+rw > avail && current.Len() > 0 {
			lines = append(lines, current.String())
			current.Reset()
			currentWidth = 0
		}
		current.WriteRune(r)
		currentWidth += rw
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return lines
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
		return fmt.Sprintf("%.2f TB", float64(n)/float64(tb))
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
	// Total width: 267mm (landscape A4 minus margins).
	widths := []float64{75, 20, 20, 20, 30, 22, 20, 22, 38}
	headers := []string{"Project", "case", "run", "plan", "p_entry", "result", "test", "Total", "Size"}
	pdfTableHeader(pdf, widths, headers)
	pdf.SetFont(FontFamily, "", 9)
	totals := make(map[string]int, len(colsKinds))
	var grandTotal int
	var grandBytes int64
	for _, row := range r.EntityBreakdown {
		name := fmt.Sprintf("%s (%d)", row.ProjectName, row.ProjectID)
		cells := []string{name}
		for _, k := range colsKinds {
			n := row.Counts[k]
			totals[k] += n
			cells = append(cells, fmt.Sprintf("%d", n))
		}
		cells = append(cells, fmt.Sprintf("%d", row.Total), humanBytes(row.Bytes))
		pdfTableRowMulti(pdf, widths, cells)
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
		pdfTableRowMulti(pdf, cols, []string{row[0], row[1], row[2]})
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
	rows := []kvRow{}
	if len(a.ReportPaths) > 0 {
		for i, p := range a.ReportPaths {
			label := "Audit report"
			if len(a.ReportPaths) > 1 {
				label = fmt.Sprintf("Audit report #%d", i+1)
			}
			rows = append(rows, kvRow{Key: label, Value: p, Kind: "path"})
		}
	}
	if a.SnapshotPath != "" {
		rows = append(rows, kvRow{Key: "Snapshot directory", Value: a.SnapshotPath, Kind: "path"})
	}
	if a.CheckpointDir != "" {
		rows = append(rows,
			kvRow{Key: "Checkpoint cache", Value: a.CheckpointDir, Kind: "path"},
			kvRow{Key: "Resume command", Value: "gotr attachments cleanup --resume <run-id>", Kind: "cmd"},
		)
	}
	renderKVTyped(pdf, rows)
}
