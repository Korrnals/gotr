package reporter

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	r := New("test")
	if r == nil {
		t.Fatal("New returned nil")
	}
	if r.title != "test" {
		t.Errorf("expected title 'test', got %q", r.title)
	}
}

func TestEmptyReport(t *testing.T) {
	var buf bytes.Buffer
	New("empty").Writer(&buf).
		Section("Empty").
		Stat("📦", "Test", 0).
		Print()

	out := buf.String()
	if !strings.Contains(out, "STATS: empty") {
		t.Errorf("missing title in output:\n%s", out)
	}
	// go-pretty uses rounded borders ╭╮╰╯
	if !strings.Contains(out, "╭") || !strings.Contains(out, "╰") {
		t.Error("missing box borders")
	}
}

func TestSectionAndStat(t *testing.T) {
	var buf bytes.Buffer
	New("items").
		Writer(&buf).
		Section("General").
		Stat("📦", "Total", 42).
		Section("Details").
		Stat("✅", "Ok", "yes").
		Print()

	out := buf.String()

	if !strings.Contains(out, "General") {
		t.Errorf("missing section 'General' in:\n%s", out)
	}
	if !strings.Contains(out, "Details") {
		t.Errorf("missing section 'Details' in:\n%s", out)
	}
	if !strings.Contains(out, "Total: 42") {
		t.Errorf("missing stat 'Total: 42' in:\n%s", out)
	}
	if !strings.Contains(out, "Ok: yes") {
		t.Errorf("missing stat 'Ok: yes' in:\n%s", out)
	}
}

func TestStatIf(t *testing.T) {
	var buf bytes.Buffer
	New("cond").
		Writer(&buf).
		StatIf(true, "✅", "Shown", 1).
		StatIf(false, "❌", "Hidden", 2).
		Print()

	out := buf.String()
	if !strings.Contains(out, "Shown") {
		t.Error("StatIf(true) should render")
	}
	if strings.Contains(out, "Hidden") {
		t.Error("StatIf(false) should not render")
	}
}

func TestStatFmt(t *testing.T) {
	var buf bytes.Buffer
	New("fmt").
		Writer(&buf).
		StatFmt("🔄", "Recovered", "%d/%d", 3, 5).
		Print()

	out := buf.String()
	if !strings.Contains(out, "Recovered: 3/5") {
		t.Errorf("StatFmt not rendered correctly:\n%s", out)
	}
}

func TestSeparator(t *testing.T) {
	var buf bytes.Buffer
	New("sep").
		Writer(&buf).
		Stat("📦", "Before", 1).
		Separator().
		Stat("📦", "After", 2).
		Print()

	out := buf.String()
	if !strings.Contains(out, "Before: 1") {
		t.Errorf("missing 'Before' in:\n%s", out)
	}
	if !strings.Contains(out, "After: 2") {
		t.Errorf("missing 'After' in:\n%s", out)
	}
}

func TestBlank(t *testing.T) {
	var buf bytes.Buffer
	New("blank").
		Writer(&buf).
		Stat("📦", "A", 1).
		Blank().
		Stat("📦", "B", 2).
		Print()

	out := buf.String()
	if !strings.Contains(out, "A: 1") {
		t.Errorf("missing 'A: 1' in:\n%s", out)
	}
	if !strings.Contains(out, "B: 2") {
		t.Errorf("missing 'B: 2' in:\n%s", out)
	}
}

func TestCompareStats(t *testing.T) {
	var buf bytes.Buffer
	CompareStats("suites", 30, 31, 5, 3, 42, 2*time.Second).
		Writer(&buf).
		Print()

	out := buf.String()

	checks := []string{
		"STATS: suites",
		"General statistics",
		"Execution time",
		"Total processed: 50",
		"Comparison results",
		"Only in project 30: 5",
		"Only in project 31: 3",
		"Common: 42",
	}

	for _, check := range checks {
		if !strings.Contains(out, check) {
			t.Errorf("missing %q in CompareStats output:\n%s", check, out)
		}
	}
}

