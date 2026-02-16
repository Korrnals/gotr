package reports

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'reports list'
// Эндпоинт: GET /get_reports/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project_id>",
		Short: "Список шаблонов отчётов проекта",
		Long: `Выводит список всех шаблонов отчётов указанного проекта.

Показывает ID, название и описание всех доступных шаблонов отчётов.
Поддерживает вывод в JSON для автоматизации.`,
		Example: `  # Список шаблонов отчётов проекта
  gotr reports list 1

  # Сохранить список в файл
  gotr reports list 1 -o reports.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || projectID <= 0 {
				return fmt.Errorf("invalid project_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetReports(projectID)
			if err != nil {
				return fmt.Errorf("failed to list reports: %w", err)
			}

			_, err = save.Output(cmd, resp, "reports", "json")
			return err
		},
	}

	save.AddFlag(cmd)

	return cmd
}
