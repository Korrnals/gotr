package roles

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd creates the 'roles list' command.
// Endpoint: GET /get_roles
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List system roles",
		Long: `Displays a list of all user roles available in the TestRail system.

Each role contains an ID and a name. Roles are used to manage
user access rights to various system features.`,
		Example: `  # Get list of all roles
  gotr roles list

  # Save to a file
  gotr roles list -o roles.json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.GetRoles(ctx)
			if err != nil {
				return fmt.Errorf("failed to get roles list: %w", err)
			}

			return output.OutputResult(cmd, resp, "roles")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
