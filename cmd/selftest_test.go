package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/Korrnals/gotr/internal/selftest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubChecker struct {
	name string
	cat  string
}

func (s stubChecker) Name() string { return s.name }

func (s stubChecker) Category() string { return s.cat }

func (s stubChecker) Check() selftest.CheckResult {
	return selftest.CheckResult{Result: selftest.ResultPass, Message: "ok"}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fn()

	require.NoError(t, w.Close())
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	require.NoError(t, r.Close())
	return buf.String()
}

func TestOutputJSON(t *testing.T) {
	report := &selftest.Report{
		Version: "v1",
		Checks: []selftest.CheckResult{{
			Name:   "check-1",
			Result: selftest.ResultPass,
		}},
	}

	out := captureStdout(t, func() {
		err := outputJSON(report)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "\"version\": \"v1\"")
	assert.Contains(t, out, "\"check-1\"")
}

func TestOutputJSON_WriteError(t *testing.T) {
	report := &selftest.Report{Version: "v1"}

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	require.NoError(t, w.Close())
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
		_ = r.Close()
	}()

	err = outputJSON(report)
	assert.Error(t, err)
}

func TestOutputHuman(t *testing.T) {
	report := &selftest.Report{
		Version: "v1",
		TotalFailed: 1,
		Checks: []selftest.CheckResult{
			{Name: "pass", Category: "cat", Result: selftest.ResultPass},
			{Name: "warn", Category: "cat", Result: selftest.ResultWarn},
			{Name: "fail", Category: "cat", Result: selftest.ResultFail},
		},
	}

	_ = captureStdout(t, func() {
		err := outputHuman(report, true)
		require.Error(t, err)
	})

	assert.Len(t, report.Checks, 2)
}

func TestOutputHuman_NoFailures(t *testing.T) {
	report := &selftest.Report{
		Version: "v1",
		Checks: []selftest.CheckResult{
			{Name: "pass", Category: "cat", Result: selftest.ResultPass},
		},
	}

	out := captureStdout(t, func() {
		err := outputHuman(report, false)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "pass")
}

func TestRunSelfTest_JSONOutput(t *testing.T) {
	original := buildSelfTestReport
	defer func() { buildSelfTestReport = original }()

	buildSelfTestReport = func() *selftest.Report {
		return &selftest.Report{
			Version: "stub-version",
			Checks: []selftest.CheckResult{{
				Name:   "stub-check",
				Result: selftest.ResultPass,
			}},
		}
	}

	cmd := &cobra.Command{Use: "self-test"}
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("failures-only", false, "")
	require.NoError(t, cmd.Flags().Set("json", "true"))

	out := captureStdout(t, func() {
		err := runSelfTest(cmd, nil)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "stub-check")
}

func TestRunSelfTest_HumanFailuresOnly(t *testing.T) {
	original := buildSelfTestReport
	defer func() { buildSelfTestReport = original }()

	buildSelfTestReport = func() *selftest.Report {
		return &selftest.Report{
			Version:     "stub-version",
			TotalFailed: 1,
			Checks: []selftest.CheckResult{
				{Name: "pass-check", Category: "cat", Result: selftest.ResultPass},
				{Name: "warn-check", Category: "cat", Result: selftest.ResultWarn},
				{Name: "fail-check", Category: "cat", Result: selftest.ResultFail},
			},
		}
	}

	cmd := &cobra.Command{Use: "self-test"}
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("failures-only", false, "")
	require.NoError(t, cmd.Flags().Set("failures-only", "true"))

	out := captureStdout(t, func() {
		err := runSelfTest(cmd, nil)
		require.Error(t, err)
	})

	assert.Contains(t, out, "fail-check")
	assert.Contains(t, out, "warn-check")
	assert.NotContains(t, out, "pass-check")
}

func TestRunSelfTest_HumanSuccess(t *testing.T) {
	original := buildSelfTestReport
	defer func() { buildSelfTestReport = original }()

	buildSelfTestReport = func() *selftest.Report {
		return &selftest.Report{
			Version: "stub-version",
			Checks: []selftest.CheckResult{
				{Name: "pass-check", Category: "cat", Result: selftest.ResultPass},
			},
		}
	}

	cmd := &cobra.Command{Use: "self-test"}
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("failures-only", false, "")

	out := captureStdout(t, func() {
		err := runSelfTest(cmd, nil)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "pass-check")
}

func TestBuildSelfTestReport_UsesInjectedCheckers(t *testing.T) {
	original := selfTestCheckers
	defer func() { selfTestCheckers = original }()

	selfTestCheckers = func() []selftest.Checker {
		return []selftest.Checker{
			stubChecker{name: "c1", cat: "cat1"},
			stubChecker{name: "c2", cat: "cat2"},
		}
	}

	report := buildSelfTestReport()
	require.NotNil(t, report)
	require.Len(t, report.Checks, 2)
	assert.Equal(t, "c1", report.Checks[0].Name)
	assert.Equal(t, "cat1", report.Checks[0].Category)
	assert.Equal(t, "c2", report.Checks[1].Name)
	assert.Equal(t, "cat2", report.Checks[1].Category)
	assert.NotEmpty(t, report.Version)
	assert.NotEmpty(t, report.GoVersion)
	assert.NotEmpty(t, report.Platform)
}

func TestSelfTestCheckers_DefaultSet(t *testing.T) {
	checkers := selfTestCheckers()
	require.Len(t, checkers, 6)

	names := make([]string, 0, len(checkers))
	categories := make([]string, 0, len(checkers))
	for _, c := range checkers {
		names = append(names, c.Name())
		categories = append(categories, c.Category())
	}

	assert.Contains(t, names, "Binary Information")
	assert.Contains(t, names, "Go Environment")
	assert.Contains(t, names, "Base Directory Structure")
	assert.Contains(t, names, "Configuration File")
	assert.Contains(t, names, "All Unit Tests")
	assert.Contains(t, names, "Code Coverage")

	assert.Contains(t, categories, "System")
	assert.Contains(t, categories, "Configuration")
	assert.Contains(t, categories, "Tests")
	assert.Contains(t, categories, "Coverage")
}
