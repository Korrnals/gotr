package reporter

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestPrettyNew(t *testing.T) {
	r := NewPretty("test")
	if r == nil {
		t.Fatal("NewPretty returned nil")
	}
	if r.title != "test" {
		t.Errorf("expected title 'test', got %q", r.title)
	}
}

func TestPrettyEmptyReport(t *testing.T) {
	var buf bytes.Buffer
	NewPretty("empty").Writer(&buf).
		Section("Пусто").
		Stat("📦", "Тест", 0).
		Print()
	out := buf.String()
	if !strings.Contains(out, "СТАТИСТИКА: empty") {
		t.Errorf("missing title in output:\n%s", out)
	}
}

func TestPrettySectionAndStat(t *testing.T) {
	var buf bytes.Buffer
	NewPretty("items").
		Writer(&buf).
		Section("General").
		Stat("📦", "Total", 42).
		Section("Details").
		Stat("✅", "Ok", "yes").
		Print()

	out := buf.String()
	if !strings.Contains(out, "◆ General") {
		t.Errorf("missing section 'General' in:\n%s", out)
	}
	if !strings.Contains(out, "📦 Total: 42") {
		t.Errorf("missing stat 'Total: 42' in:\n%s", out)
	}
}

func TestPrettyStatIf(t *testing.T) {
	var buf bytes.Buffer
	NewPretty("cond").
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

func TestPrettyStatFmt(t *testing.T) {
	var buf bytes.Buffer
	NewPretty("fmt").
		Writer(&buf).
		StatFmt("🔄", "Recovered", "%d/%d", 3, 5).
		Print()

	out := buf.String()
	if !strings.Contains(out, "Recovered: 3/5") {
		t.Errorf("StatFmt not rendered correctly:\n%s", out)
	}
}

func TestPrettyCompareStats(t *testing.T) {
	var buf bytes.Buffer
	CompareStatsPretty("suites", 30, 31, 5, 3, 42, 2*time.Second).
		Writer(&buf).
		Print()

	out := buf.String()
	checks := []string{
		"СТАТИСТИКА: suites",
		"Общая статистика",
		"Всего обработано: 50",
		"Только в проекте 30: 5",
		"Только в проекте 31: 3",
		"Общих: 42",
	}
	for _, check := range checks {
		if !strings.Contains(out, check) {
			t.Errorf("missing %q in output:\n%s", check, out)
		}
	}
}

func TestPrettyChaining(t *testing.T) {
	var buf bytes.Buffer
	r := NewPretty("chain").Writer(&buf)
	r2 := r.Section("S")
	r3 := r2.Stat("📦", "V", 1)
	r4 := r3.StatIf(true, "✅", "C", 2)
	r5 := r4.StatFmt("🔄", "F", "%d", 3)
	r6 := r5.Blank()
	r7 := r6.Separator()
	if r != r2 || r2 != r3 || r3 != r4 || r4 != r5 || r5 != r6 || r6 != r7 {
		t.Error("chaining should return same *PrettyReport pointer")
	}
	r.Print()
	if buf.Len() == 0 {
		t.Error("chained report should produce output")
	}
}

func TestPrettyString(t *testing.T) {
	s := NewPretty("test").
		Section("A").
		Stat("📦", "Val", 1).
		String()
	if !strings.Contains(s, "СТАТИСТИКА: test") {
		t.Error("String() should return rendered report")
	}
}
