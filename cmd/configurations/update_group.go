package configurations

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

// newUpdateGroupCmd creates the 'configurations update-group' command.
// Endpoint: POST /update_config_group/{group_id}
func newUpdateGroupCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group [group_id]",
		Short: "Update a configuration group",
		Long:  `Updates the name of an existing configuration group.`,
		Example: `  # Change group name
  gotr configurations update-group 5 --name="New Name"

  # Preview before updating
  gotr configurations update-group 5 --name="New Name" --dry-run`,
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
					return fmt.Errorf("group_id is required in non-interactive mode: gotr configurations update-group [group_id]")
				}
				if interactive.IsNonInteractive(ctx) {
					return fmt.Errorf("group_id is required in non-interactive mode: gotr configurations update-group [group_id]")
				}

				groupID, err = resolveGroupIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("configurations update-group")
				dr.PrintSimple("Update group", fmt.Sprintf("Group ID: %d, New Name: %s", groupID, name))
				return nil
			}

			req := data.UpdateConfigGroupRequest{Name: name}
			quiet, _ := cmd.Flags().GetBool("quiet")
			resp, err := ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Updating config group",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (*data.ConfigGroup, error) {
				return cli.UpdateConfigGroup(ctx, groupID, &req)
			})
			if err != nil {
				return fmt.Errorf("failed to update group: %w", err)
			}

			ui.Successf(os.Stdout, "Group %d updated", groupID)
			return output.OutputResult(cmd, resp, "configurations")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview what would be done without applying changes")
	output.AddFlag(cmd)
	cmd.Flags().String("name", "", "New group name (required)")

	return cmd
}
