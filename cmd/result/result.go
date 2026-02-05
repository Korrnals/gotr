package result

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) *client.HTTPClient

// Cmd — родительская команда для управления результатами тестов
var Cmd = &cobra.Command{
	Use:   "result",
	Short: "Управление результатами тестов в TestRail",
	Long: `Команды для добавления и получения результатов тестов (test results) в TestRail.

Test result — это результат выполнения отдельного теста (passed, failed, blocked и т.д.)

Подкоманды:
	list       — получить результаты для test run (с интерактивным выбором)
	get        — получить результаты для test
	get-case   — получить результаты для кейса в run
	add        — добавить результат для test
	add-case   — добавить результат для кейса в run
	add-bulk   — массовое добавление результатов

Примеры:
	# Получить результаты с интерактивным выбором run
	gotr result list

	# Получить результаты для конкретного run
	gotr result list 12345

	# Получить результаты test
	gotr result get 12345

	# Добавить passed результат
	gotr result add 12345 --status-id 1 --comment "Test passed successfully"

	# Добавить failed результат с дефектом
	gotr result add 12345 --status-id 5 --comment "Found bug" --defects "BUG-123"
`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var getClient GetClientFunc

// SetGetClientForTests устанавливает getClient для тестов
func SetGetClientForTests(fn GetClientFunc) {
	getClient = fn
}

// getClientSafe безопасно вызывает getClient с проверкой на nil
func getClientSafe(cmd *cobra.Command) *client.HTTPClient {
	if getClient == nil {
		return nil
	}
	return getClient(cmd)
}

// Register регистрирует команду result и все её подкоманды
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
	getClient = clientFn
	rootCmd.AddCommand(Cmd)

	// Добавляем подкоманды
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(getCaseCmd)
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(addCaseCmd)
	Cmd.AddCommand(addBulkCmd)

	// Общие флаги для всех подкоманд
	for _, subCmd := range Cmd.Commands() {
		subCmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл")
		subCmd.Flags().BoolP("quiet", "q", false, "Тихий режим")
	}

	// Флаги для add
	addCmd.Flags().Int64("status-id", 0, "ID статуса результата (обязательный)")
	addCmd.Flags().String("comment", "", "Комментарий к результату")
	addCmd.Flags().String("version", "", "Версия ПО")
	addCmd.Flags().String("elapsed", "", "Затраченное время (например: '1m 30s')")
	addCmd.Flags().String("defects", "", "ID дефектов (через запятую)")
	addCmd.Flags().Int64("assigned-to", 0, "ID пользователя для назначения")
	addCmd.MarkFlagRequired("status-id")

	// Флаги для add-case (те же + case-id)
	addCaseCmd.Flags().Int64("case-id", 0, "ID тест-кейса (обязательный)")
	addCaseCmd.Flags().Int64("status-id", 0, "ID статуса результата (обязательный)")
	addCaseCmd.Flags().String("comment", "", "Комментарий к результату")
	addCaseCmd.Flags().String("version", "", "Версия ПО")
	addCaseCmd.Flags().String("elapsed", "", "Затраченное время")
	addCaseCmd.Flags().String("defects", "", "ID дефектов (через запятую)")
	addCaseCmd.Flags().Int64("assigned-to", 0, "ID пользователя для назначения")
	addCaseCmd.MarkFlagRequired("case-id")
	addCaseCmd.MarkFlagRequired("status-id")

	// Флаги для add-bulk
	addBulkCmd.Flags().String("results-file", "", "JSON файл с результатами (обязательный)")
	addBulkCmd.MarkFlagRequired("results-file")
}
