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
		Stat("📦", "Всего уникальных кейсов", total).
		Section(fmt.Sprintf("Проект %d", pid1)).
		Stat("📋", "Сьютов", stats.Project1.Suites).
		Stat("📂", "Секций", stats.Project1.Sections).
		Stat("📄", "Кейсов (уникальных)", stats.Project1.CasesUnique).
		StatIf(stats.Project1.CasesRaw != stats.Project1.CasesUnique,
			"📄", "Кейсов (raw до дедупа)", stats.Project1.CasesRaw).
		StatIf(stats.Project1.EmptyTitles > 0,
			"⚠️", "Кейсов без заголовка", stats.Project1.EmptyTitles).
		Stat("⏱️", "Загрузка", stats.Project1.Elapsed.Round(time.Millisecond)).
		Section(fmt.Sprintf("Проект %d", pid2)).
		Stat("📋", "Сьютов", stats.Project2.Suites).
		Stat("📂", "Секций", stats.Project2.Sections).
		Stat("📄", "Кейсов (уникальных)", stats.Project2.CasesUnique).
		StatIf(stats.Project2.CasesRaw != stats.Project2.CasesUnique,
			"📄", "Кейсов (raw до дедупа)", stats.Project2.CasesRaw).
		StatIf(stats.Project2.EmptyTitles > 0,
			"⚠️", "Кейсов без заголовка", stats.Project2.EmptyTitles).
		Stat("⏱️", "Загрузка", stats.Project2.Elapsed.Round(time.Millisecond)).
		Section("Ошибки и ретраи").
		StatFmt("⚠️", "Ошибки загрузки", "П%d=%d, П%d=%d", pid1, stats.LoadErrorsP1, pid2, stats.LoadErrorsP2).
		Stat("⚠️", "Failed pages до авто-ретрая", stats.FailedPagesBefore).
		StatIf(stats.RetryAttempted, "🔄", "Восстановлено страниц",
			fmt.Sprintf("%d/%d", stats.RetryStats.RecoveredPages, stats.RetryStats.UniquePages)).
		StatIf(stats.RetryAttempted, "📥", "Получено кейсов при ретрае", stats.RetryStats.RecoveredCases).
		StatIf(stats.RetryAttempted, "⚠️", "Failed pages после авто-ретрая", stats.FailedPagesAfter).
		Section("Результат сравнения").
		Stat("🔹", fmt.Sprintf("Уникальных в проекте %d", pid1), onlyFirst).
		Stat("🔹", fmt.Sprintf("Уникальных в проекте %d", pid2), onlySecond).
		Stat("🔗", "Общих", common)

	r.Print()
}
