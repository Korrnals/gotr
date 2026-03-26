package reports

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
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
			templateID, err := flags.ValidateRequiredID(args, 0, "template_id")
			if err != nil {
				return err
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("reports run")
				dr.PrintOperation(
					fmt.Sprintf("Run report template %d", templateID),
					"GET",
					fmt.Sprintf("/index.php?/api/v2/run_report/%d", templateID),
					nil,
				)
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Running report...",
				Writer: os.Stderr,
			}, func(ctx context.Context) (any, error) {
				return cli.RunReport(ctx, templateID)
			})
			if err != nil {
				return fmt.Errorf("failed to run report: %w", err)
			}

			_, err = output.Output(cmd, resp, "reports", "json")
			return err
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без запуска генерации")
	output.AddFlag(cmd)

	return cmd
}
