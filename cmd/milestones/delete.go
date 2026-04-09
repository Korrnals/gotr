package milestones

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newDeleteCmd creates the 'milestones delete' command.
// Endpoint: POST /delete_milestone/{milestone_id}
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [milestone_id]",
		Short: "Delete a milestone",
		Long: `Deletes a milestone by its identifier.

⚠️ Warning: deletion cannot be undone!
A deleted milestone cannot be restored; you will need to create a new one.
Use --dry-run to verify before deleting.`,
		Example: `  # Delete a milestone (with danger confirmation)
  gotr milestones delete 12345

  # Check what would be deleted (without actually deleting)
  gotr milestones delete 12345 --dry-run`,
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
					return fmt.Errorf("milestone_id is required in non-interactive mode: gotr milestones delete [milestone_id]")
				}
				cli := getClient(cmd)
				var err error
				milestoneID, err = resolveMilestoneIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("milestones delete")
				dr.PrintSimple("Delete Milestone", fmt.Sprintf("Milestone ID: %d", milestoneID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			if err := cli.DeleteMilestone(ctx, milestoneID); err != nil {
				return fmt.Errorf("failed to delete milestone: %w", err)
			}

			ui.Successf(os.Stdout, "Milestone %d deleted", milestoneID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be deleted without actually deleting")

	return cmd
}
