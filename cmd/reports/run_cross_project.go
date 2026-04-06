package reports

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newRunCrossProjectCmd creates the 'reports run-cross-project' command.
// Endpoint: GET /run_cross_project_report/{template_id}
func newRunCrossProjectCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run-cross-project [template_id]",
		Short: "Запустить кросс-проектный отчёт",
		Long: `Запускает генерацию кросс-проектного отчёта по указанному шаблону.

Кросс-проектные отчёты охватывают несколько проектов TestRail.
Возвращает ID отчёта, URL для скачивания и статус генерации.`,
		Example: `  # Запустить кросс-проектный отчёт
  gotr reports run-cross-project 42

  # Сохранить результат в файл
  gotr reports run-cross-project 42 -o cross_project_report.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var templateID int64
			var err error
			if len(args) > 0 {
				templateID, err = flags.ValidateRequiredID(args, 0, "template_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("template_id is required in non-interactive mode: gotr reports run-cross-project [template_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("template_id is required in non-interactive mode: gotr reports run-cross-project [template_id]")
				}

				templateID, err = resolveCrossProjectReportTemplateIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("reports run-cross-project")
				dr.PrintOperation(
					fmt.Sprintf("Run cross-project report template %d", templateID),
					"GET",
					fmt.Sprintf("/index.php?/api/v2/run_cross_project_report/%d", templateID),
					nil,
				)
				return nil
			}

			resp, err := ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Running cross-project report...",
				Writer: os.Stderr,
			}, func(ctx context.Context) (any, error) {
				return cli.RunCrossProjectReport(ctx, templateID)
			})
			if err != nil {
				return fmt.Errorf("failed to run cross-project report: %w", err)
			}

			_, err = output.Output(cmd, resp, "reports", "json")
			return err
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без запуска генерации")
	output.AddFlag(cmd)

	return cmd
}
