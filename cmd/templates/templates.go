// Package templates implements CLI commands for managing TestRail case templates.
package templates

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining an API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register adds all template-related subcommands to the root command.
func Register(root *cobra.Command, getClient GetClientFunc) {
	templatesCmd := &cobra.Command{
		Use:   "templates",
		Short: "Управление шаблонами тест-кейсов",
		Long: `Управление шаблонами (templates) — форматами отображения тест-кейсов.

Шаблоны определяют структуру и поля, доступные при создании
и редактировании тест-кейсов in project.

Доступные операции:
  • list   — список шаблонов проекта`,
	}

	// Register subcommands
	templatesCmd.AddCommand(newListCmd(getClient))

	root.AddCommand(templatesCmd)
}
