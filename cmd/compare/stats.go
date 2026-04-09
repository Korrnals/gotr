package compare

import (
	"fmt"
	"time"

	"github.com/Korrnals/gotr/pkg/reporter"
)

func bold(v interface{}) string {
	return reporter.Bold(fmt.Sprintf("%v", v))
}

// PrintCompareStats prints universal statistics for compare commands.
// Delegates to reporter.CompareStats for go-pretty rendering.
func PrintCompareStats(resource string, pid1, pid2 int64, onlyFirst, onlySecond, common int, elapsed time.Duration, status ...CompareStatus) {
	resolvedStatus := CompareStatusComplete
	if len(status) > 0 && status[0] != "" {
		resolvedStatus = status[0]
	}

	total := onlyFirst + onlySecond + common
	r := reporter.New(resource).
		Section("Execution status")

	switch resolvedStatus {
	case CompareStatusInterrupted:
		r.BannerError("INTERRUPTED - data is INCOMPLETE")
	case CompareStatusPartial:
		r.BannerWarn("PARTIAL - data may be incomplete")
	default:
		r.BannerOK("COMPLETE - all data fully loaded")
	}

	r.Section("General statistics").
		Stat("⏱️", "Execution time", bold(elapsed.Round(time.Millisecond))).
		Stat("📦", "Total processed", bold(total)).
		Section("Comparison results").
		Stat("🔹", fmt.Sprintf("Only in project %d", pid1), bold(onlyFirst)).
		Stat("🔹", fmt.Sprintf("Only in project %d", pid2), bold(onlySecond)).
		Stat("🔗", "Common", bold(common))

	r.Print()
}

// PrintCasesStatsWithErrors prints compare-cases statistics with error/retry diagnostics.
// Uses reporter.Report builder for dynamic sections.
func PrintCasesStatsWithErrors(
	pid1, pid2 int64,
	onlyFirst, onlySecond, common int,
	elapsed time.Duration,
	stats casesExecutionStats,
) {
	total := onlyFirst + onlySecond + common

	r := reporter.New("cases").
		Section("Execution status")

	// Status banner — must be the very first thing users see.
	switch {
	case stats.Interrupted:
		r.BannerError("INTERRUPTED — data is INCOMPLETE  (Ctrl+C detected; results are partial)")
	case stats.FailedPagesAfter > 0:
		r.BannerWarn(fmt.Sprintf("PARTIAL — %d page(s) failed to load  (data may be missing entries)", stats.FailedPagesAfter))
	default:
		r.BannerOK("COMPLETE — all data fully loaded")
	}

	r.Section("General statistics").
		Stat("⏱️", "Execution time", bold(elapsed.Round(time.Millisecond))).
		Stat("📦", "Total unique cases", bold(total)).
		Section(fmt.Sprintf("Project %d", pid1)).
		Stat("📋", "Suites", bold(stats.Project1.Suites)).
		Stat("📂", "Sections", bold(stats.Project1.Sections)).
		Stat("📄", "Cases (unique)", bold(stats.Project1.CasesUnique)).
		StatIf(stats.Project1.CasesRaw != stats.Project1.CasesUnique,
			"📄", "Cases (raw before dedup)", bold(stats.Project1.CasesRaw)).
		StatIf(stats.Project1.CasesExpected > 0,
			"📊", "Cases (expected by API)", fmt.Sprintf("%d from %d suites",
				stats.Project1.CasesExpected, stats.Project1.SuitesWithTotal)).
		StatFmt("📈", "Download completeness", "%s",
			formatCompleteness(stats.Project1.CasesRaw, stats.Project1.CasesExpected,
				stats.Project1.Suites, stats.Project1.SuitesVerified, stats.Project1.FailedPages)).
		StatIf(stats.Project1.SuiteDetailsCount > 0,
			"📊", "Total cases in project", formatIntegrityCheck(
				stats.Project1.CasesRaw, stats.Project1.SuiteDetailsCount,
				stats.Project1.SuiteDetailsSum, stats.Project1.SuiteDetailsEmpty)).
		StatIf(stats.Project1.TotalPages > 0,
			"📃", "Pages loaded", fmt.Sprintf("%d (errors: %d)",
				stats.Project1.TotalPages, stats.Project1.FailedPages)).
		StatIf(stats.Project1.EmptyTitles > 0,
			"⚠️", "Cases without title", bold(stats.Project1.EmptyTitles)).
		Stat("⏱️", "Loading", bold(stats.Project1.Elapsed.Round(time.Millisecond))).
		Section(fmt.Sprintf("Project %d", pid2)).
		Stat("📋", "Suites", bold(stats.Project2.Suites)).
		Stat("📂", "Sections", bold(stats.Project2.Sections)).
		Stat("📄", "Cases (unique)", bold(stats.Project2.CasesUnique)).
		StatIf(stats.Project2.CasesRaw != stats.Project2.CasesUnique,
			"📄", "Cases (raw before dedup)", bold(stats.Project2.CasesRaw)).
		StatIf(stats.Project2.CasesExpected > 0,
			"📊", "Cases (expected by API)", fmt.Sprintf("%d from %d suites",
				stats.Project2.CasesExpected, stats.Project2.SuitesWithTotal)).
		StatFmt("📈", "Download completeness", "%s",
			formatCompleteness(stats.Project2.CasesRaw, stats.Project2.CasesExpected,
				stats.Project2.Suites, stats.Project2.SuitesVerified, stats.Project2.FailedPages)).
		StatIf(stats.Project2.SuiteDetailsCount > 0,
			"📊", "Total cases in project", formatIntegrityCheck(
				stats.Project2.CasesRaw, stats.Project2.SuiteDetailsCount,
				stats.Project2.SuiteDetailsSum, stats.Project2.SuiteDetailsEmpty)).
		StatIf(stats.Project2.TotalPages > 0,
			"📃", "Pages loaded", fmt.Sprintf("%d (errors: %d)",
				stats.Project2.TotalPages, stats.Project2.FailedPages)).
		StatIf(stats.Project2.EmptyTitles > 0,
			"⚠️", "Cases without title", bold(stats.Project2.EmptyTitles)).
		Stat("⏱️", "Loading", bold(stats.Project2.Elapsed.Round(time.Millisecond))).
		Section("Errors and retries").
		StatFmt("⚠️", "Load errors", "P%d=%d, P%d=%d", pid1, stats.LoadErrorsP1, pid2, stats.LoadErrorsP2).
		Stat("⚠️", "Failed pages before auto-retry", bold(stats.FailedPagesBefore)).
		StatIf(stats.RetryAttempted, "🔄", "Recovered pages",
			fmt.Sprintf("%d/%d", stats.RetryStats.RecoveredPages, stats.RetryStats.UniquePages)).
		StatIf(stats.RetryAttempted, "📥", "Cases recovered on retry", bold(stats.RetryStats.RecoveredCases)).
		StatIf(stats.RetryAttempted, "⚠️", "Failed pages after auto-retry", bold(stats.FailedPagesAfter)).
		Section("Comparison results").
		Stat("🔹", fmt.Sprintf("Unique cases in project %d", pid1), bold(onlyFirst)).
		Stat("🔹", fmt.Sprintf("Unique cases in project %d", pid2), bold(onlySecond)).
		Stat("🔗", "Common cases", bold(common))

	r.Print()
}

