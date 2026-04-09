// internal/selftest/checks.go
// Concrete checks for the self-test command.
package selftest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Korrnals/gotr/internal/paths"
)

var execCommand = exec.Command

// ConfigChecker verifies configuration in ~/.gotr/config/.
type ConfigChecker struct{}

// Name returns the display name of the configuration check.
func (c ConfigChecker) Name() string { return "Configuration File" }

// Category returns the check category shown in self-test output.
func (c ConfigChecker) Category() string { return "Configuration" }

// Check verifies that the default configuration file exists.
func (c ConfigChecker) Check() CheckResult {
	// Check path ~/.gotr/config/default.yaml
	configPath, err := paths.ConfigFile()
	if err != nil {
		return CheckResult{
			Result:  ResultFail,
			Message: "Cannot determine config path",
			Error:   err,
		}
	}

	// Check existence
	if _, err := os.Stat(configPath); err == nil {
		return CheckResult{
			Result:  ResultPass,
			Message: "Config file found",
			Details: configPath,
		}
	}

	// Config not found
	return CheckResult{
		Result:     ResultFail,
		Message:    "Config file not found",
		Details:    fmt.Sprintf("Expected: %s", configPath),
		CanFix:     true,
		FixCommand: "gotr config init",
	}
}

// BaseDirChecker verifies the ~/.gotr/ directory structure.
type BaseDirChecker struct{}

// Name returns the display name of the base-directory check.
func (c BaseDirChecker) Name() string { return "Base Directory Structure" }

// Category returns the check category shown in self-test output.
func (c BaseDirChecker) Category() string { return "Configuration" }

// Check validates and creates required gotr runtime directories.
func (c BaseDirChecker) Check() CheckResult {
	missing := []string{}
	failedDirs := []string{}

	// Check all directories
	dirChecks := []struct {
		name string
		fn   func() (string, error)
	}{
		{"config", paths.ConfigDirPath},
		{"logs", paths.LogsDirPath},
		{"selftest", paths.SelftestDirPath},
		{"cache", paths.CacheDirPath},
		{"exports", paths.ExportsDirPath},
		{"temp", paths.TempDirPath},
	}

	for _, check := range dirChecks {
		dir, err := check.fn()
		if err != nil {
			return CheckResult{
				Result:  ResultFail,
				Message: fmt.Sprintf("Cannot determine %s path", check.name),
				Error:   err,
			}
		}

		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			missing = append(missing, check.name)
			// Auto-create missing directory
			if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
				failedDirs = append(failedDirs, check.name)
			}
		}
	}

	if len(missing) > 0 {
		if len(failedDirs) > 0 {
			return CheckResult{
				Result:  ResultWarn,
				Message: "Some directories could not be created",
				Details: fmt.Sprintf("Failed: %s", strings.Join(failedDirs, ", ")),
			}
		}
		return CheckResult{
			Result:  ResultPass,
			Message: "Directories created",
			Details: fmt.Sprintf("Created: %s", strings.Join(missing, ", ")),
		}
	}

	return CheckResult{
		Result:  ResultPass,
		Message: "All directories exist",
	}
}

// BinaryInfoChecker reports binary build information.
type BinaryInfoChecker struct {
	Version   string
	Commit    string
	BuildTime string
}

// Name returns the display name of the binary information check.
func (c BinaryInfoChecker) Name() string { return "Binary Information" }

// Category returns the check category shown in self-test output.
func (c BinaryInfoChecker) Category() string { return "System" }

// Check reports build metadata embedded into the binary.
func (c BinaryInfoChecker) Check() CheckResult {
	return CheckResult{
		Result:  ResultPass,
		Message: fmt.Sprintf("Version %s", c.Version),
		Details: fmt.Sprintf("Commit: %.8s, Built: %s", c.Commit, c.BuildTime),
	}
}

// GoEnvChecker verifies the Go runtime environment.
type GoEnvChecker struct{}

// Name returns the display name of the Go environment check.
func (c GoEnvChecker) Name() string { return "Go Environment" }

// Category returns the check category shown in self-test output.
func (c GoEnvChecker) Category() string { return "System" }

// Check reports the active Go runtime and platform environment.
func (c GoEnvChecker) Check() CheckResult {
	goVersion := runtime.Version()
	goOS := runtime.GOOS
	goArch := runtime.GOARCH
	numCPU := runtime.NumCPU()

	return CheckResult{
		Result:  ResultPass,
		Message: fmt.Sprintf("Go %s", goVersion),
		Details: fmt.Sprintf("%s/%s, %d CPUs", goOS, goArch, numCPU),
	}
}

// AllTestsChecker runs all project unit tests.
type AllTestsChecker struct{}

// Name returns the display name of the full test-suite check.
func (c AllTestsChecker) Name() string { return "All Unit Tests" }

// Category returns the check category shown in self-test output.
func (c AllTestsChecker) Category() string { return "Tests" }

