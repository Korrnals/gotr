package datasets

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду 'datasets get'
// Эндпоинт: GET /get_dataset/{dataset_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <dataset_id>",
		Short: "Получить датасет по ID",
		Long: `Получает детальную информацию о датасете по его ID.

Включает название, структуру таблицы (колонки) и все строки
с тестовыми данными для параметризованного тестирования.`,
		Example: `  # Получить информацию о датасете
  gotr datasets get 123

  # Сохранить в файл
  gotr datasets get 456 -o dataset.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			datasetID, err := flags.ValidateRequiredID(args, 0, "dataset_id")
			if err != nil {
				return err
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.GetDataset(ctx, datasetID)
			if err != nil {
				return fmt.Errorf("failed to get dataset: %w", err)
			}

			return output.OutputResult(cmd, resp, "datasets")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
