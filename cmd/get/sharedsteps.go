package get

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// sharedStepsCmd — подкоманда для списка shared steps проекта
var sharedStepsCmd = &cobra.Command{
	Use:   "sharedsteps [project-id]",
	Short: "Получить shared steps проекта",
	Long: `Получить shared steps (общие шаги) проекта.

Если ID проекта не указан, будет предложено выбрать проект из списка.

Примеры:
	# Интерактивный выбор проекта
	gotr get sharedsteps

	# Явное указание проекта
	gotr get sharedsteps 30
	gotr get sharedsteps --project-id 30
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

		steps, err := client.GetSharedSteps(projectID)
		if err != nil {
			return err
		}

		return handleOutput(command, steps, start)
	},
}

// sharedStepCmd — подкоманда для одного shared step
var sharedStepCmd = &cobra.Command{
	Use:   "sharedstep <step-id>",
	Short: "Получить один shared step по ID шага",
	Long:  "Получает детальную информацию о конкретном shared step по его ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		start := time.Now()
		client := getClient(command)

		idStr := args[0]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("некорректный ID шага: %w", err)
		}

		step, err := client.GetSharedStep(id)
		if err != nil {
			return err
		}

		return handleOutput(command, step, start)
	},
}
