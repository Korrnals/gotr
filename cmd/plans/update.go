package plans

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newUpdateCmd создаёт команду 'plans update'
// Эндпоинт: POST /update_plan/{plan_id}
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [plan_id]",
		Short: "Обновить тест-план",
		Long:  `Обновляет существующий тест-план.`,
		Example: `  # Изменить название плана
  gotr plans update 12345 --name="Новое название плана"

  # Изменить описание
  gotr plans update 12345 --description="Новое описание"`,
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
					return fmt.Errorf("plan_id is required in non-interactive mode: gotr plans update [plan_id]")
				}
				var err error
				planID, err = resolvePlanIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
			}

			req := data.UpdatePlanRequest{}

			if v, _ := cmd.Flags().GetString("name"); v != "" {
				req.Name = v
			}
			if v, _ := cmd.Flags().GetString("description"); v != "" {
				req.Description = v
			}
			if v, _ := cmd.Flags().GetInt64("milestone-id"); v > 0 {
				req.MilestoneID = v
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("plans update")
				dr.PrintSimple("Update Plan", fmt.Sprintf("Plan ID: %d", planID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.UpdatePlan(ctx, planID, &req)
			if err != nil {
				return fmt.Errorf("failed to update plan: %w", err)
			}

			ui.Successf(os.Stdout, "Plan %d updated", planID)
			return output.OutputResult(cmd, resp, "plans")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "Новое название плана")
	cmd.Flags().String("description", "", "Новое описание")
	cmd.Flags().Int64("milestone-id", 0, "ID майлстона")

	return cmd
}
