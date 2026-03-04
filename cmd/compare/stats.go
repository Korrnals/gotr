package compare

import (
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/ui/reporter"
)

// PrintCompareStats prints universal statistics for compare commands.
// Delegates to reporter.CompareStats for centralized formatting.
func PrintCompareStats(resource string, pid1, pid2 int64, onlyFirst, onlySecond, common int, elapsed time.Duration) {
	reporter.CompareStats(resource, pid1, pid2, onlyFirst, onlySecond, common, elapsed).Print()
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
		Section("Общая статистика").
		Stat("⏱️", "Время выполнения", elapsed.Round(time.Millisecond)).
		Stat("📦", "Всего обработано", total).
		Section("Ошибки и ретраи").
		StatFmt("⚠️", "Ошибки загрузки", "П%d=%d, П%d=%d", pid1, stats.LoadErrorsP1, pid2, stats.LoadErrorsP2).
		Stat("⚠️", "Failed pages до авто-ретрая", stats.FailedPagesBefore).
		StatIf(stats.RetryAttempted, "🔄", "Восстановлено страниц",
			fmt.Sprintf("%d/%d", stats.RetryStats.RecoveredPages, stats.RetryStats.UniquePages)).
		StatIf(stats.RetryAttempted, "📥", "Получено кейсов при ретрае", stats.RetryStats.RecoveredCases).
		StatIf(stats.RetryAttempted, "⚠️", "Failed pages после авто-ретрая", stats.FailedPagesAfter).
		Section("Результат сравнения").
		Stat("✅", fmt.Sprintf("Только в проекте %d", pid1), onlyFirst).
		Stat("✅", fmt.Sprintf("Только в проекте %d", pid2), onlySecond).
		Stat("🔗", "Общих", common)

	r.Print()
}
