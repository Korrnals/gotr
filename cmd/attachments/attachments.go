// Package attachments реализует CLI команды для работы с вложениями TestRail
package attachments

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register регистрирует все команды для работы с вложениями
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

	// Создание родительской команды 'add'
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Добавить вложение к ресурсу",
		Long: `Загружает файл и прикрепляет его к указанному ресурсу.

Поддерживаются различные типы ресурсов: тест-кейс, план, запись плана,
результат теста или тестовый прогон.`,
	}

	// Общие флаги для всех подкоманд add
	addCmd.PersistentFlags().Bool("dry-run", false, "Показать, что будет сделано без загрузки файла")
	addCmd.PersistentFlags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")

	// Добавление подкоманд к 'add'
	addCmd.AddCommand(newAddCaseCmd(getClient))
	addCmd.AddCommand(newAddPlanCmd(getClient))
	addCmd.AddCommand(newAddPlanEntryCmd(getClient))
	addCmd.AddCommand(newAddResultCmd(getClient))
	addCmd.AddCommand(newAddRunCmd(getClient))

	// Добавление 'add' в attachments
	attachmentsCmd.AddCommand(addCmd)

	root.AddCommand(attachmentsCmd)
}
