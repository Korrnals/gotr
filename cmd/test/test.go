// Package test provides CLI commands for managing tests in TestRail
package test

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining a client.
type GetClientFunc = client.GetClientFunc

// Cmd is the parent command for managing tests.
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

// clientAccessor is the global accessor for obtaining a client.
var clientAccessor *client.Accessor

// getClientInterface returns the client as ClientInterface.
func getClientInterface(cmd *cobra.Command) client.ClientInterface {
	if clientAccessor == nil {
		return nil
	}
	return clientAccessor.GetClientSafe(cmd)
}

// SetGetClientForTests sets the getClient function for testing.
func SetGetClientForTests(fn GetClientFunc) {
	if clientAccessor == nil {
		clientAccessor = client.NewAccessor(fn)
	} else {
		clientAccessor.SetClientForTests(fn)
	}
}

// Register registers the test command and all its subcommands.
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
	clientAccessor = client.NewAccessor(clientFn)
	rootCmd.AddCommand(Cmd)

	// Create and register subcommands using constructors.
	// Flags are defined inside the constructors.
	getCmd := newGetCmd(getClientInterface)
	listCmd := newListCmd(getClientInterface)

	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(listCmd)
}
