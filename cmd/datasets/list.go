package datasets

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd creates the 'datasets list' command.
// Endpoint: GET /get_datasets/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project_id]",
		Short: "Список датасетов проекта",
		Long: `Выводит список всех датасетов (наборов тестовых данных),
доступных в указанном проекте.

Каждый датасет содержит название и таблицу с параметрами для
параметризованного тестирования.`,
		Example: `  # Получить список датасетов проекта
  gotr datasets list 1

  # Сохранить в файл
  gotr datasets list 5 -o datasets.json`,
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
					return fmt.Errorf("project_id required: gotr datasets list [project_id]")
				}
				projectID, err = resolveProjectIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			resp, err := cli.GetDatasets(ctx, projectID)
			if err != nil {
				return fmt.Errorf("failed to get datasets list: %w", err)
			}

			return output.OutputResult(cmd, resp, "datasets")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
