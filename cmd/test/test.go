// Package test provides CLI commands for managing tests in TestRail
package test

import (
	"github.com/Korrnals/gotr/cmd/common"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc = common.GetClientFunc

// Cmd — родительская команда для управления тестами
var Cmd = &cobra.Command{
	Use:   "test",
	Short: "Управление тестами в TestRail",
	Long: `Команды для получения и управления тестами (tests) в TestRail.

Test — это конкретный экземпляр тест-кейса в рамках тест-рана.
Каждый тест имеет статус (passed, failed, blocked и т.д.) и может быть 
назначен на конкретного пользователя.

Подкоманды:
	get     — получить информацию о тесте по ID
	list    — получить список тестов в ране

Примеры:
	# Получить информацию о тесте
	gotr test get 12345

	# Получить список тестов в ране
	gotr test list 100

	# Получить только failed тесты
	gotr test list 100 --status-id 5

	# Получить тесты, назначенные на пользователя
	gotr test list 100 --assigned-to 10
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

// Register регистрирует команду test и все её подкоманды
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
	clientAccessor = common.NewClientAccessor(clientFn)
	rootCmd.AddCommand(Cmd)

	// Добавляем подкоманды
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(listCmd)

	// Общие флаги для всех подкоманд
	for _, subCmd := range Cmd.Commands() {
		subCmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл")
		subCmd.Flags().BoolP("quiet", "q", false, "Тихий режим")
	}

	// Флаги для list
	listCmd.Flags().Int64("status-id", 0, "Фильтр по ID статуса")
	listCmd.Flags().Int64("assigned-to", 0, "Фильтр по ID назначенного пользователя")
}
