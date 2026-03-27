package selftest

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

type stubChecker struct {
	name     string
	category string
	result   CheckResult
}

func (s stubChecker) Name() string       { return s.name }
func (s stubChecker) Category() string   { return s.category }
func (s stubChecker) Check() CheckResult { return s.result }

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe error: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()
	_ = w.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestResultStringAndColor(t *testing.T) {
	cases := []struct {
		in   Result
		icon string
		clr  string
	}{
		{ResultPass, "✓", "\033[32m"},
		{ResultFail, "✗", "\033[31m"},
		{ResultWarn, "⚠", "\033[33m"},
		{ResultSkip, "⊘", "\033[90m"},
		{Result("X"), "?", "\033[0m"},
	}

	for _, tc := range cases {
		if got := tc.in.String(); got != tc.icon {
			t.Fatalf("String(%q)=%q, want %q", tc.in, got, tc.icon)
		}
		if got := tc.in.Color(); got != tc.clr {
			t.Fatalf("Color(%q)=%q, want %q", tc.in, got, tc.clr)
		}
		if got := tc.in.ResetColor(); got != "\033[0m" {
			t.Fatalf("ResetColor=%q", got)
		}
	}
}

func TestReportCalculateHealthAndOverallStatus(t *testing.T) {
	r := &Report{Checks: []CheckResult{{Result: ResultPass}, {Result: ResultWarn}, {Result: ResultSkip}}}
	r.CalculateHealth()
	if r.Health != ResultWarn {
		t.Fatalf("health=%s, want WARN", r.Health)
	}
	if r.TotalPassed != 1 || r.TotalWarn != 1 || r.TotalSkip != 1 || r.TotalFailed != 0 {
		t.Fatalf("unexpected counters: %+v", r)
	}
	if r.OverallStatus() != "Degraded" {
		t.Fatalf("status=%q", r.OverallStatus())
	}

	r2 := &Report{Checks: []CheckResult{{Result: ResultFail}}}
	r2.CalculateHealth()
	if r2.OverallStatus() != "Unhealthy" {
		t.Fatalf("status=%q", r2.OverallStatus())
	}

	r3 := &Report{Checks: []CheckResult{{Result: ResultPass}}}
	r3.CalculateHealth()
	if r3.OverallStatus() != "Healthy" {
		t.Fatalf("status=%q", r3.OverallStatus())
	}

	r4 := &Report{Health: Result("mystery")}
	if r4.OverallStatus() != "Unknown" {
		t.Fatalf("status=%q", r4.OverallStatus())
	}
}

func TestRunnerRun(t *testing.T) {
	runner := NewRunner()
	runner.Register(stubChecker{name: "cfg", category: "config", result: CheckResult{Result: ResultPass, Message: "ok"}})
	runner.Register(stubChecker{name: "api", category: "api", result: CheckResult{Result: ResultWarn, Message: "slow"}})

	report := runner.Run()
	if len(report.Checks) != 2 {
		t.Fatalf("checks=%d", len(report.Checks))
	}
	if report.Checks[0].Name != "cfg" || report.Checks[0].Category != "config" {
		t.Fatalf("first check metadata not populated")
	}
	if report.Duration <= 0 {
		t.Fatalf("duration should be set")
	}
}

func TestReportPrintHumanAndFormatDetails(t *testing.T) {
	rep := &Report{
		Version:   "v1",
		Commit:    "abc",
		GoVersion: "go1.25",
		Platform:  "linux/amd64",
		Checks: []CheckResult{
			{Name: "cfg", Category: "config", Result: ResultPass, Message: "ok"},
			{Name: "api", Category: "api", Result: ResultFail, Error: errors.New("boom"), Details: "details"},
		},
	}
	rep.CalculateHealth()

	out := captureStdout(t, func() {
		rep.PrintHuman()
	})

	mustContain := []string{"gotr Self-Test Report", "Version:", "[config]", "[api]", "Results:", "Overall:"}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Fatalf("output missing %q", s)
		}
	}

	if got := formatDetails(CheckResult{Error: errors.New("x")}); got != "— x" {
		t.Fatalf("formatDetails error path=%q", got)
	}
	if got := formatDetails(CheckResult{Message: "msg"}); got != "— msg" {
		t.Fatalf("formatDetails message path=%q", got)
	}
	if got := formatDetails(CheckResult{}); got != "" {
		t.Fatalf("formatDetails empty path=%q", got)
	}
}
