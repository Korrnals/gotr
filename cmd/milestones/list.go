package milestones

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'milestones list'
// Эндпоинт: GET /get_milestones/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project_id]",
		Short: "Список майлстонов проекта",
		Long: `Выводит список всех майлстонов указанного проекта.

Показывает ID, название, статус завершения и дедлайны всех майлстонов.
Поддерживает вывод в JSON для автоматизации.`,
		Example: `  # Список майлстонов проекта
  gotr milestones list 1

  # Сохранить список в файл
  gotr milestones list 1 -o milestones.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projectID int64
			if len(args) > 0 {
				var err error
				projectID, err = flags.ValidateRequiredID(args, 0, "project_id")
				if err != nil {
					return err
				}
			} else {
				ctx := cmd.Context()
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr milestones list [project_id]")
				}
				cli := getClient(cmd)
				var err error
				projectID, err = resolveProjectIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.GetMilestones(ctx, projectID)
			if err != nil {
				return fmt.Errorf("failed to list milestones: %w", err)
			}

			return output.OutputResult(cmd, resp, "milestones")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
