package result

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'result list'
// Эндпоинт: GET /get_results_for_run/{run_id}
func newListCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "list [run-id]",
		Short: "Получить результаты для test run",
		Long: `Получает список результатов для указанного test run.

Если run-id не указан, будет предложен интерактивный выбор:
1. Выбор проекта из списка
2. Выбор test run из проекта

Примеры:
	# Получить результаты с интерактивным выбором run
	gotr result list

	# Получить результаты для конкретного run
	gotr result list 12345

	# Сохранить в файл
	gotr result list 12345 -o results.json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			svc := newResultServiceFromInterface(cli)

			var runID int64
			var err error

			if len(args) > 0 {
				// Явно указан run-id
				runID, err = flags.ValidateRequiredID(args, 0, "run")
				if err != nil {
					return err
				}
			} else {
				// Интерактивный выбор: проект → run
				p := interactive.PrompterFromContext(ctx)
				projectID, err := interactive.SelectProject(ctx, p, cli, "")
				if err != nil {
					return err
				}

				// Получаем список runs проекта
				runs, err := svc.GetRunsForProject(ctx, projectID)
				if err != nil {
					return fmt.Errorf("failed to get runs list: %w", err)
				}

				if len(runs) == 0 {
					return fmt.Errorf("no test runs found in project %d", projectID)
				}

				// Выбираем run интерактивно
				runID, err = interactive.SelectRun(ctx, p, runs, "")
				if err != nil {
					return err
				}
			}

			results, err := svc.GetForRun(ctx, runID)
			if err != nil {
				return fmt.Errorf("failed to get results: %w", err)
			}

			return svc.Output(ctx, cmd, results)
		},
	}
}

// Обратная совместимость: глобальная переменная для использования в result.go
var listCmd = newListCmd(func(cmd *cobra.Command) client.ClientInterface { return getClientSafe(cmd) })
