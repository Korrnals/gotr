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
