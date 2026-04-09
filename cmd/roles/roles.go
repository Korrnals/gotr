// Package roles implements CLI commands for managing TestRail user roles.
package roles

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining an API client.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register adds all role-related subcommands to the root command.
func Register(root *cobra.Command, getClient GetClientFunc) {
	rolesCmd := &cobra.Command{
		Use:   "roles",
		Short: "Manage user roles",
		Long: `Manage TestRail user roles.

Roles define the set of access rights for users in the TestRail system.
Each role has a unique ID and name (e.g., Administrator, Tester, Guest).

Roles are used to control access to projects, test cases,
test runs, and other system entities.

Available operations:
  • list — list all roles in the system
  • get  — information about a specific role by ID`,
	}

	// Register subcommands
	rolesCmd.AddCommand(newListCmd(getClient))
	rolesCmd.AddCommand(newGetCmd(getClient))

	root.AddCommand(rolesCmd)
}
