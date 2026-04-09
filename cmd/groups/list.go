package groups

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd creates the 'groups list' command.
// Endpoint: GET /get_groups/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [project_id]",
		Short: "List project groups",
		Long: `List user groups available in the specified project.

Each group contains an ID, name, and information about users
belonging to the group. Used for viewing team structure
and managing access rights within a project.`,
		Example: `  # List project groups
  gotr groups list 1

  # Save to file
  gotr groups list 5 -o groups.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var projectID int64
			var err error
			if len(args) > 0 {
				projectID, err = flags.ValidateRequiredID(args, 0, "project_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("project_id required: gotr groups list [project_id]")
				}
				projectID, err = resolveProjectIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
			}

			resp, err := cli.GetGroups(ctx, projectID)
			if err != nil {
				return fmt.Errorf("failed to get groups list: %w", err)
			}

			return output.OutputResult(cmd, resp, "groups")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
