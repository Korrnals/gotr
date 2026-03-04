// Package reporter provides a centralized, dynamic stats/report builder
// for gotr CLI output.
//
// All compare commands (and any future output) should use this package
// so that format changes happen in ONE place, not across every command file.
//
// Usage (simple resource):
//
//	reporter.CompareStats("suites", pid1, pid2, 5, 3, 42, elapsed).Print()
//
// Usage (dynamic builder):
//
//	reporter.New("cases").
//	    Section("Общая статистика").
//	    Stat("⏱️", "Время выполнения", elapsed).
//	    Stat("📦", "Всего обработано", total).
//	    Section("Проект 30").
//	    Stat("📋", "Сьютов", 15).
//	    Stat("📄", "Кейсов", 8250).
//	    Section("Результат сравнения").
//	    Stat("✅", "Только в П30", 145).
//	    Stat("🔗", "Общих", 12273).
//	    Print()
package reporter

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
)

// boxWidth is the width of the box border (number of ─ characters).
const boxWidth = 62

// ---------------------------------------------------------------------------
// Element types
// ---------------------------------------------------------------------------

type elementType int

const (
	elemSection   elementType = iota // section header (◆ Name)
	elemStat                         // stat line (icon label: value)
	elemSeparator                    // ├───...───┤
	elemBlank                        // │
)

type element struct {
	typ   elementType
	icon  string
	label string
	value string
}

// ---------------------------------------------------------------------------
// Report — dynamic builder
// ---------------------------------------------------------------------------

// Report accumulates sections/stats and renders them as a formatted box.
// All formatting logic is here — callers only supply data.
type Report struct {
	title    string
	elements []element
	w        io.Writer
}

// New creates a new Report with the given title.
func New(title string) *Report {
	return &Report{
		title: title,
		w:     os.Stdout,
	}
}

// Writer sets the output destination (default: os.Stdout).
func (r *Report) Writer(w io.Writer) *Report {
	r.w = w
	return r
}

// Section adds a named section header.
// A blank line is inserted before each section (except the first).
func (r *Report) Section(name string) *Report {
	if len(r.elements) > 0 {
		r.elements = append(r.elements, element{typ: elemBlank})
	}
	r.elements = append(r.elements, element{typ: elemSection, label: name})
	return r
}

// Stat adds a stat line: icon + label + value.
// value is formatted via fmt.Sprintf("%v", ...).
func (r *Report) Stat(icon, label string, value interface{}) *Report {
	r.elements = append(r.elements, element{
		typ:   elemStat,
		icon:  icon,
		label: label,
		value: fmt.Sprintf("%v", value),
	})
	return r
}

// StatIf adds a stat line only when cond is true.
// Useful for conditional fields (errors, retries, etc.).
func (r *Report) StatIf(cond bool, icon, label string, value interface{}) *Report {
	if cond {
		return r.Stat(icon, label, value)
	}
	return r
}

// StatFmt adds a stat line with a pre-formatted value string.
func (r *Report) StatFmt(icon, label, format string, args ...interface{}) *Report {
	r.elements = append(r.elements, element{
		typ:   elemStat,
		icon:  icon,
		label: label,
		value: fmt.Sprintf(format, args...),
	})
	return r
}

// Blank adds an empty visual line inside the box.
func (r *Report) Blank() *Report {
	r.elements = append(r.elements, element{typ: elemBlank})
	return r
}

// Separator adds a horizontal separator line (├───┤).
func (r *Report) Separator() *Report {
	r.elements = append(r.elements, element{typ: elemSeparator})
	return r
}

// Print renders the report to the configured writer.
func (r *Report) Print() {
	fmt.Fprint(r.w, r.String())
}

// padRight pads content to fill boxWidth using visual (display) width,
// then appends the closing │. Emoji and CJK characters are handled correctly.
func padRight(content string) string {
	visualWidth := runewidth.StringWidth(content)
	if visualWidth >= boxWidth {
		return content + "│"
	}
	return content + strings.Repeat(" ", boxWidth-visualWidth) + "│"
}

// String renders the entire report to a string.
func (r *Report) String() string {
	border := strings.Repeat("─", boxWidth)

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("┌" + border + "┐\n")
	b.WriteString(padRight(fmt.Sprintf("│          📊 СТАТИСТИКА: %s", r.title)) + "\n")
	b.WriteString("├" + border + "┤\n")

	for _, e := range r.elements {
		switch e.typ {
		case elemSection:
			b.WriteString(padRight(fmt.Sprintf("│  ◆ %s", e.label)) + "\n")
		case elemStat:
			b.WriteString(padRight(fmt.Sprintf("│  %s %s: %s", e.icon, e.label, e.value)) + "\n")
		case elemSeparator:
			b.WriteString("├" + border + "┤\n")
		case elemBlank:
			b.WriteString(padRight("│") + "\n")
		}
	}

	b.WriteString("└" + border + "┘\n")
	return b.String()
}

// ---------------------------------------------------------------------------
// Convenience constructors — one-call reports for common patterns
// ---------------------------------------------------------------------------

// CompareStats builds a standard compare stats report for simple resources
// (suites, sections, sharedsteps, runs, plans, milestones, etc.).
// Returns the Report so callers can .Print() or extend it further.
func CompareStats(resource string, pid1, pid2 int64, onlyFirst, onlySecond, common int, elapsed time.Duration) *Report {
	total := onlyFirst + onlySecond + common
	return New(resource).
		Section("Общая статистика").
		Stat("⏱️", "Время выполнения", elapsed.Round(time.Millisecond)).
		Stat("📦", "Всего обработано", total).
		Section("Результат сравнения").
		Stat("✅", fmt.Sprintf("Только в проекте %d", pid1), onlyFirst).
		Stat("✅", fmt.Sprintf("Только в проекте %d", pid2), onlySecond).
		Stat("🔗", "Общих", common)
}

// CompareResultShort returns a single-line summary string.
func CompareResultShort(resource string, pid1, pid2 int64, onlyFirst, onlySecond, common int) string {
	return fmt.Sprintf("%s: П%d=%d, П%d=%d, общих=%d",
		resource, pid1, onlyFirst, pid2, onlySecond, common)
}

// FormatDuration formats a duration for human-readable display.
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dмс", d.Milliseconds())
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if m > 0 {
		return fmt.Sprintf("%dм%02dс", m, s)
	}
	return fmt.Sprintf("%dс", s)
}
