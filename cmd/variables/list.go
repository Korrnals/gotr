package variables

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'variables list'
// Эндпоинт: GET /get_variables/{dataset_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <dataset_id>",
		Short: "Список переменных датасета",
		Long: `Выводит список переменных, определённых в указанном датасете.

Переменные используются для параметризованного тестирования
и представляют собой колонки в таблице тестовых данных.`,
		Example: `  # Получить список переменных датасета
  gotr variables list 123

  # Сохранить в файл
  gotr variables list 456 -o vars.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			datasetID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || datasetID <= 0 {
				return fmt.Errorf("некорректный dataset_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetVariables(datasetID)
			if err != nil {
				return fmt.Errorf("не удалось получить список переменных: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}

	save.AddFlag(cmd)

	return cmd
}

// outputResult выводит результат в JSON или сохраняет в файл
func outputResult(cmd *cobra.Command, data interface{}) error {
	_, err := save.Output(cmd, data, "variables", "json")
	return err
}
