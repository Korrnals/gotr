// Package milestones implements CLI commands for managing TestRail milestones.
package milestones

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining an API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all milestone management commands on the given root.
func Register(root *cobra.Command, getClient GetClientFunc) {
	milestonesCmd := &cobra.Command{
		Use:   "milestones",
		Short: "Управление майлстонами (этапами) проекта",
		Long: `Управление майлстонами (вехами/этапами) проекта.

Майлстоны используются для группировки тестовых прогонов по этапам разработки
(например: "Релиз 1.0", "Спринт 5", "Бета-версия").

Доступные операции:
  • add    — создать новый майлстон
  • get    — получить информацию о майлстоне
  • list   — список всех майлстонов проекта  
  • update — обновить майлстон
  • delete — удалить майлстон`,
	}

	// Add subcommands
	milestonesCmd.AddCommand(newAddCmd(getClient))
	milestonesCmd.AddCommand(newGetCmd(getClient))
	milestonesCmd.AddCommand(newListCmd(getClient))
	milestonesCmd.AddCommand(newUpdateCmd(getClient))
	milestonesCmd.AddCommand(newDeleteCmd(getClient))

	root.AddCommand(milestonesCmd)
}
