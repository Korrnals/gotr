package templates

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'templates list'
// Эндпоинт: GET /get_templates/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project_id>",
		Short: "Список шаблонов проекта",
		Long: `Выводит список всех шаблонов указанного проекта.

Показывает ID, название и признак шаблона по умолчанию.
Поддерживает вывод в JSON для автоматизации.`,
		Example: `  # Список шаблонов проекта
  gotr templates list 1

  # Сохранить список в файл
  gotr templates list 1 -o templates.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || projectID <= 0 {
				return fmt.Errorf("invalid project_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetTemplates(projectID)
			if err != nil {
				return fmt.Errorf("failed to list templates: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}

	output.AddFlag(cmd)

	return cmd
}

// outputResult выводит результат в JSON или сохраняет в файл
func outputResult(cmd *cobra.Command, data interface{}) error {
	_, err := output.Output(cmd, data, "templates", "json")
	return err
}
