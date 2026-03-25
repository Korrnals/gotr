package cases

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'cases list'
// Эндпоинт: GET /get_cases/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project_id]",
		Short: "Список тест-кейсов",
		Long:  `Выводит список тест-кейсов проекта с возможностью фильтрации.`,
		Example: `  # Список всех кейсов проекта
  gotr cases list 1

  # Фильтрация по сьюте и секции
  gotr cases list 1 --suite-id=100 --section-id=50`,
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
					return fmt.Errorf("project_id required: gotr cases list [project_id]")
				}

				projectID, err = interactive.SelectProject(ctx, interactive.PrompterFromContext(ctx), cli, "")
				if err != nil {
					return err
				}
			}

			suiteID, _ := cmd.Flags().GetInt64("suite-id")
			sectionID, _ := cmd.Flags().GetInt64("section-id")

			resp, err := cli.GetCases(ctx, projectID, suiteID, sectionID)
			if err != nil {
				return fmt.Errorf("failed to list cases: %w", err)
			}

			return output.OutputResult(cmd, resp, "cases")
		},
	}

	cmd.Flags().Int64("suite-id", 0, "Фильтр по ID сьюты")
	cmd.Flags().Int64("section-id", 0, "Фильтр по ID секции")

	return cmd
}
