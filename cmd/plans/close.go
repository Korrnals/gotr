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

// newCloseCmd создаёт команду 'plans close'
// Эндпоинт: POST /close_plan/{plan_id}
func newCloseCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close [plan_id]",
		Short: "Закрыть тест-план",
		Long:  `Закрывает открытый тест-план (отмечает как завершённый).`,
		Example: `  # Закрыть план
  gotr plans close 12345

  # Проверить перед закрытием
  gotr plans close 12345 --dry-run`,
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
					return fmt.Errorf("plan_id is required in non-interactive mode: gotr plans close [plan_id]")
				}
				var err error
				planID, err = resolvePlanIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("plans close")
				dr.PrintSimple("Close Plan", fmt.Sprintf("Plan ID: %d", planID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.ClosePlan(ctx, planID)
			if err != nil {
				return fmt.Errorf("failed to close plan: %w", err)
			}

			ui.Successf(os.Stdout, "Plan %d closed", planID)
			return output.OutputResult(cmd, resp, "plans")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без реального закрытия")
	output.AddFlag(cmd)

	return cmd
}
