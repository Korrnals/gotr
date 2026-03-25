package cases

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newDeleteCmd создаёт команду 'cases delete'
// Эндпоинт: POST /delete_case/{case_id}
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [case_id]",
		Short: "Удалить тест-кейс",
		Long:  `Удаляет тест-кейс по его ID.`,
		Example: `  # Удалить тест-кейс
  gotr cases delete 12345

  # Проверить перед удалением
  gotr cases delete 12345 --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var caseID int64
			var err error
			if len(args) > 0 {
				caseID, err = flags.ValidateRequiredID(args, 0, "case_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("case_id required: gotr cases delete [case_id]")
				}

				caseID, err = resolveCaseIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("cases delete")
				dr.PrintSimple("Delete Case", fmt.Sprintf("Case ID: %d", caseID))
				return nil
			}

			if err := cli.DeleteCase(ctx, caseID); err != nil {
				return fmt.Errorf("failed to delete case: %w", err)
			}

			ui.Successf(os.Stdout, "Case %d deleted", caseID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
