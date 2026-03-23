package get

import (
	"context"
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
)

// newSuitesCmd создаёт команду для получения списка сьют проекта
func newSuitesCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
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
			cli := getClient(command)
			ctx := command.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
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
				projectID, err = interactive.SelectProject(ctx, interactive.PrompterFromContext(ctx), cli, "")
				if err != nil {
					return err
				}
			} else {
				projectID, err = flags.ParseID(projectIDStr)
				if err != nil {
					return fmt.Errorf("invalid project_id: %w", err)
				}
			}

			suites, err := runGetStatus(command, "Loading suites...", func(ctx context.Context) (any, error) {
				return cli.GetSuites(ctx, projectID)
			})
			if err != nil {
				return err
			}

			return handleOutput(command, suites, start)
		},
	}

	cmd.Flags().String("project-id", "", "ID проекта (альтернатива позиционному аргументу)")

	return cmd
}

// newSuiteCmd создаёт команду для получения одной сьюты
func newSuiteCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "suite <suite-id>",
		Short: "Получить одну тест-сюиту по ID сюиты",
		Args:  cobra.ExactArgs(1),
		Long: `Получить информацию о конкретной тест-сюите по её ID.

Пример:
	gotr get suite 20069
`,
		RunE: func(command *cobra.Command, args []string) error {
			start := time.Now()
			cli := getClient(command)
			ctx := command.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			idStr := args[0]
			id, err := flags.ParseID(idStr)
			if err != nil {
				return fmt.Errorf("invalid suite ID: %w", err)
			}

			suite, err := cli.GetSuite(ctx, id)
			if err != nil {
				return err
			}

			return handleOutput(command, suite, start)
		},
	}
}

// suitesCmd — экспортированная команда для регистрации
var suitesCmd = newSuitesCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})

// suiteCmd — экспортированная команда для регистрации
var suiteCmd = newSuiteCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClient(cmd)
})
