package plans

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newUpdateCmd creates 'plans update' command
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <plan_id>",
		Short: "Update a test plan",
		Long:  `Update an existing test plan.`,
		Example: `  gotr plans update 12345 --name="Updated Plan Name"
  gotr plans update 12345 --description="New description"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || planID <= 0 {
				return fmt.Errorf("invalid plan_id: %s", args[0])
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
				dr := dryrun.New("plans update")
				dr.PrintSimple("Update Plan", fmt.Sprintf("Plan ID: %d", planID))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.UpdatePlan(planID, &req)
			if err != nil {
				return fmt.Errorf("failed to update plan: %w", err)
			}

			fmt.Printf("✅ Plan %d updated\n", planID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done")
	cmd.Flags().StringP("output", "o", "", "Save response to file")
	cmd.Flags().String("name", "", "New plan name")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().Int64("milestone-id", 0, "Milestone ID")

	return cmd
}
