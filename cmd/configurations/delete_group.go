package configurations

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newDeleteGroupCmd creates the 'configurations delete-group' command.
// Endpoint: POST /delete_config_group/{group_id}
func newDeleteGroupCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-group [group_id]",
		Short: "Delete a configuration group",
		Long: `Deletes a configuration group and all its configurations.

⚠️ Warning: deletion cannot be undone! All configurations in the group
will also be deleted. Make sure the group is not used
in active test plans.`,
		Example: `  # Delete a group
  gotr configurations delete-group 5

  # Preview before deleting
  gotr configurations delete-group 5 --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var groupID int64
			var err error
			if len(args) > 0 {
				groupID, err = flags.ValidateRequiredID(args, 0, "group_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("group_id is required in non-interactive mode: gotr configurations delete-group [group_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("group_id is required in non-interactive mode: gotr configurations delete-group [group_id]")
				}

				groupID, err = resolveGroupIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("configurations delete-group")
				dr.PrintSimple("Delete group", fmt.Sprintf("Group ID: %d", groupID))
				return nil
			}

			if err := cli.DeleteConfigGroup(ctx, groupID); err != nil {
				return fmt.Errorf("failed to delete group: %w", err)
			}

			ui.Successf(os.Stdout, "Group %d deleted", groupID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview what would be deleted without actually deleting")

	return cmd
}
