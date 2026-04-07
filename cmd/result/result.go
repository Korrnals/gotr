package result

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining an HTTP client.
type GetClientFunc = client.GetClientFunc

// Cmd is the parent command for managing test results.
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
}

var clientAccessor *client.Accessor

// SetGetClientForTests overrides the client accessor for testing.
func SetGetClientForTests(fn GetClientFunc) {
	if clientAccessor == nil {
		clientAccessor = client.NewAccessor(fn)
	} else {
		clientAccessor.SetClientForTests(fn)
	}
}

// getClientSafe safely calls getClient with a nil guard.
func getClientSafe(cmd *cobra.Command) *client.HTTPClient {
	if clientAccessor == nil {
		return nil
	}
	return clientAccessor.GetClientSafe(cmd)
}

// Register adds the result command and all its subcommands to rootCmd.
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
	clientAccessor = client.NewAccessor(clientFn)
	rootCmd.AddCommand(Cmd)

	// Register subcommands
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(getCaseCmd)
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(addCaseCmd)
	Cmd.AddCommand(addBulkCmd)
	Cmd.AddCommand(fieldsCmd)

	// Shared flags for all subcommands
	for _, subCmd := range Cmd.Commands() {
		output.AddFlag(subCmd)
	}

	// Mark required flags (already defined in constructors)
	addCmd.MarkFlagRequired("status-id")
	addCaseCmd.MarkFlagRequired("case-id")
	addCaseCmd.MarkFlagRequired("status-id")
	addBulkCmd.MarkFlagRequired("results-file")
}
