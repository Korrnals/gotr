package compare

import (
	"fmt"
	"time"
)

// PrintCompareStats prints universal statistics for compare commands
func PrintCompareStats(resource string, pid1, pid2 int64, onlyFirst, onlySecond, common int, elapsed time.Duration) {
	total := onlyFirst + onlySecond + common
	
	fmt.Println()
	fmt.Println("┌──────────────────────────────────────────────────────────────┐")
	fmt.Printf("│          📊 СТАТИСТИКА: %s\n", resource)
	fmt.Println("├──────────────────────────────────────────────────────────────┤")
	fmt.Printf("│  ⏱️  Время выполнения: %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("│  📦 Всего обработано: %d\n", total)
	fmt.Println("├──────────────────────────────────────────────────────────────┤")
	fmt.Printf("│  ✅ Только в проекте %d: %d\n", pid1, onlyFirst)
	fmt.Printf("│  ✅ Только в проекте %d: %d\n", pid2, onlySecond)
	fmt.Printf("│  🔗 Общих: %d\n", common)
	fmt.Println("└──────────────────────────────────────────────────────────────┘")
}

// PrintCompareResultShort prints short result (for non-compare commands)
func PrintCompareResultShort(resource string, pid1, pid2 int64, onlyFirst, onlySecond, common int) {
	fmt.Printf("\n%s: П%d=%d, П%d=%d, общих=%d\n", 
		resource, pid1, onlyFirst, pid2, onlySecond, common)
}

// PrintCasesStatsWithErrors prints compare-cases statistics with error/retry diagnostics.
func PrintCasesStatsWithErrors(
	pid1, pid2 int64,
	onlyFirst, onlySecond, common int,
	elapsed time.Duration,
	stats casesExecutionStats,
) {
	total := onlyFirst + onlySecond + common

	fmt.Println()
	fmt.Println("┌──────────────────────────────────────────────────────────────┐")
	fmt.Println("│          📊 СТАТИСТИКА: cases")
	fmt.Println("├──────────────────────────────────────────────────────────────┤")
	fmt.Printf("│  ⏱️  Время выполнения: %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("│  📦 Всего обработано: %d\n", total)
	fmt.Printf("│  ⚠️  Ошибки загрузки: П%d=%d, П%d=%d\n", pid1, stats.LoadErrorsP1, pid2, stats.LoadErrorsP2)
	fmt.Printf("│  ⚠️  Failed pages до авто-ретрая: %d\n", stats.FailedPagesBefore)
	if stats.RetryAttempted {
		fmt.Printf("│  🔄 Авто-ретрай: восстановлено страниц %d/%d\n", stats.RetryStats.RecoveredPages, stats.RetryStats.UniquePages)
		fmt.Printf("│  📥 Авто-ретрай: получено кейсов %d\n", stats.RetryStats.RecoveredCases)
		fmt.Printf("│  ⚠️  Failed pages после авто-ретрая: %d\n", stats.FailedPagesAfter)
	}
	fmt.Println("├──────────────────────────────────────────────────────────────┤")
	fmt.Printf("│  ✅ Только в проекте %d: %d\n", pid1, onlyFirst)
	fmt.Printf("│  ✅ Только в проекте %d: %d\n", pid2, onlySecond)
	fmt.Printf("│  🔗 Общих: %d\n", common)
	fmt.Println("└──────────────────────────────────────────────────────────────┘")
}
