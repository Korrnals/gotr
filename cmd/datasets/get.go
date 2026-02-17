package datasets

import (
	"fmt"
	"strconv"

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
			datasetID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || datasetID <= 0 {
				return fmt.Errorf("некорректный dataset_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetDataset(datasetID)
			if err != nil {
				return fmt.Errorf("не удалось получить датасет: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}

	output.AddFlag(cmd)

	return cmd
}
