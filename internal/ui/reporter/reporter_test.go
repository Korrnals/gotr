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
	New("empty").Writer(&buf).Print()

	out := buf.String()
	if !strings.Contains(out, "СТАТИСТИКА: empty") {
		t.Errorf("missing title in output:\n%s", out)
	}
	if !strings.Contains(out, "┌") || !strings.Contains(out, "└") {
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

	// Check section headers
	if !strings.Contains(out, "◆ General") {
		t.Errorf("missing section 'General' in:\n%s", out)
	}
	if !strings.Contains(out, "◆ Details") {
		t.Errorf("missing section 'Details' in:\n%s", out)
	}

	// Check stats
	if !strings.Contains(out, "📦 Total: 42") {
		t.Errorf("missing stat 'Total: 42' in:\n%s", out)
	}
	if !strings.Contains(out, "✅ Ok: yes") {
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
	// Count separator lines (├...┤)
	lines := strings.Split(out, "\n")
	sepCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "├") {
			sepCount++
		}
	}
	// One from header + one explicit = 2
	if sepCount != 2 {
		t.Errorf("expected 2 separator lines, got %d", sepCount)
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
	if !strings.Contains(out, "│\n│  📦 B") {
		t.Errorf("blank line not rendered correctly:\n%s", out)
	}
}

func TestCompareStats(t *testing.T) {
	var buf bytes.Buffer
	CompareStats("suites", 30, 31, 5, 3, 42, 2*time.Second).
		Writer(&buf).
		Print()

	out := buf.String()

	checks := []string{
		"СТАТИСТИКА: suites",
		"Общая статистика",
		"Время выполнения",
		"Всего обработано: 50",
		"Результат сравнения",
		"Только в проекте 30: 5",
		"Только в проекте 31: 3",
		"Общих: 42",
	}

	for _, check := range checks {
		if !strings.Contains(out, check) {
			t.Errorf("missing %q in CompareStats output:\n%s", check, out)
		}
	}
}

func TestCompareResultShort(t *testing.T) {
	s := CompareResultShort("runs", 10, 20, 5, 3, 42)
	expected := "runs: П10=5, П20=3, общих=42"
	if s != expected {
		t.Errorf("expected %q, got %q", expected, s)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Millisecond, "500мс"},
		{5 * time.Second, "5с"},
		{90 * time.Second, "1м30с"},
		{3*time.Minute + 5*time.Second, "3м05с"},
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

	if !strings.Contains(s, "СТАТИСТИКА: test") {
		t.Error("String() should return rendered report")
	}
}

func TestChaining(t *testing.T) {
	// Ensure all methods return *Report for chaining
	var buf bytes.Buffer
	r := New("chain").Writer(&buf)
	r2 := r.Section("S")
	r3 := r2.Stat("📦", "V", 1)
	r4 := r3.StatIf(true, "✅", "C", 2)
	r5 := r4.StatFmt("🔄", "F", "%d", 3)
	r6 := r5.Blank()
	r7 := r6.Separator()

	// All should be the same pointer
	if r != r2 || r2 != r3 || r3 != r4 || r4 != r5 || r5 != r6 || r6 != r7 {
		t.Error("chaining should return same *Report pointer")
	}

	r.Print()
	if buf.Len() == 0 {
		t.Error("chained report should produce output")
	}
}
