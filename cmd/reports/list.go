package reports

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd creates the 'reports list' command.
// Endpoint: GET /get_reports/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project_id]",
		Short: "Список шаблонов отчётов проекта",
		Long: `Выводит список всех шаблонов отчётов указанного проекта.

Показывает ID, название и описание всех доступных шаблонов отчётов.
Поддерживает вывод в JSON для автоматизации.`,
		Example: `  # Список шаблонов отчётов проекта
  gotr reports list 1

  # Сохранить список в файл
  gotr reports list 1 -o reports.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var projectID int64
			var err error
			if len(args) > 0 {
				projectID, err = flags.ValidateRequiredID(args, 0, "project_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr reports list [project_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("project_id is required in non-interactive mode: gotr reports list [project_id]")
				}

				projectID, err = resolveProjectIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			resp, err := cli.GetReports(ctx, projectID)
			if err != nil {
				return fmt.Errorf("failed to list reports: %w", err)
			}

			_, err = output.Output(cmd, resp, "reports", "json")
			return err
		},
	}

	output.AddFlag(cmd)

	return cmd
}
