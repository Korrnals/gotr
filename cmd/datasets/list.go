package datasets

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
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
			projectID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || projectID <= 0 {
				return fmt.Errorf("некорректный project_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetDatasets(projectID)
			if err != nil {
				return fmt.Errorf("не удалось получить список датасетов: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}

	save.AddFlag(cmd)

	return cmd
}

// outputResult выводит результат в JSON или сохраняет в файл
func outputResult(cmd *cobra.Command, data interface{}) error {
	_, err := save.Output(cmd, data, "datasets", "json")
	return err
}
