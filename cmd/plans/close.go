package plans

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/spf13/cobra"
)

// newCloseCmd creates 'plans close' command
func newCloseCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close <plan_id>",
		Short: "Close a test plan",
		Long:  `Close an open test plan (mark as completed).`,
		Example: `  gotr plans close 12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || planID <= 0 {
				return fmt.Errorf("invalid plan_id: %s", args[0])
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("plans close")
				dr.PrintSimple("Close Plan", fmt.Sprintf("Plan ID: %d", planID))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.ClosePlan(planID)
			if err != nil {
				return fmt.Errorf("failed to close plan: %w", err)
			}

			fmt.Printf("✅ Plan %d closed\n", planID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done")
	cmd.Flags().StringP("output", "o", "", "Save response to file")

	return cmd
}
