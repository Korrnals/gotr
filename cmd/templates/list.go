package templates

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'templates list'
// Эндпоинт: GET /get_templates/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project_id]",
		Short: "Список шаблонов проекта",
		Long: `Выводит список всех шаблонов указанного проекта.

Показывает ID, название и признак шаблона по умолчанию.
Поддерживает вывод в JSON для автоматизации.`,
		Example: `  # Список шаблонов проекта
  gotr templates list 1

  # Сохранить список в файл
  gotr templates list 1 -o templates.json`,
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
					return fmt.Errorf("project_id is required in non-interactive mode: gotr templates list [project_id]")
				}
				if _, ok := interactive.PrompterFromContext(ctx).(*interactive.NonInteractivePrompter); ok {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr templates list [project_id]")
				}

				projectID, err = resolveProjectIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			resp, err := cli.GetTemplates(ctx, projectID)
			if err != nil {
				return fmt.Errorf("failed to list templates: %w", err)
			}

			return output.OutputResult(cmd, resp, "templates")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
