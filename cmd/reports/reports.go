// Package reports implements CLI commands for managing TestRail reports.
package reports

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining an API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register adds all report-related subcommands to the root command.
func Register(root *cobra.Command, getClient GetClientFunc) {
	reportsCmd := &cobra.Command{
		Use:   "reports",
		Short: "Управление отчётами проекта",
		Long: `Управление шаблонами отчётов и генерация отчётов TestRail.

Шаблоны отчётов используются для создания различных типов отчётов
о тестировании (сводные отчёты, отчёты по покрытию, сравнительные отчёты).

Доступные операции:
  • list               — список шаблонов отчётов проекта
  • list-cross-project — список кросс-проектных отчётов
  • run                — запустить генерацию отчёта по шаблону
  • run-cross-project  — запустить кросс-проектный отчёт`,
	}

	// Register subcommands
	reportsCmd.AddCommand(newListCmd(getClient))
	reportsCmd.AddCommand(newListCrossProjectCmd(getClient))
	reportsCmd.AddCommand(newRunCmd(getClient))
	reportsCmd.AddCommand(newRunCrossProjectCmd(getClient))

	root.AddCommand(reportsCmd)
}
