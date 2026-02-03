package get

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// suitesCmd — подкоманда для списка тест-сюит проекта
var suitesCmd = &cobra.Command{
	Use:   "suites [project-id]",
	Short: "Получить тест-сюиты проекта",
	Long: `Получить тест-сюиты проекта.

Если ID проекта не указан, будет предложено выбрать проект из списка.

Примеры:
	# Интерактивный выбор проекта
	gotr get suites

	# Явное указание проекта
	gotr get suites 30
	gotr get suites --project-id 30
`,
	RunE: func(command *cobra.Command, args []string) error {
		start := time.Now()
		client := getClient(command)

		// Получаем ID проекта
		projectIDStr := ""
		if len(args) > 0 {
			projectIDStr = args[0]
		}
		if pid, _ := command.Flags().GetString("project-id"); pid != "" {
			projectIDStr = pid
		}

		var projectID int64
		var err error

		if projectIDStr == "" {
			// Интерактивный выбор проекта
			projectID, err = selectProjectInteractively(client)
			if err != nil {
				return err
			}
		} else {
			projectID, err = strconv.ParseInt(projectIDStr, 10, 64)
			if err != nil {
				return fmt.Errorf("некорректный ID проекта: %w", err)
			}
		}

		suites, err := client.GetSuites(projectID)
		if err != nil {
			return err
		}

		return handleOutput(command, suites, start)
	},
}

// suiteCmd — подкоманда для одной тест-сюиты по ID
var suiteCmd = &cobra.Command{
	Use:   "suite <suite-id>",
	Short: "Получить одну тест-сюиту по ID сюиты",
	Args:  cobra.ExactArgs(1),
	Long: `Получить информацию о конкретной тест-сюите по её ID.

Пример:
	gotr get suite 20069
`,
	RunE: func(command *cobra.Command, args []string) error {
		start := time.Now()
		client := getClient(command)

		idStr := args[0]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректный ID сюиты: %w", err)
		}

		suite, err := client.GetSuite(id)
		if err != nil {
			return err
		}

		return handleOutput(command, suite, start)
	},
}
