package get

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/progress"
	"github.com/spf13/cobra"
)

// newSharedStepsCmd создаёт команду для получения shared steps проекта
func newSharedStepsCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
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
			cli := getClient(command)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

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
				projectID, err = selectProjectInteractively(cli)
				if err != nil {
					return err
				}
			} else {
				projectID, err = strconv.ParseInt(projectIDStr, 10, 64)
				if err != nil {
					return fmt.Errorf("некорректный ID проекта: %w", err)
				}
			}

			// Create progress manager and spinner
			pm := progress.NewManager()
			spinner := pm.NewSpinner("")
			spinner.Describe("Загрузка shared steps...")

			steps, err := cli.GetSharedSteps(projectID)
			if err != nil {
				return err
			}

			spinner.Finish()
			return handleOutput(command, steps, start)
		},
	}

	cmd.Flags().String("project-id", "", "ID проекта (альтернатива позиционному аргументу)")

	return cmd
}

// newSharedStepCmd создаёт команду для получения одного shared step
func newSharedStepCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "sharedstep <step-id>",
		Short: "Получить один shared step по ID шага",
		Long:  "Получает детальную информацию о конкретном shared step по его ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			idStr := args[0]
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return fmt.Errorf("некорректный ID шага: %w", err)
			}

			step, err := cli.GetSharedStep(id)
			if err != nil {
				return err
			}

			return handleOutput(command, step, start)
		},
	}
}

// sharedStepsCmd — экспортированная команда для регистрации
var sharedStepsCmd = newSharedStepsCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})

// sharedStepCmd — экспортированная команда для регистрации
var sharedStepCmd = newSharedStepCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})
