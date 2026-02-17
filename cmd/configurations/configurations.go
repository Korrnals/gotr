// Package configurations реализует CLI команды для работы с конфигурациями TestRail
package configurations

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register регистрирует все команды для работы с конфигурациями
func Register(root *cobra.Command, getClient GetClientFunc) {
	configsCmd := &cobra.Command{
		Use:   "configurations",
		Short: "Управление конфигурациями тестовых сред",
		Long: `Управление конфигурациями (configurations) — тестовыми средами для прогонов.

Конфигурации представляют собой различные среды для тестирования:
  • Браузеры (Chrome, Firefox, Safari)
  • Операционные системы (Windows, macOS, Linux)
  • Мобильные устройства (iOS, Android)
  • Версии ПО и другие параметры

Конфигурации организованы в группы (например: "Browsers", "OS").
Каждая группа содержит отдельные конфигурации (например: "Chrome", "Firefox").

Доступные операции:
  • list          — список конфигураций проекта
  • add-group     — создать группу конфигураций
  • add-config    — добавить конфигурацию в группу
  • update-group  — обновить группу
  • update-config — обновить конфигурацию
  • delete-group  — удалить группу
  • delete-config — удалить конфигурацию`,
	}

	// Добавление подкоманд
	configsCmd.AddCommand(newListCmd(getClient))
	configsCmd.AddCommand(newAddGroupCmd(getClient))
	configsCmd.AddCommand(newAddConfigCmd(getClient))
	configsCmd.AddCommand(newUpdateGroupCmd(getClient))
	configsCmd.AddCommand(newUpdateConfigCmd(getClient))
	configsCmd.AddCommand(newDeleteGroupCmd(getClient))
	configsCmd.AddCommand(newDeleteConfigCmd(getClient))

	root.AddCommand(configsCmd)
}