// Check runs the project test suite and summarizes pass/fail statistics.
func (c AllTestsChecker) Check() CheckResult {
	// Run all tests in all packages
	cmd := execCommand("go", "test", "./...", "-v")
	cmd.Dir = getProjectRoot()
	output, err := cmd.CombinedOutput()

	outStr := string(output)

	// Count results
	passed := countMatches(outStr, "--- PASS:")
	failed := countMatches(outStr, "--- FAIL:")
	skipped := countMatches(outStr, "--- SKIP:")

	// Parse per-package summary
	pkgResults := parsePackageResults(outStr)

	result := CheckResult{
		Result:  ResultPass,
		Message: fmt.Sprintf("%d passed, %d failed, %d skipped", passed, failed, skipped),
		Details: pkgResults,
	}

	if err != nil || failed > 0 {
		result.Result = ResultFail
		if failed > 0 {
			result.Error = fmt.Errorf("%d test(s) failed", failed)
		} else {
			result.Error = err
		}
	}

	// Save full report to ~/.gotr/selftest/
	if reportDir, pathErr := paths.SelftestDirPath(); pathErr == nil {
		if mkErr := os.MkdirAll(reportDir, 0o755); mkErr == nil {
			timestamp := time.Now().Format("2006-01-02_150405")
			reportPath := filepath.Join(reportDir, fmt.Sprintf("test-report-%s.log", timestamp))

			// Add meta-information
			reportContent := fmt.Sprintf("Test Report generated: %s\n", time.Now().Format(time.RFC3339))
			reportContent += fmt.Sprintf("Results: %d passed, %d failed, %d skipped\n\n", passed, failed, skipped)
			reportContent += outStr

			if writeErr := os.WriteFile(reportPath, []byte(reportContent), 0o644); writeErr == nil {
				result.Details += fmt.Sprintf(" | Report: %s", reportPath)

				// Update "latest" symlink
				latestLink := filepath.Join(reportDir, "latest.log")
				os.Remove(latestLink) // Ignore error if not exists
				_ = os.Symlink(reportPath, latestLink)
			}
		}
	}

	return result
}

// CoverageChecker checks code coverage.
type CoverageChecker struct{}

// Name returns the display name of the coverage check.
func (c CoverageChecker) Name() string { return "Code Coverage" }

// Category returns the check category shown in self-test output.
func (c CoverageChecker) Category() string { return "Coverage" }

// Check runs coverage collection and reports overall coverage health.
func (c CoverageChecker) Check() CheckResult {
	// Run tests with coverage for all packages
	cmd := execCommand("go", "test", "./...", "-coverprofile=/tmp/gotr-coverage.out")
	cmd.Dir = getProjectRoot()
	output, err := cmd.CombinedOutput()

	if err != nil {
		return CheckResult{
			Result:  ResultWarn,
			Message: "Cannot calculate coverage",
			Error:   err,
		}
	}

	// Parse overall coverage
	outStr := string(output)
	coverage := extractOverallCoverage(outStr)

	result := CheckResult{
		Result:  ResultPass,
		Message: coverage,
		Details: "Target: 80%",
	}

	// Warn if coverage is below 50%
	if strings.Contains(coverage, "%") {
		var percent float64
		if _, err := fmt.Sscanf(coverage, "%f%%", &percent); err == nil && percent < 50 {
			result.Result = ResultWarn
			result.Details = fmt.Sprintf("Current: %s, Target: 80%%", coverage)
		}
	}

	return result
}

// Helper functions

func countMatches(s, substr string) int {
	count := 0
	for {
		idx := strings.Index(s, substr)
		if idx == -1 {
			break
		}
		count++
		s = s[idx+len(substr):]
	}
	return count
}

func getProjectRoot() string {
	// Try to find project root
	if _, err := os.Stat("go.mod"); err == nil {
		wd, err := os.Getwd()
		if err != nil {
			return "."
		}
		return wd
	}

	// Walk up the directory tree
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	for wd != "/" {
		parent := filepath.Dir(wd)
		if _, err := os.Stat(filepath.Join(parent, "go.mod")); err == nil {
			return parent
		}
		wd = parent
	}

	return "."
}

func parsePackageResults(output string) string {
	var results []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		// Find lines like "ok   \tpkg/path\t0.123s  coverage: 45.0%"
		if strings.HasPrefix(line, "ok  ") || strings.HasPrefix(line, "FAIL") {
			// Abbreviate path
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				pkg := filepath.Base(parts[1])
				status := "✓"
				if strings.HasPrefix(line, "FAIL") {
					status = "✗"
				}
				results = append(results, fmt.Sprintf("%s %s", status, pkg))
			}
		}
	}

	if len(results) > 5 {
		return fmt.Sprintf("%s +%d more", strings.Join(results[:5], ", "), len(results)-5)
	}
	return strings.Join(results, ", ")
}

func extractOverallCoverage(output string) string {
	// Find the line with overall coverage
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "coverage:") {
			// Extract percentage
			if start := strings.Index(line, "coverage:"); start != -1 {
				rest := line[start+9:]
				if end := strings.Index(rest, "%"); end != -1 {
					return strings.TrimSpace(rest[:end+1])
				}
			}
		}
	}
	return "unknown"
}
