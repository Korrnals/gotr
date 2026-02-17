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

	// Add persistent flags FIRST (before subcommands) for completion to work
	Cmd.PersistentFlags().StringP("pid1", "1", "", "ID первого проекта (обязательно)")
	Cmd.PersistentFlags().StringP("pid2", "2", "", "ID второго проекта (обязательно)")
	Cmd.PersistentFlags().StringP("format", "f", "table", "Формат вывода: table, json, yaml, csv")
	Cmd.PersistentFlags().Bool("save", false, "Сохранить результат в файл (по умолчанию в ~/.gotr/exports/)")
	Cmd.PersistentFlags().String("save-to", "", "Сохранить результат в указанный файл")

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
