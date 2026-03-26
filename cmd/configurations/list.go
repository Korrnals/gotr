package configurations

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'configurations list'
// Эндпоинт: GET /get_configs/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project_id]",
		Short: "Список конфигураций проекта",
		Long: `Выводит список конфигураций, доступных в указанном проекте.

Конфигурации представляют собой тестовые среды (браузеры, ОС, устройства)
и группируются по типам. Используются при создании тест-планов
с множественными конфигурациями.

Каждая конфигурация имеет ID, который используется для указания
в параметрах при создании записей плана с конфигурациями.`,
		Example: `  # Получить конфигурации проекта
  gotr configurations list 1

  # Сохранить в файл для анализа
  gotr configurations list 5 -o configs.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var projectID int64
			var err error
			if len(args) > 0 {
				projectID, err = flags.ValidateRequiredID(args, 0, "project_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr configurations list [project_id]")
				}
				if _, ok := interactive.PrompterFromContext(ctx).(*interactive.NonInteractivePrompter); ok {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr configurations list [project_id]")
				}

				projectID, err = resolveProjectIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			resp, err := cli.GetConfigs(ctx, projectID)
			if err != nil {
				return fmt.Errorf("failed to get configurations: %w", err)
			}

			return output.OutputResult(cmd, resp, "configurations")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
