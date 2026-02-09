// Package plans реализует CLI команды для работы с тест-планами TestRail
package plans

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register регистрирует все команды для работы с тест-планами
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

	// Добавление подкоманд
	plansCmd.AddCommand(newAddCmd(getClient))
	plansCmd.AddCommand(newGetCmd(getClient))
	plansCmd.AddCommand(newListCmd(getClient))
	plansCmd.AddCommand(newUpdateCmd(getClient))
	plansCmd.AddCommand(newCloseCmd(getClient))
	plansCmd.AddCommand(newDeleteCmd(getClient))
	plansCmd.AddCommand(newEntryCmd(getClient))

	root.AddCommand(plansCmd)
}
