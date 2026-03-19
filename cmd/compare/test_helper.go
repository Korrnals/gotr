package compare

import (
	"time"

	"github.com/spf13/cobra"
)

// addPersistentFlagsForTests adds persistent flags that are normally registered
// in Register() function. Use this in tests that create commands directly.
func addPersistentFlagsForTests(cmd *cobra.Command) {
	cmd.Flags().StringP("pid1", "1", "", "ID первого проекта")
	cmd.Flags().StringP("pid2", "2", "", "ID второго проекта")
	cmd.Flags().StringP("format", "f", "table", "Формат вывода")
	cmd.Flags().BoolP("quiet", "q", false, "Подавить служебный вывод")
	cmd.Flags().Bool("save", false, "Сохранить результат")
	cmd.Flags().String("save-to", "", "Сохранить в указанный файл")
	cmd.Flags().Int("rate-limit", -1, "")
	cmd.Flags().Int("parallel-suites", 8, "")
	cmd.Flags().Int("parallel-pages", 10, "")
	cmd.Flags().Int("page-retries", 5, "")
	cmd.Flags().Duration("timeout", 30*time.Minute, "")
	cmd.Flags().Int("retry-attempts", 3, "")
	cmd.Flags().Int("retry-workers", 6, "")
	cmd.Flags().Duration("retry-delay", 500*time.Millisecond, "")
}
