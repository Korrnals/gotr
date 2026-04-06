package plans

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newDeleteCmd creates the 'plans delete' command.
// Endpoint: POST /delete_plan/{plan_id}
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [plan_id]",
		Short: "Удалить тест-план",
		Long:  `Удаляет тест-план по его ID.`,
		Example: `  # Удалить план
  gotr plans delete 12345

  # Проверить перед удалением
  gotr plans delete 12345 --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var planID int64
			if len(args) > 0 {
				var err error
				planID, err = flags.ValidateRequiredID(args, 0, "plan_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("plan_id is required in non-interactive mode: gotr plans delete [plan_id]")
				}
				var err error
				planID, err = resolvePlanIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("plans delete")
				dr.PrintSimple("Delete Plan", fmt.Sprintf("Plan ID: %d", planID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			if err := cli.DeletePlan(ctx, planID); err != nil {
				return fmt.Errorf("failed to delete plan: %w", err)
			}

			ui.Successf(os.Stdout, "Plan %d deleted", planID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет удалено без реального удаления")

	return cmd
}
