package configurations

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'configurations list'
// Эндпоинт: GET /get_configs/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project_id>",
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
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || projectID <= 0 {
				return fmt.Errorf("некорректный project_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetConfigs(projectID)
			if err != nil {
				return fmt.Errorf("не удалось получить конфигурации: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}

	output.AddFlag(cmd)

	return cmd
}

// outputResult выводит результат в JSON или сохраняет в файл
func outputResult(cmd *cobra.Command, data interface{}) error {
	_, err := output.Output(cmd, data, "configurations", "json")
	return err
}
