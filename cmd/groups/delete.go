package groups

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// newDeleteCmd creates the 'groups delete' command.
func newDeleteCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [group_id]",
		Short: "Delete a group",
		Long:  `Delete a user group.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := getClient(cmd)
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
					return fmt.Errorf("group_id required: gotr groups delete [group_id]")
				}
				groupID, err = resolveGroupIDInteractive(ctx, client)
				if err != nil {
					return err
				}
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				ui.Infof(os.Stdout, "[DRY-RUN] Will delete group %d", groupID)
				return nil
			}

			quiet, _ := cmd.Flags().GetBool("quiet")
			_, err = ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Deleting group",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (struct{}, error) {
				return struct{}{}, client.DeleteGroup(ctx, groupID)
			})
			if err != nil {
				return err
			}

			if !quiet {
				color.New(color.FgGreen).Fprintf(cmd.OutOrStdout(), "✓ Group %d deleted\n", groupID)
			}

			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be done without executing")

	return cmd
}
