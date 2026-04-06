// Package plans implements CLI commands for managing TestRail test plans.
package plans

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining an API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all test plan management commands on the given root.
func Register(root *cobra.Command, getClient GetClientFunc) {
	plansCmd := &cobra.Command{
		Use:   "plans",
		Short: "Управление тест-планами",
		Long: `Управление тест-планами: создание, обновление, закрытие, удаление и управление записями.

Тест-план — это набор тестовых прогонов (entries), объединённых общей целью.

Основные операции:
  • add    — создать тест-план
  • get    — получить информацию о плане
  • list   — список планов проекта
  • update — обновить план
  • close  — закрыть план (завершить)
  • delete — удалить план
  • entry  — управление записями плана (add/update/delete)`,
	}

	// Add subcommands
	plansCmd.AddCommand(newAddCmd(getClient))
	plansCmd.AddCommand(newGetCmd(getClient))
	plansCmd.AddCommand(newListCmd(getClient))
	plansCmd.AddCommand(newUpdateCmd(getClient))
	plansCmd.AddCommand(newCloseCmd(getClient))
	plansCmd.AddCommand(newDeleteCmd(getClient))
	plansCmd.AddCommand(newEntryCmd(getClient))

	root.AddCommand(plansCmd)
}
