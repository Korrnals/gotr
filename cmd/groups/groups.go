// Package groups implements CLI commands for managing TestRail user groups.
package groups

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is a function type for obtaining the API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all group management subcommands.
func Register(root *cobra.Command, getClient GetClientFunc) {
	groupsCmd := &cobra.Command{
		Use:   "groups",
		Short: "Manage user groups",
		Long: `Manage TestRail user groups.

Groups are used to organize users by teams, departments,
or projects for convenient access rights management.

Each group has a unique ID, a name, and a list of users
belonging to the group. Groups are used when configuring
project access rights and assigning roles.

Available operations:
  • list — list project groups
  • get  — get information about a specific group by ID`,
	}

	// Register subcommands
	groupsCmd.AddCommand(newListCmd(getClient))
	groupsCmd.AddCommand(newGetCmd(getClient))
	groupsCmd.AddCommand(newAddCmd(getClient))
	groupsCmd.AddCommand(newUpdateCmd(getClient))
	groupsCmd.AddCommand(newDeleteCmd(getClient))

	root.AddCommand(groupsCmd)
}
