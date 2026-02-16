package compare

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientInterfaceFunc — тип функции для получения клиента
type GetClientInterfaceFunc func(cmd *cobra.Command) client.ClientInterface

var getClient GetClientInterfaceFunc

// Cmd — основная команда для сравнения (заполняется в Register)
var Cmd *cobra.Command

// SetGetClientForTests устанавливает getClient для тестов
func SetGetClientForTests(fn GetClientInterfaceFunc) {
	getClient = fn
}

// Register регистрирует команду compare и все её подкоманды
func Register(rootCmd *cobra.Command, clientFn GetClientInterfaceFunc) {
	getClient = clientFn

	// Create main compare command
	Cmd = &cobra.Command{
		Use:   "compare",
		Short: "Сравнение данных между проектами",
		Long: `Выполнение сравнения ресурсов между двумя проектами.

Поддерживаемые ресурсы:
  cases          - сравнить тест-кейсы
  suites         - сравнить тест-сюиты
  sections       - сравнить секции
  sharedsteps    - сравнить shared steps
  runs           - сравнить test runs
  plans          - сравнить test plans
  milestones     - сравнить milestones
  datasets       - сравнить datasets
  groups         - сравнить группы
  labels         - сравнить метки
  templates      - сравнить шаблоны
  configurations - сравнить конфигурации
  all            - сравнить все ресурсы

Примеры:
  gotr compare cases --pid1 30 --pid2 31
  gotr compare all --pid1 30 --pid2 31 --save
  gotr compare all --pid1 30 --pid2 31 --save-to result.json
`,
	}

	// Add all subcommands
	Cmd.AddCommand(casesCmd)
	Cmd.AddCommand(suitesCmd)
	Cmd.AddCommand(sectionsCmd)
	Cmd.AddCommand(sharedStepsCmd)
	Cmd.AddCommand(runsCmd)
	Cmd.AddCommand(plansCmd)
	Cmd.AddCommand(milestonesCmd)
	Cmd.AddCommand(datasetsCmd)
	Cmd.AddCommand(groupsCmd)
	Cmd.AddCommand(labelsCmd)
	Cmd.AddCommand(templatesCmd)
	Cmd.AddCommand(configurationsCmd)
	Cmd.AddCommand(allCmd)

	rootCmd.AddCommand(Cmd)
}
