package run

import (
	"github.com/Korrnals/gotr/cmd/common"
	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc = common.GetClientFunc

// Cmd — родительская команда для управления test runs
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

// Register регистрирует команду run и все её подкоманды
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
	clientAccessor = common.NewClientAccessor(clientFn)
	rootCmd.AddCommand(Cmd)

	// Добавляем подкоманды
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(closeCmd)
	Cmd.AddCommand(deleteCmd)

	// Общие флаги для всех подкоманд
	for _, subCmd := range Cmd.Commands() {
		save.AddFlag(subCmd)
		subCmd.Flags().BoolP("quiet", "q", false, "Тихий режим")
	}

	// Mark required flags for create (already defined in constructor)
	createCmd.MarkFlagRequired("suite-id")
	createCmd.MarkFlagRequired("name")
}
