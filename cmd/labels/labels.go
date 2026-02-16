// Package labels реализует CLI команды для работы с метками тестов TestRail
package labels

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register регистрирует все команды для работы с метками
func Register(root *cobra.Command, getClient GetClientFunc) {
	labelsCmd := &cobra.Command{
		Use:   "labels",
		Short: "Управление метками тестов",
		Long: `Обновление меток (labels) для тестов и тестовых прогонов.

Метки позволяют категоризировать и группировать тесты для удобного анализа.
Можно обновлять метки как для одного теста, так и для всех тестов в прогоне.`,
	}

	// Добавление команд получения и управления метками
	labelsCmd.AddCommand(newGetCmd(getClient))
	labelsCmd.AddCommand(newListCmd(getClient))
	labelsCmd.AddCommand(newUpdateLabelCmd(getClient))

	// Создание родительской команды 'update'
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Обновить метки для тестов",
		Long: `Обновляет метки для одного теста или сразу для всех тестов в прогоне.

Доступные подкоманды:
  • test  — обновить метки одного теста по ID
  • tests — обновить метки всех тестов в прогоне`,
	}

	// Общие флаги для всех подкоманд update
	updateCmd.PersistentFlags().Bool("dry-run", false, "Показать, что будет сделано без изменений")

	// Добавление подкоманд к 'update'
	updateCmd.AddCommand(newUpdateTestCmd(getClient))
	updateCmd.AddCommand(newUpdateTestsCmd(getClient))

	// Добавление 'update' в labels
	labelsCmd.AddCommand(updateCmd)

	root.AddCommand(labelsCmd)
}
