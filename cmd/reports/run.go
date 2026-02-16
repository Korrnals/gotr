package reports

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/internal/output"
	"github.com/spf13/cobra"
)

// newRunCmd создаёт команду 'reports run'
// Эндпоинт: GET /run_report/{template_id}
func newRunCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <template_id>",
		Short: "Запустить генерацию отчёта по шаблону",
		Long: `Запускает генерацию отчёта по указанному шаблону.

Возвращает ID отчёта, URL для скачивания и статус генерации.
Для проверки статуса готовности отчёта выполните команду повторно.`,
		Example: `  # Запустить генерацию отчёта
  gotr reports run 42

  # Сохранить результат в файл
  gotr reports run 42 -o report_result.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || templateID <= 0 {
				return fmt.Errorf("invalid template_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.RunReport(templateID)
			if err != nil {
				return fmt.Errorf("failed to run report: %w", err)
			}

			return output.Result(cmd, resp)
		},
	}

	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")

	return cmd
}
