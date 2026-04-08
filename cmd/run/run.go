package run

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining a client.
type GetClientFunc = client.GetClientFunc

// Cmd is the parent command for managing test runs.
var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Управление test runs в TestRail",
	Long: `Команды для управления test runs (тестовыми прогонами) в TestRail.

Test run — это экземпляр тест-сюиты, запущенный для выполнения тестов.

Подкоманды:
	get     — получить информацию о test run по ID
	list    — получить список test runs проекта
	create  — создать новый test run
	update  — обновить существующий test run
	close   — закрыть test run (завершить)
	delete  — удалить test run

Примеры:
	# Получить информацию о test run
	gotr run get 12345

	# Получить список runs проекта
	gotr run list 30

	# Создать новый test run
	gotr run create 30 --name "Smoke Tests v2.0" --suite-id 20069

	# Закрыть test run
	gotr run close 12345
`,
}

var clientAccessor *client.Accessor

// SetGetClientForTests sets getClient for tests.
func SetGetClientForTests(fn GetClientFunc) {
	if clientAccessor == nil {
		clientAccessor = client.NewAccessor(fn)
	} else {
		clientAccessor.SetClientForTests(fn)
	}
}

// getClientSafe safely calls getClient with a nil check.
func getClientSafe(cmd *cobra.Command) client.ClientInterface {
	if clientAccessor == nil {
		return nil
	}
	return clientAccessor.GetClientSafe(cmd.Context())
}

// Register registers the run command and all its subcommands.
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
	clientAccessor = client.NewAccessor(clientFn)
	rootCmd.AddCommand(Cmd)

	// Add subcommands
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(closeCmd)
	Cmd.AddCommand(deleteCmd)

	// Common flags for all subcommands
	for _, subCmd := range Cmd.Commands() {
		output.AddFlag(subCmd)
	}

	// Mark required flags for create (already defined in constructor)
	_ = createCmd.MarkFlagRequired("name")
}
