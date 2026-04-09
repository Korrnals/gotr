// Package users implements CLI commands for managing TestRail users.
package users

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is a function type that returns a client instance.
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register registers all user-related subcommands on the given root command.
func Register(root *cobra.Command, getClient GetClientFunc) {
	usersCmd := &cobra.Command{
		Use:   "users",
		Short: "Manage TestRail users",
		Long: `Manage TestRail users and references.

Available operations:
  • list          — list all users
  • get           — get a user by ID
  • get-by-email  — get a user by email
  • add           — create a new user
  • update        — update a user`,
	}

	// Register subcommands
	usersCmd.AddCommand(newListCmd(getClient))
	usersCmd.AddCommand(newGetCmd(getClient))
	usersCmd.AddCommand(newGetByEmailCmd(getClient))
	usersCmd.AddCommand(newAddCmd(getClient))
	usersCmd.AddCommand(newUpdateCmd(getClient))

	root.AddCommand(usersCmd)
}
