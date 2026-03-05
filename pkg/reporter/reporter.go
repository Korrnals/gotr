// Package reporter provides a centralized stats/report builder for gotr CLI output.
//
// All compare commands (and any future output) should use this package
// so that format changes happen in ONE place, not across every command file.
//
// Uses go-pretty/v6 for rendering with ANSI-colored width-1 symbols instead
// of emoji. Emoji have unpredictable visual width across terminals (VS Code,
// iTerm, GNOME Terminal all disagree), which breaks table alignment.
// Callers still pass emoji strings as icon hints — the reporter maps them
// to safe colored indicators automatically.
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

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// ---------------------------------------------------------------------------
// ANSI helpers — safe width-1 colored indicators
// ---------------------------------------------------------------------------

// ANSI color codes.
const (
	ansiReset   = "\033[0m"
	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiBlue    = "\033[34m"
	ansiMagenta = "\033[35m"
	ansiCyan    = "\033[36m"
	ansiWhite   = "\033[37m"
	ansiBold    = "\033[1m"
)

// colored wraps a single character in ANSI color.
func colored(color, ch string) string {
	return color + ch + ansiReset
}

// emojiMap maps emoji strings to ANSI-colored width-1 indicators.
// Every value is exactly 1 visible cell wide, so go-pretty can align perfectly.
var emojiMap = map[string]string{
	// Timing
	"⏱️": colored(ansiCyan, "*"),
	"⏱":  colored(ansiCyan, "*"),
	// Counts / totals
	"📦": colored(ansiBlue, "*"),
	"📋": colored(ansiCyan, "*"),
	"📂": colored(ansiBlue, "*"),
	"📄": colored(ansiWhite, "*"),
	"📈": colored(ansiGreen, "*"),
	"📊": colored(ansiYellow, "*"),
	"📃": colored(ansiCyan, "*"),
	// Results
	"🔹": colored(ansiYellow, "*"),
	"🔗": colored(ansiGreen, "*"),
	"✅": colored(ansiGreen, "+"),
	// Warnings / errors
	"⚠️": colored(ansiYellow, "!"),
	"⚠":  colored(ansiYellow, "!"),
	// Retry / recovery
	"🔄": colored(ansiMagenta, "*"),
	"📥": colored(ansiCyan, "*"),
}

// safeIcon converts an emoji icon to a width-1 ANSI-colored indicator.
// Unknown emoji fall back to a white bullet.
func safeIcon(emoji string) string {
	if s, ok := emojiMap[emoji]; ok {
		return s
	}
	return colored(ansiWhite, "*")
}

// ---------------------------------------------------------------------------
// Element types
// ---------------------------------------------------------------------------

type elementType int

const (
	elemSection   elementType = iota // section header (◆ Name)
	elemStat                         // stat line (icon  label: value)
	elemSeparator                    // ├───...───┤
	elemBlank                        // empty row
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

// Report accumulates sections/stats and renders them via go-pretty.
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
// Emoji in value text are automatically replaced with ASCII equivalents.
func (r *Report) Stat(icon, label string, value interface{}) *Report {
	r.elements = append(r.elements, element{
		typ:   elemStat,
		icon:  icon,
		label: label,
		value: StripEmoji(fmt.Sprintf("%v", value)),
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
// Emoji in the formatted value are automatically replaced with ASCII equivalents.
func (r *Report) StatFmt(icon, label, format string, args ...interface{}) *Report {
	r.elements = append(r.elements, element{
		typ:   elemStat,
		icon:  icon,
		label: label,
		value: StripEmoji(fmt.Sprintf(format, args...)),
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

// String renders the entire report to a string via go-pretty.
// Single-column layout with YAML-like indentation:
//   - Section headers at level 0 (bold)
//   - Stat lines indented with "  " prefix and colored icon
//
// All emoji are replaced with safe width-1 symbols to ensure alignment
// in any terminal (VS Code, iTerm, GNOME Terminal, etc.).
func (r *Report) String() string {
	tw := table.NewWriter()

	// Style: rounded corners (╭╮╰╯).
	tw.SetStyle(table.StyleRounded)

	// Title row — no emoji, just colored text.
	tw.SetTitle(fmt.Sprintf("%s%sСТАТИСТИКА: %s%s", ansiBold, ansiCyan, r.title, ansiReset))

	// Center the title.
	tw.Style().Title.Align = text.AlignCenter

	// Single column, left-aligned, min width for readability.
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignLeft, WidthMin: 55},
	})

	for _, e := range r.elements {
		switch e.typ {
		case elemSection:
			tw.AppendSeparator()
			tw.AppendRow(table.Row{
				fmt.Sprintf("%s%s%s", ansiBold, e.label, ansiReset),
			})
		case elemStat:
			tw.AppendRow(table.Row{
				fmt.Sprintf("  %s %s: %s", safeIcon(e.icon), e.label, e.value),
			})
		case elemSeparator:
			tw.AppendSeparator()
		case elemBlank:
			tw.AppendRow(table.Row{""})
		}
	}

	return "\n" + tw.Render() + "\n"
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
		Stat("🔹", fmt.Sprintf("Только в проекте %d", pid1), onlyFirst).
		Stat("🔹", fmt.Sprintf("Только в проекте %d", pid2), onlySecond).
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

// StripEmoji removes common emoji from text, replacing them with ASCII equivalents.
// Use this for values that go into table cells to avoid width issues.
func StripEmoji(s string) string {
	replacements := []struct{ from, to string }{
		{"✅", "(OK)"},
		{"⚠️", "(!)"},
		{"⚠", "(!)"},
		{"❌", "(X)"},
	}
	for _, r := range replacements {
		s = strings.ReplaceAll(s, r.from, r.to)
	}
	return s
}
