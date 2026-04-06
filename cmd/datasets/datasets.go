// Package datasets implements CLI commands for managing TestRail datasets.
package datasets

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is a function type for obtaining the API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all dataset management subcommands.
func Register(root *cobra.Command, getClient GetClientFunc) {
	datasetsCmd := &cobra.Command{
		Use:   "datasets",
		Short: "Управление датасетами (тестовыми данными)",
		Long: `Управление датасетами (datasets) — таблицами тестовых данных
для параметризованного тестирования.

Датасеты позволяют запускать один и тот же тест-кейс с разными
наборами входных данных без создания дубликатов кейсов.
Каждый датасет содержит название и таблицу с колонками (параметрами)
и строками (значениями).

Используются при создании тест-планов с параметризованными тест-ранами.

Доступные операции:
  • list   — список датасетов проекта
  • get    — получить датасет по ID
  • add    — создать новый датасет
  • update — обновить датасет
  • delete — удалить датасет`,
	}

	// Register subcommands
	datasetsCmd.AddCommand(newListCmd(getClient))
	datasetsCmd.AddCommand(newGetCmd(getClient))
	datasetsCmd.AddCommand(newAddCmd(getClient))
	datasetsCmd.AddCommand(newUpdateCmd(getClient))
	datasetsCmd.AddCommand(newDeleteCmd(getClient))

	root.AddCommand(datasetsCmd)
}