func TestCompareResultShort(t *testing.T) {
	s := CompareResultShort("runs", 10, 20, 5, 3, 42)
	expected := "runs: P10=5, P20=3, common=42"
	if s != expected {
		t.Errorf("expected %q, got %q", expected, s)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Millisecond, "500ms"},
		{5 * time.Second, "5s"},
		{90 * time.Second, "1m30s"},
		{3*time.Minute + 5*time.Second, "3m05s"},
	}
	for _, tt := range tests {
		got := FormatDuration(tt.d)
		if got != tt.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestStringMethod(t *testing.T) {
	s := New("test").
		Section("A").
		Stat("📦", "Val", 1).
		String()

	if !strings.Contains(s, "STATS: test") {
		t.Error("String() should return rendered report")
	}
}

func TestChaining(t *testing.T) {
	var buf bytes.Buffer
	r := New("chain").Writer(&buf)
	r2 := r.Section("S")
	r3 := r2.Stat("📦", "V", 1)
	r4 := r3.StatIf(true, "✅", "C", 2)
	r5 := r4.StatFmt("🔄", "F", "%d", 3)
	r6 := r5.Blank()
	r7 := r6.Separator()

	if r != r2 || r2 != r3 || r3 != r4 || r4 != r5 || r5 != r6 || r6 != r7 {
		t.Error("chaining should return same *Report pointer")
	}

	r.Print()
	if buf.Len() == 0 {
		t.Error("chained report should produce output")
	}
}

func TestSafeIcon(t *testing.T) {
	// All known emoji should map to a non-empty colored indicator.
	knownEmoji := []string{"⏱️", "⏱", "📦", "📋", "📂", "📄", "📈", "📊", "📃",
		"🔹", "🔗", "✅", "⚠️", "⚠", "🔄", "📥"}

	for _, e := range knownEmoji {
		result := safeIcon(e)
		if result == "" {
			t.Errorf("safeIcon(%q) returned empty", e)
		}
		// Should contain ANSI reset code — means it's colored.
		if !strings.Contains(result, ansiReset) {
			t.Errorf("safeIcon(%q) = %q, expected ANSI coloring", e, result)
		}
	}

	// Unknown emoji should get a default.
	result := safeIcon("🦄")
	if result == "" || !strings.Contains(result, ansiReset) {
		t.Errorf("safeIcon for unknown emoji should return colored default, got %q", result)
	}
}

func TestStripEmoji(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"10/10 suites completed ✅", "10/10 suites completed (OK)"},
		{"error ⚠️ found", "error (!) found"},
		{"all ❌ bad", "all (X) bad"},
		{"no emoji", "no emoji"},
		{"OK ✅ and ⚠️ nearby", "OK (OK) and (!) nearby"},
	}
	for _, tt := range tests {
		got := StripEmoji(tt.input)
		if got != tt.want {
			t.Errorf("StripEmoji(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNoEmojiInOutput(t *testing.T) {
	var buf bytes.Buffer
	New("noemoji").
		Writer(&buf).
		Section("Test").
		Stat("⏱️", "Time", "1s").
		Stat("📦", "Cases", 100).
		Stat("⚠️", "Errors", 0).
		Stat("📊", "Total", "all good").
		Print()

	out := buf.String()

	// The output should NOT contain any emoji — only ANSI-colored ASCII.
	problematicEmoji := []string{"📊", "📦", "📋", "📂", "📄", "📈", "📃",
		"🔹", "🔗", "⏱", "⚠", "✅"}
	for _, e := range problematicEmoji {
		if strings.Contains(out, e) {
			t.Errorf("output should not contain emoji %q, got:\n%s", e, out)
		}
	}
}

func TestConsistentLineWidth(t *testing.T) {
	var buf bytes.Buffer
	New("align").
		Writer(&buf).
		Section("Test").
		Stat("⏱️", "Time", "1s").
		Stat("📦", "Cases", 100).
		Stat("⚠️", "Errors", 0).
		Stat("📊", "Total", "all good").
		Print()

	out := buf.String()
	lines := strings.Split(out, "\n")

	// All lines between ╭ and ╰ should have │ as first and last visible char.
	inside := false
	for _, line := range lines {
		if strings.HasPrefix(line, "╭") {
			inside = true
			continue
		}
		if strings.HasPrefix(line, "╰") {
			break
		}
		if inside && len(line) > 0 {
			trimmed := strings.TrimSpace(line)
			if len(trimmed) == 0 {
				continue
			}
			// go-pretty lines start with │ and end with │
			if !strings.HasPrefix(trimmed, "│") && !strings.HasPrefix(trimmed, "├") {
				t.Errorf("line should start with │ or ├: %q", line)
			}
			if !strings.HasSuffix(trimmed, "│") && !strings.HasSuffix(trimmed, "┤") {
				t.Errorf("line should end with │ or ┤: %q", line)
			}
		}
	}
}

func TestIndentation(t *testing.T) {
	var buf bytes.Buffer
	New("indent").
		Writer(&buf).
		Section("Project 30").
		Stat("📋", "Suites", 10).
		Stat("📂", "Sections", 740).
		Print()

	out := buf.String()

	// Section should NOT be indented (level 0).
	if !strings.Contains(out, "Project 30") {
		t.Error("missing section name")
	}

	// Stats should be indented (level 1) — starts with "  " inside the cell.
	if !strings.Contains(out, "  ") {
		t.Error("stats should have indentation")
	}
	if !strings.Contains(out, "Suites: 10") {
		t.Error("missing stat")
	}
}

func TestInlineEmojiStripped(t *testing.T) {
	var buf bytes.Buffer
	New("strip").
		Writer(&buf).
		Stat("📈", "Completeness", "10/10 suites completed ✅").
		Stat("⚠️", "Problem", "error found").
		Print()

	out := buf.String()
	// Inline emoji should be replaced with ASCII.
	if strings.Contains(out, "✅") {
		t.Errorf("inline ✅ should be stripped, got:\n%s", out)
	}
	if !strings.Contains(out, "(OK)") {
		t.Errorf("✅ should be replaced with (OK), got:\n%s", out)
	}
}
// ---------------------------------------------------------------------------
// Color helpers and BannerWarn/BannerError (previously 0% coverage)
// ---------------------------------------------------------------------------

func TestGreen(t *testing.T) {
	s := Green("success")
	if !strings.Contains(s, "success") {
		t.Errorf("Green() should contain the input string, got %q", s)
	}
	// Should contain ANSI escape sequences for green color
	if !strings.Contains(s, "\x1b[32m") {
		t.Errorf("Green() should contain green ANSI code (\\x1b[32m), got %q", s)
	}
	if !strings.Contains(s, "\x1b[0m") {
		t.Errorf("Green() should contain ANSI reset code, got %q", s)
	}
}

func TestYellow(t *testing.T) {
	s := Yellow("warning")
	if !strings.Contains(s, "warning") {
		t.Errorf("Yellow() should contain the input string, got %q", s)
	}
	// Should contain ANSI escape sequences for yellow color
	if !strings.Contains(s, "\x1b[33m") {
		t.Errorf("Yellow() should contain yellow ANSI code (\\x1b[33m), got %q", s)
	}
	if !strings.Contains(s, "\x1b[0m") {
		t.Errorf("Yellow() should contain ANSI reset code, got %q", s)
	}
}

func TestBold(t *testing.T) {
	s := Bold("important")
	if !strings.Contains(s, "important") {
		t.Errorf("Bold() should contain the input string, got %q", s)
	}
	// Should contain ANSI escape sequences for bold style
	if !strings.Contains(s, "\x1b[1m") {
		t.Errorf("Bold() should contain bold ANSI code (\\x1b[1m), got %q", s)
	}
	if !strings.Contains(s, "\x1b[0m") {
		t.Errorf("Bold() should contain ANSI reset code, got %q", s)
	}
}

func TestRed(t *testing.T) {
	s := Red("error text")
	if !strings.Contains(s, "error text") {
		t.Errorf("Red() should contain the input string, got %q", s)
	}
	// Should contain ANSI escape sequences for red color
	if !strings.Contains(s, "\x1b[31m") {
		t.Errorf("Red() should contain red ANSI code (\\x1b[31m), got %q", s)
	}
	if !strings.Contains(s, "\x1b[0m") {
		t.Errorf("Red() should contain ANSI reset code, got %q", s)
	}
}

func TestReporterStyling_AllColors(t *testing.T) {
	var buf bytes.Buffer
	New("colors").
		Writer(&buf).
		Section("Color examples").
		Stat("✅", "Green text", Green("OK")).
		Stat("⚠️", "Yellow text", Yellow("WARNING")).
		Stat("❌", "Red text", Red("FAILED")).
		Section("Style examples").
		Stat("📝", "Bold text", Bold("IMPORTANT")).
		Print()

	out := buf.String()
	// All text should be present (with stripped emoji)
	if !strings.Contains(out, "OK") {
		t.Error("Green('OK') output not found")
	}
	if !strings.Contains(out, "WARNING") {
		t.Error("Yellow('WARNING') output not found")
	}
	if !strings.Contains(out, "FAILED") {
		t.Error("Red('FAILED') output not found")
	}
	if !strings.Contains(out, "IMPORTANT") {
		t.Error("Bold('IMPORTANT') output not found")
	}
}

func TestReporterStyling_BannerOK(t *testing.T) {
	var buf bytes.Buffer
	New("ok_banner").
		Writer(&buf).
		Section("Status").
		Stat("📊", "Results", "all passed").
		BannerOK("All tests completed successfully").
		Print()

	out := buf.String()
	if !strings.Contains(out, "All tests completed successfully") {
		t.Errorf("BannerOK text not found in output:\n%s", out)
	}
	// Banner should be a full-width element
	if !strings.Contains(out, "╭") || !strings.Contains(out, "╰") {
		t.Error("Banner should have box borders")
	}
}

func TestReporterStyling_BoldFormatting(t *testing.T) {
	var buf bytes.Buffer
	New("bold_test").
		Writer(&buf).
		Section("Important").
		Stat("📌", "Key metric", Bold("42")).
		Stat("📌", "Key metric", Bold("HIGH")).
		Print()

	out := buf.String()
	if !strings.Contains(out, "42") {
		t.Error("Bold numeric value not rendered")
	}
	if !strings.Contains(out, "HIGH") {
		t.Error("Bold text value not rendered")
	}
}

func TestBannerWarn(t *testing.T) {
	var buf bytes.Buffer
	New("warn test").
		Writer(&buf).
		BannerWarn("partial results").
		Print()
	out := buf.String()
	if !strings.Contains(out, "partial results") {
		t.Errorf("BannerWarn text not found in output:\n%s", out)
	}
}

func TestBannerError(t *testing.T) {
	var buf bytes.Buffer
	New("error test").
		Writer(&buf).
		BannerError("critical failure").
		Print()
	out := buf.String()
	if !strings.Contains(out, "critical failure") {
		t.Errorf("BannerError text not found in output:\n%s", out)
	}
}