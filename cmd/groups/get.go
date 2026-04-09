package groups

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd creates the 'groups get' command.
// Endpoint: GET /get_group/{group_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [group_id]",
		Short: "Get a group by ID",
		Long: `Get detailed information about a user group by its ID.

Includes the group name and a full list of users
belonging to the group with their roles and contact information.`,
		Example: `  # Get group information
  gotr groups get 1

  # Save to file
  gotr groups get 5 -o group.json`,
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
					return fmt.Errorf("group_id required: gotr groups get [group_id]")
				}
				groupID, err = resolveGroupIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			resp, err := cli.GetGroup(ctx, groupID)
			if err != nil {
				return fmt.Errorf("failed to get group: %w", err)
			}

			return output.OutputResult(cmd, resp, "groups")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
