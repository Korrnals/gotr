// Package variables реализует CLI команды для работы с переменными TestRail
package variables

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register регистрирует все команды для работы с переменными
func Register(root *cobra.Command, getClient GetClientFunc) {
	variablesCmd := &cobra.Command{
		Use:   "variables",
		Short: "Управление переменными тест-кейсов",
		Long: `Управление переменными (variables) — конфигурационными значениями
для тест-кейсов.

Переменные позволяют создавать гибкие тест-кейсы, которые могут
адаптироваться под разные условия без создания дубликатов.
Значения переменных можно изменять на уровне датасета.

Доступные операции:
  • list   — список переменных датасета
  • add    — создать переменную
  • update — обновить переменную
  • delete — удалить переменную`,
	}

	// Добавление подкоманд
	variablesCmd.AddCommand(newListCmd(getClient))
	variablesCmd.AddCommand(newAddCmd(getClient))
	variablesCmd.AddCommand(newUpdateCmd(getClient))
	variablesCmd.AddCommand(newDeleteCmd(getClient))

	root.AddCommand(variablesCmd)
}
