// internal/selftest/checks.go
// Конкретные проверки для self-test
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

// ConfigChecker проверяет конфигурацию в ~/.gotr/config/
type ConfigChecker struct{}

func (c ConfigChecker) Name() string   { return "Configuration File" }
func (c ConfigChecker) Category() string { return "Configuration" }

func (c ConfigChecker) Check() CheckResult {
	// Проверяем путь ~/.gotr/config/default.yaml
	configPath, err := paths.ConfigFile()
	if err != nil {
		return CheckResult{
			Result:  ResultFail,
			Message: "Cannot determine config path",
			Error:   err,
		}
	}

	// Проверяем существование
	if _, err := os.Stat(configPath); err == nil {
		return CheckResult{
			Result:  ResultPass,
			Message: "Config file found",
			Details: configPath,
		}
	}

	// Конфиг не найден
	return CheckResult{
		Result:     ResultFail,
		Message:    "Config file not found",
		Details:    fmt.Sprintf("Expected: %s", configPath),
		CanFix:     true,
		FixCommand: "gotr config init",
	}
}

// BaseDirChecker проверяет структуру ~/.testrail/
type BaseDirChecker struct{}

func (c BaseDirChecker) Name() string   { return "Base Directory Structure" }
func (c BaseDirChecker) Category() string { return "Configuration" }

func (c BaseDirChecker) Check() CheckResult {
	missing := []string{}
	
	// Проверяем все директории
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
			// Автоматически создаём
			os.MkdirAll(dir, 0755)
		}
	}

	if len(missing) > 0 {
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

// BinaryInfoChecker проверяет информацию о бинарнике
type BinaryInfoChecker struct {
	Version   string
	Commit    string
	BuildTime string
}

func (c BinaryInfoChecker) Name() string   { return "Binary Information" }
func (c BinaryInfoChecker) Category() string { return "System" }

func (c BinaryInfoChecker) Check() CheckResult {
	return CheckResult{
		Result:  ResultPass,
		Message: fmt.Sprintf("Version %s", c.Version),
		Details: fmt.Sprintf("Commit: %.8s, Built: %s", c.Commit, c.BuildTime),
	}
}

// GoEnvChecker проверяет окружение Go
type GoEnvChecker struct{}

func (c GoEnvChecker) Name() string   { return "Go Environment" }
func (c GoEnvChecker) Category() string { return "System" }

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

// AllTestsChecker запускает все тесты проекта
type AllTestsChecker struct{}

func (c AllTestsChecker) Name() string   { return "All Unit Tests" }
func (c AllTestsChecker) Category() string { return "Tests" }

func (c AllTestsChecker) Check() CheckResult {
	// Запускаем все тесты во всех пакетах
	cmd := exec.Command("go", "test", "./...", "-v")
	cmd.Dir = getProjectRoot()
	output, err := cmd.CombinedOutput()

	outStr := string(output)
	
	// Считаем результаты
	passed := countMatches(outStr, "--- PASS:")
	failed := countMatches(outStr, "--- FAIL:")
	skipped := countMatches(outStr, "--- SKIP:")

	// Парсим сводку по пакетам
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

	// Сохраняем полный отчёт в ~/.testrail/selftest/
	if reportDir, pathErr := paths.SelftestDirPath(); pathErr == nil {
		os.MkdirAll(reportDir, 0755)
		timestamp := time.Now().Format("2006-01-02_150405")
		reportPath := filepath.Join(reportDir, fmt.Sprintf("test-report-%s.log", timestamp))
		
		// Добавляем мета-информацию
		reportContent := fmt.Sprintf("Test Report generated: %s\n", time.Now().Format(time.RFC3339))
		reportContent += fmt.Sprintf("Results: %d passed, %d failed, %d skipped\n\n", passed, failed, skipped)
		reportContent += outStr
		
		if writeErr := os.WriteFile(reportPath, []byte(reportContent), 0644); writeErr == nil {
			result.Details += fmt.Sprintf(" | Report: %s", reportPath)
			
			// Обновляем симлинк latest
			latestLink := filepath.Join(reportDir, "latest.log")
			os.Remove(latestLink) // Игнорируем ошибку если не существует
			os.Symlink(reportPath, latestLink)
		}
	}

	return result
}

// CoverageChecker проверяет покрытие кода
type CoverageChecker struct{}

func (c CoverageChecker) Name() string   { return "Code Coverage" }
func (c CoverageChecker) Category() string { return "Coverage" }

func (c CoverageChecker) Check() CheckResult {
	// Запускаем тесты с coverage для всех пакетов
	cmd := exec.Command("go", "test", "./...", "-coverprofile=/tmp/gotr-coverage.out")
	cmd.Dir = getProjectRoot()
	output, err := cmd.CombinedOutput()

	if err != nil {
		return CheckResult{
			Result:  ResultWarn,
			Message: "Cannot calculate coverage",
			Error:   err,
		}
	}

	// Парсим общее покрытие
	outStr := string(output)
	coverage := extractOverallCoverage(outStr)

	result := CheckResult{
		Result:  ResultPass,
		Message: coverage,
		Details: "Target: 80%",
	}

	// Если покрытие ниже 50% — предупреждение
	if strings.Contains(coverage, "%") {
		var percent float64
		fmt.Sscanf(coverage, "%f%%", &percent)
		if percent < 50 {
			result.Result = ResultWarn
			result.Details = fmt.Sprintf("Current: %s, Target: 80%%", coverage)
		}
	}

	return result
}

// Вспомогательные функции

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
	// Пытаемся найти корень проекта
	if _, err := os.Stat("go.mod"); err == nil {
		wd, _ := os.Getwd()
		return wd
	}

	// Ищем вверх по дереву
	wd, _ := os.Getwd()
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
		// Ищем строки вида "ok   	pkg/path	0.123s  coverage: 45.0%"
		if strings.HasPrefix(line, "ok  ") || strings.HasPrefix(line, "FAIL") {
			// Сокращаем путь
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
	// Ищем строку с общим покрытием
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "coverage:") {
			// Извлекаем процент
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