// formatCompleteness returns a human-readable completeness string.
// totalSuites/suitesVerified track suite-level exhaustion verification.
// failedPages indicates how many pages had permanent fetch errors.
func formatCompleteness(actual, expected, totalSuites, suitesVerified, failedPages int) string {
	// Suite verification status
	suiteStatus := ""
	if totalSuites > 0 {
		if suitesVerified == totalSuites && failedPages == 0 {
			suiteStatus = reporter.Green(fmt.Sprintf("%d/%d suites completed (OK)", suitesVerified, totalSuites))
		} else {
			incomplete := totalSuites - suitesVerified
			suiteStatus = reporter.Red(fmt.Sprintf("%d/%d suites completed, %d incomplete", suitesVerified, totalSuites, incomplete))
		}
	}

	if expected <= 0 {
		if suiteStatus != "" {
			return reporter.Bold(fmt.Sprintf("%d", actual)) + " (" + suiteStatus + ")"
		}
		if failedPages == 0 {
			return reporter.Bold(fmt.Sprintf("%d", actual)) + " (" + reporter.Green("fully loaded (OK)") + ")"
		}
		return reporter.Bold(fmt.Sprintf("%d", actual)) + " (" + reporter.Red(fmt.Sprintf("errors: %d pages", failedPages)) + ")"
	}
	pct := float64(actual) / float64(expected) * 100
	if actual == expected && failedPages == 0 {
		return reporter.Green(reporter.Bold(fmt.Sprintf("%d/%d (100%%)", actual, expected)))
	}
	if actual > expected {
		return reporter.Red(reporter.Bold(fmt.Sprintf("%d/%d (%.1f%%, possible duplicates)", actual, expected, pct)))
	}
	return reporter.Red(reporter.Bold(fmt.Sprintf("%d/%d (%.1f%%)", actual, expected, pct)))
}

// formatIntegrityCheck shows total cases across all suites in the project.
// User compares this visually with "Download completeness" to see if everything was fetched.
func formatIntegrityCheck(casesRaw, suiteCount, suiteSum, emptySuites int) string {
	if suiteCount == 0 {
		return ""
	}
	emptyNote := ""
	if emptySuites > 0 {
		emptyNote = fmt.Sprintf(", empty: %d", emptySuites)
	}
	if suiteSum == casesRaw {
		return fmt.Sprintf("%d (%d suites%s)",
			suiteSum, suiteCount, emptyNote)
	}
	diff := casesRaw - suiteSum
	return fmt.Sprintf("%d (%d suites%s) ⚠️ loaded %d, delta %+d",
		suiteSum, suiteCount, emptyNote, casesRaw, diff)
}
