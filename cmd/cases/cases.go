// Package cases реализует CLI команды для работы с тест-кейсами TestRail
package cases

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register регистрирует все команды для работы с тест-кейсами
func Register(root *cobra.Command, getClient GetClientFunc) {
	casesCmd := &cobra.Command{
		Use:   "cases",
		Short: "Управление тест-кейсами",
		Long: `Управление тест-кейсами: создание, чтение, обновление, удаление и массовые операции.

Основные операции:
  • add    — создать новый тест-кейс
  • get    — получить тест-кейс по ID
  • list   — список тест-кейсов с фильтрами
  • update — обновить тест-кейс
  • delete — удалить тест-кейс
  • bulk   — массовые операции (update/delete/copy/move)`,
	}

	// Добавление подкоманд
	casesCmd.AddCommand(newAddCmd(getClient))
	casesCmd.AddCommand(newGetCmd(getClient))
	casesCmd.AddCommand(newListCmd(getClient))
	casesCmd.AddCommand(newUpdateCmd(getClient))
	casesCmd.AddCommand(newDeleteCmd(getClient))
	casesCmd.AddCommand(newBulkCmd(getClient))

	root.AddCommand(casesCmd)
}

