// Package templates реализует CLI команды для работы с шаблонами TestRail
package templates

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register регистрирует все команды для работы с шаблонами
func Register(root *cobra.Command, getClient GetClientFunc) {
	templatesCmd := &cobra.Command{
		Use:   "templates",
		Short: "Управление шаблонами тест-кейсов",
		Long: `Управление шаблонами (templates) — форматами отображения тест-кейсов.

Шаблоны определяют структуру и поля, доступные при создании
и редактировании тест-кейсов в проекте.

Доступные операции:
  • list   — список шаблонов проекта`,
	}

	// Добавление подкоманд
	templatesCmd.AddCommand(newListCmd(getClient))

	root.AddCommand(templatesCmd)
}
