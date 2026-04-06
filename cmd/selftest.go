// cmd/selftest.go
// Self-diagnostic command: gotr self-test
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/Korrnals/gotr/internal/paths"
	"github.com/Korrnals/gotr/internal/selftest"
	"github.com/spf13/cobra"
)

// selfTestCmd represents the self-test command
var selfTestCmd = &cobra.Command{
	Use:   "self-test",
	Short: "Run self-diagnostic tests",
	Long: `Run comprehensive self-diagnostic tests to verify gotr installation,
configuration, and internal health.

Checks performed:
  - Binary information (version, commit, build time)
  - Go environment (version, OS/arch, CPUs)
  - Base directory structure (~/.testrail/)
  - Configuration file (validates ~/.testrail/config/gotr.yaml)
  - All unit tests (runs go test ./...)
  - Code coverage metrics

Reports are saved to: ~/.testrail/selftest/

Examples:
  # Run all checks
  gotr self-test

  # Output as JSON for CI/CD
  gotr self-test --json

  # Show only failed checks
  gotr self-test --failures-only`,
	RunE: runSelfTest,
}

func init() {
	rootCmd.AddCommand(selfTestCmd)

	selfTestCmd.Flags().Bool("json", false, "Output results as JSON")
	selfTestCmd.Flags().Bool("failures-only", false, "Show only failed checks")
	selfTestCmd.Flags().Bool("include-skipped", false, "Include skipped checks in output")
}

var buildSelfTestReport = func() *selftest.Report {
	runner := selftest.NewRunner()

	// Register checks (order matters for the report)
	for _, checker := range selfTestCheckers() {
		runner.Register(checker)
	}

	// Run checks
	report := runner.Run()

	// Fill meta information
	report.Version = Version
	report.Commit = Commit
	report.GoVersion = runtime.Version()
	report.Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	return report
}

var selfTestCheckers = func() []selftest.Checker {
	return []selftest.Checker{
		selftest.BinaryInfoChecker{
			Version:   Version,
			Commit:    Commit,
			BuildTime: Date,
		},
		selftest.GoEnvChecker{},
		selftest.BaseDirChecker{},
		selftest.ConfigChecker{},
		selftest.AllTestsChecker{},
		selftest.CoverageChecker{},
	}
}

func runSelfTest(cmd *cobra.Command, args []string) error {
	report := buildSelfTestReport()

	// Output results
	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		return outputJSON(report)
	}

	failuresOnly, _ := cmd.Flags().GetBool("failures-only")
	return outputHuman(report, failuresOnly)
}

func outputJSON(report *selftest.Report) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func outputHuman(report *selftest.Report, failuresOnly bool) error {
	// Show path to the latest report
	if selftestDir, err := paths.SelftestDirPath(); err == nil {
		fmt.Fprintf(os.Stderr, "Detailed reports saved to: %s/latest.log\n\n", selftestDir)
	}

	// Filter if needed
	checks := report.Checks
	if failuresOnly {
		filtered := make([]selftest.CheckResult, 0)
		for _, c := range checks {
			if c.Result == selftest.ResultFail || c.Result == selftest.ResultWarn {
				filtered = append(filtered, c)
			}
		}
		checks = filtered
		report.Checks = filtered
	}

	report.PrintHuman()

	// Exit with error if there are failures
	if report.TotalFailed > 0 {
		return fmt.Errorf("%d check(s) failed", report.TotalFailed)
	}

	return nil
}
