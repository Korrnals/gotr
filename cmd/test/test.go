// Package test provides CLI commands for managing tests in TestRail
package test

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc = client.GetClientFunc

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

// clientAccessor — глобальный accessor для получения клиента
var clientAccessor *client.Accessor

// getClientInterface возвращает клиент как ClientInterface
func getClientInterface(cmd *cobra.Command) client.ClientInterface {
	if clientAccessor == nil {
		return nil
	}
	return clientAccessor.GetClientSafe(cmd)
}

// SetGetClientForTests устанавливает getClient для тестов
func SetGetClientForTests(fn GetClientFunc) {
	if clientAccessor == nil {
		clientAccessor = client.NewAccessor(fn)
	} else {
		clientAccessor.SetClientForTests(fn)
	}
}

// Register регистрирует команду test и все её подкоманды
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
	clientAccessor = client.NewAccessor(clientFn)
	rootCmd.AddCommand(Cmd)

	// Создаём и добавляем подкоманды используя конструкторы
	// Флаги определяются внутри конструкторов
	getCmd := newGetCmd(getClientInterface)
	listCmd := newListCmd(getClientInterface)

	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(listCmd)
}
