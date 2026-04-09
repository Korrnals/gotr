package milestones

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd creates the 'milestones get' command.
// Endpoint: GET /get_milestone/{milestone_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [milestone_id]",
		Short: "Get milestone information by ID",
		Long: `Gets detailed information about a milestone by its identifier.

Displays full information: name, description, dates, completion status,
number of associated test runs, etc.`,
		Example: `  # Get milestone information
  gotr milestones get 12345

  # Save result to file
  gotr milestones get 12345 -o milestone.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var milestoneID int64
			if len(args) > 0 {
				var err error
				milestoneID, err = flags.ValidateRequiredID(args, 0, "milestone_id")
				if err != nil {
					return err
				}
			} else {
				ctx := cmd.Context()
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("milestone_id is required in non-interactive mode: gotr milestones get [milestone_id]")
				}
				cli := getClient(cmd)
				var err error
				milestoneID, err = resolveMilestoneIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.GetMilestone(ctx, milestoneID)
			if err != nil {
				return fmt.Errorf("failed to get milestone: %w", err)
			}

			return output.OutputResult(cmd, resp, "milestones")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
