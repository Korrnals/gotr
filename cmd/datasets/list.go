package datasets

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'datasets list'
// Эндпоинт: GET /get_datasets/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project_id>",
		Short: "Список датасетов проекта",
		Long: `Выводит список всех датасетов (наборов тестовых данных),
доступных в указанном проекте.

Каждый датасет содержит название и таблицу с параметрами для
параметризованного тестирования.`,
		Example: `  # Получить список датасетов проекта
  gotr datasets list 1

  # Сохранить в файл
  gotr datasets list 5 -o datasets.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := flags.ValidateRequiredID(args, 0, "project_id")
			if err != nil {
				return err
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
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
