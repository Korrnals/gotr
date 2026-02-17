// Package tests реализует CLI команды для работы с тестами TestRail
package tests

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register регистрирует все команды для работы с тестами
func Register(root *cobra.Command, getClient GetClientFunc) {
	testsCmd := &cobra.Command{
		Use:   "tests",
		Short: "Управление тестами",
		Long: `Управление тестами (tests) — результатами выполнения тест-кейсов
в тестовых прогонах.

Тест представляет собой конкретное выполнение тест-кейса в рамках
тестового прогона с определённым статусом и результатом.

Доступные операции:
  • update — обновить тест (статус, комментарий, время)`,
	}

	testsCmd.AddCommand(newUpdateCmd(getClient))
	testsCmd.AddCommand(newGetCmd(getClient))
	testsCmd.AddCommand(newListCmd(getClient))

	root.AddCommand(testsCmd)
}
