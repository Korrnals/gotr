// Package labels implements CLI commands for managing TestRail test labels.
package labels

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining an API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all label management commands on the given root.
func Register(root *cobra.Command, getClient GetClientFunc) {
	labelsCmd := &cobra.Command{
		Use:   "labels",
		Short: "Управление метками тестов",
		Long: `Обновление меток (labels) для тестов и тестовых прогонов.

Метки позволяют категоризировать и группировать тесты для удобного анализа.
Можно обновлять метки как для одного теста, так и для всех тестов в прогоне.`,
	}

	// Add get and management subcommands
	labelsCmd.AddCommand(newGetCmd(getClient))
	labelsCmd.AddCommand(newListCmd(getClient))
	labelsCmd.AddCommand(newUpdateLabelCmd(getClient))

	// Create the parent 'update' command
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Обновить метки для тестов",
		Long: `Обновляет метки для одного теста или сразу для всех тестов в прогоне.

Доступные подкоманды:
  • test  — обновить метки одного теста по ID
  • tests — обновить метки всех тестов в прогоне`,
	}

	// Shared flags for all update subcommands
	updateCmd.PersistentFlags().Bool("dry-run", false, "Показать, что будет сделано без изменений")

	// Add subcommands to 'update'
	updateCmd.AddCommand(newUpdateTestCmd(getClient))
	updateCmd.AddCommand(newUpdateTestsCmd(getClient))

	// Attach 'update' to the labels command
	labelsCmd.AddCommand(updateCmd)

	root.AddCommand(labelsCmd)
}
