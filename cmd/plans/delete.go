package plans

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/spf13/cobra"
)

// newDeleteCmd creates 'plans delete' command
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <plan_id>",
		Short: "Delete a test plan",
		Long:  `Delete a test plan by ID.`,
		Example: `  gotr plans delete 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || planID <= 0 {
				return fmt.Errorf("invalid plan_id: %s", args[0])
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("plans delete")
				dr.PrintSimple("Delete Plan", fmt.Sprintf("Plan ID: %d", planID))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.DeletePlan(planID); err != nil {
				return fmt.Errorf("failed to delete plan: %w", err)
			}

			fmt.Printf("✅ Plan %d deleted\n", planID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done")

	return cmd
}
