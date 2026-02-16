package reports

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/internal/output"
	"github.com/spf13/cobra"
)

// newRunCrossProjectCmd создаёт команду 'reports run-cross-project'
// Эндпоинт: GET /run_cross_project_report/{template_id}
func newRunCrossProjectCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run-cross-project <template_id>",
		Short: "Запустить кросс-проектный отчёт",
		Long: `Запускает генерацию кросс-проектного отчёта по указанному шаблону.

Кросс-проектные отчёты охватывают несколько проектов TestRail.
Возвращает ID отчёта, URL для скачивания и статус генерации.`,
		Example: `  # Запустить кросс-проектный отчёт
  gotr reports run-cross-project 42

  # Сохранить результат в файл
  gotr reports run-cross-project 42 -o cross_project_report.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || templateID <= 0 {
				return fmt.Errorf("invalid template_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.RunCrossProjectReport(templateID)
			if err != nil {
				return fmt.Errorf("failed to run cross-project report: %w", err)
			}

			return output.Result(cmd, resp)
		},
	}

	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")

	return cmd
}
