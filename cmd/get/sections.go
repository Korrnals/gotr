package get

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/progress"
	"github.com/spf13/cobra"
)

// newSectionsCmd создаёт родительскую команду для sections
func newSectionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sections",
		Short: "Управление секциями",
		Long: `Управление секциями (sections) — организационными единицами
внутри тест-сюитов для группировки тест-кейсов.

Секции позволяют структурировать тесты в иерархию для удобной навигации
и организации тестовой документации.

Доступные операции:
  • get    — получить информацию о секции по ID
  • list   — список секций проекта/сюиты`,
	}
}

// newSectionGetCmd создаёт команду для получения одной секции
func newSectionGetCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "section <section_id>",
		Short: "Получить информацию о секции",
		Long:  `Получить детальную информацию о секции по её ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			sectionID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || sectionID <= 0 {
				return fmt.Errorf("section_id должен быть положительным числом")
			}

			// Create progress manager and spinner
			pm := progress.NewManager()
			spinner := pm.NewSpinner("")
			spinner.Describe("Загрузка секции...")

			section, err := cli.GetSection(sectionID)
			if err != nil {
				return err
			}

			spinner.Finish()
			return handleOutput(command, section, start)
		},
	}
}

// newSectionsListCmd создаёт команду для получения списка секций
func newSectionsListCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project_id>",
		Short: "Получить список секций проекта",
		Long: `Получить список всех секций для указанного проекта.

Для фильтрации по конкретной сюите используйте флаг --suite-id.`,
		Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			projectID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || projectID <= 0 {
				return fmt.Errorf("project_id должен быть положительным числом")
			}

			suiteID, _ := command.Flags().GetInt64("suite-id")

			// Create progress manager and spinner
			pm := progress.NewManager()
			spinner := pm.NewSpinner("")
			spinner.Describe("Загрузка секций...")

			sections, err := cli.GetSections(projectID, suiteID)
			if err != nil {
				return err
			}

			spinner.Finish()
			return handleOutput(command, sections, start)
		},
	}

	// Флаг для фильтрации по suite_id
	cmd.Flags().Int64P("suite-id", "s", 0, "ID сюиты для фильтрации")

	return cmd
}

// sectionsCmd — экспортированная родительская команда
var sectionsCmd = newSectionsCmd()

// sectionGetCmd — экспортированная команда для регистрации
var sectionGetCmd = newSectionGetCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})

// sectionsListCmd — экспортированная команда для регистрации
var sectionsListCmd = newSectionsListCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})

func init() {
	// Добавляем подкоманды к родительской
	sectionsCmd.AddCommand(sectionGetCmd)
	sectionsCmd.AddCommand(sectionsListCmd)
}
