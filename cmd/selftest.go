// cmd/selftest.go
// Команда gotr self-test для самодиагностики
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

var (
	jsonOutput     bool
	failuresOnly   bool
	includeSkipped bool
)

func init() {
	rootCmd.AddCommand(selfTestCmd)

	selfTestCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")
	selfTestCmd.Flags().BoolVar(&failuresOnly, "failures-only", false, "Show only failed checks")
	selfTestCmd.Flags().BoolVar(&includeSkipped, "include-skipped", false, "Include skipped checks in output")
}

func runSelfTest(cmd *cobra.Command, args []string) error {
	// Создаём runner
	runner := selftest.NewRunner()

	// Регистрируем проверки (порядок важен для отчета)
	runner.Register(selftest.BinaryInfoChecker{
		Version:   Version,
		Commit:    Commit,
		BuildTime: Date,
	})
	runner.Register(selftest.GoEnvChecker{})
	runner.Register(selftest.BaseDirChecker{})
	runner.Register(selftest.ConfigChecker{})
	runner.Register(selftest.AllTestsChecker{})
	runner.Register(selftest.CoverageChecker{})

	// Запускаем проверки
	report := runner.Run()

	// Заполняем мета-информацию
	report.Version = Version
	report.Commit = Commit
	report.GoVersion = runtime.Version()
	report.Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	// Выводим результаты
	if jsonOutput {
		return outputJSON(report)
	}

	return outputHuman(report)
}

func outputJSON(report *selftest.Report) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func outputHuman(report *selftest.Report) error {
	// Показываем путь к последнему отчёту
	if selftestDir, err := paths.SelftestDirPath(); err == nil {
		fmt.Fprintf(os.Stderr, "Detailed reports saved to: %s/latest.log\n\n", selftestDir)
	}

	// Фильтруем если нужно
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

	// Выходим с ошибкой если есть failures
	if report.TotalFailed > 0 {
		return fmt.Errorf("%d check(s) failed", report.TotalFailed)
	}

	return nil
}
