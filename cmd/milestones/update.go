package milestones

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newUpdateCmd creates the 'milestones update' command.
// Endpoint: POST /update_milestone/{milestone_id}
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [milestone_id]",
		Short: "Update an existing milestone",
		Long: `Updates data of an existing milestone.

You can change the name, description, deadline, and completion status.
All flags are optional — only specified fields will be changed.`,
		Example: `  # Change milestone name
  gotr milestones update 12345 --name="Release 1.1"

  # Change deadline
  gotr milestones update 12345 --due-on="2026-04-01"

  # Mark as completed
  gotr milestones update 12345 --is-completed=true

  # Change multiple fields
  gotr milestones update 12345 --name="New Name" --description="New description"`,
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
					return fmt.Errorf("milestone_id is required in non-interactive mode: gotr milestones update [milestone_id]")
				}
				cli := getClient(cmd)
				var err error
				milestoneID, err = resolveMilestoneIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			req := data.UpdateMilestoneRequest{}

			if v, _ := cmd.Flags().GetString("name"); v != "" {
				req.Name = v
			}
			if v, _ := cmd.Flags().GetString("description"); v != "" {
				req.Description = v
			}
			if v, _ := cmd.Flags().GetString("due-on"); v != "" {
				req.DueOn = v
			}
			if v, _ := cmd.Flags().GetBool("is-completed"); v {
				req.IsCompleted = true
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("milestones update")
				dr.PrintSimple("Update Milestone", fmt.Sprintf("Milestone ID: %d", milestoneID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			quiet, _ := cmd.Flags().GetBool("quiet")
			resp, err := ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Updating milestone",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (*data.Milestone, error) {
				return cli.UpdateMilestone(ctx, milestoneID, &req)
			})
			if err != nil {
				return fmt.Errorf("failed to update milestone: %w", err)
			}

			ui.Successf(os.Stdout, "Milestone %d updated", milestoneID)
			return output.OutputResult(cmd, resp, "milestones")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done without actually executing")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "New milestone name")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().String("due-on", "", "New deadline (YYYY-MM-DD)")
	cmd.Flags().Bool("is-completed", false, "Mark as completed")

	return cmd
}
