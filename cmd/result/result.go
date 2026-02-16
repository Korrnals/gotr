package result

import (
	"github.com/Korrnals/gotr/cmd/common"
	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc = common.GetClientFunc

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

var clientAccessor *common.ClientAccessor

// SetGetClientForTests устанавливает getClient для тестов
func SetGetClientForTests(fn GetClientFunc) {
	if clientAccessor == nil {
		clientAccessor = common.NewClientAccessor(fn)
	} else {
		clientAccessor.SetClientForTests(fn)
	}
}

// getClientSafe безопасно вызывает getClient с проверкой на nil
func getClientSafe(cmd *cobra.Command) *client.HTTPClient {
	if clientAccessor == nil {
		return nil
	}
	return clientAccessor.GetClientSafe(cmd)
}

// Register регистрирует команду result и все её подкоманды
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
	clientAccessor = common.NewClientAccessor(clientFn)
	rootCmd.AddCommand(Cmd)

	// Добавляем подкоманды
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(getCaseCmd)
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(addCaseCmd)
	Cmd.AddCommand(addBulkCmd)
	Cmd.AddCommand(fieldsCmd)

	// Общие флаги для всех подкоманд
	for _, subCmd := range Cmd.Commands() {
		save.AddFlag(subCmd)
		subCmd.Flags().BoolP("quiet", "q", false, "Тихий режим")
	}

	// Mark required flags (already defined in constructors)
	addCmd.MarkFlagRequired("status-id")
	addCaseCmd.MarkFlagRequired("case-id")
	addCaseCmd.MarkFlagRequired("status-id")
	addBulkCmd.MarkFlagRequired("results-file")
}
