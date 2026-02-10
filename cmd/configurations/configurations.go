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

Конфигурации привязаны к проекту и используются при создании тест-планов
с множественными конфигурациями для тестирования одного набора кейсов
в разных средах.

Доступные операции:
  • list — список конфигураций проекта`,
	}

	// Добавление подкоманд
	configsCmd.AddCommand(newListCmd(getClient))

	root.AddCommand(configsCmd)
}
