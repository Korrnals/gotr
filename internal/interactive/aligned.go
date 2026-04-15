package interactive

import (
	"fmt"
	"strings"
)

// Column defines a single column in an aligned label table.
type Column struct {
	Header string
	// MinWidth overrides the minimum width. If 0, uses header length.
	MinWidth int
}

// AlignedLabels formats rows into equal-width columns separated by │.
// Returns a header string and formatted label strings for each row.
// Each row must have the same number of elements as columns.
func AlignedLabels(columns []Column, rows [][]string) (header string, labels []string) {
	if len(columns) == 0 || len(rows) == 0 {
		return "", nil
	}

	n := len(columns)

	// Compute max widths.
	widths := make([]int, n)
	for i, c := range columns {
		widths[i] = len(c.Header)
		if c.MinWidth > widths[i] {
			widths[i] = c.MinWidth
		}
	}
	for _, row := range rows {
		for i := 0; i < n && i < len(row); i++ {
			if len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}

	// Format function.
	fmtRow := func(cells []string) string {
		var b strings.Builder
		for i := 0; i < n; i++ {
			cell := ""
			if i < len(cells) {
				cell = cells[i]
			}
			if i > 0 {
				b.WriteString(" │ ")
			}
			fmt.Fprintf(&b, "%-*s", widths[i], cell)
		}
		return b.String()
	}

	// Header.
	hdrs := make([]string, n)
	for i, c := range columns {
		hdrs[i] = c.Header
	}
	header = fmtRow(hdrs)

	// Data rows.
	labels = make([]string, len(rows))
	for i, row := range rows {
		labels[i] = fmtRow(row)
	}

	return header, labels
}
