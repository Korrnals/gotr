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
