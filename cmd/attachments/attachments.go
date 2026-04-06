// Package attachments implements CLI commands for managing TestRail attachments.
package attachments

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining a client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all attachment-related commands.
func Register(root *cobra.Command, getClient GetClientFunc) {
	attachmentsCmd := &cobra.Command{
		Use:   "attachments",
		Short: "Управление файловыми вложениями",
		Long: `Управление файловыми вложениями к тест-кейсам, планам, результатам и прогонам.

Поддерживаемые типы ресурсов для прикрепления файлов:
  • case       — вложение к тест-кейсу
  • plan       — вложение к тест-плану
  • plan-entry — вложение к записи плана
  • result     — вложение к результату теста
  • run        — вложение к тестовому прогону`,
	}

	// Create the parent 'add' command
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Добавить вложение к ресурсу",
		Long: `Загружает файл и прикрепляет его к указанному ресурсу.

Поддерживаются различные типы ресурсов: тест-кейс, план, запись плана,
результат теста или тестовый прогон.`,
	}

	// Shared flags for all 'add' subcommands
	addCmd.PersistentFlags().Bool("dry-run", false, "Показать, что будет сделано без загрузки файла")
	output.AddFlag(addCmd)

	// Register subcommands under 'add'
	addCmd.AddCommand(newAddCaseCmd(getClient))
	addCmd.AddCommand(newAddPlanCmd(getClient))
	addCmd.AddCommand(newAddPlanEntryCmd(getClient))
	addCmd.AddCommand(newAddResultCmd(getClient))
	addCmd.AddCommand(newAddRunCmd(getClient))

	// Add 'add' to the attachments command
	attachmentsCmd.AddCommand(addCmd)

	root.AddCommand(attachmentsCmd)
}
