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

// newUpdateCmd creates the 'plans update' command.
// Endpoint: POST /update_plan/{plan_id}
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [plan_id]",
		Short: "Update a test plan",
		Long:  `Updates an existing test plan.`,
		Example: `  # Change plan name
  gotr plans update 12345 --name="New plan name"

  # Change description
  gotr plans update 12345 --description="New description"`,
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

	cmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "New plan name")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().Int64("milestone-id", 0, "Milestone ID")

	return cmd
}
