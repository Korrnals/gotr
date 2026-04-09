package groups

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newUpdateCmd creates the 'groups update' command.
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [group_id]",
		Short: "Update a group",
		Long:  `Update an existing user group.`,
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
					return fmt.Errorf("group_id required: gotr groups update [group_id] --name <name>")
				}
				groupID, err = resolveGroupIDInteractive(ctx, client)
				if err != nil {
					return err
				}
			}

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				ui.Infof(os.Stdout, "[DRY-RUN] Will update group %d, new name: '%s'", groupID, name)
				return nil
			}

			group, err := client.UpdateGroup(ctx, groupID, name, nil)
			if err != nil {
				return err
			}

			return output.OutputResult(cmd, group, "groups")
		},
	}

	cmd.Flags().StringP("name", "n", "", "New group name (required)")
	cmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	output.AddFlag(cmd)

	return cmd
}
