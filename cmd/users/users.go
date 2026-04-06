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
		Short: "Управление пользователями TestRail",
		Long: `Управление пользователями и справочниками TestRail.

Доступные операции:
  • list          — список всех пользователей
  • get           — получить пользователя по ID
  • get-by-email  — получить пользователя по email
  • add           — создать нового пользователя
  • update        — обновить пользователя`,
	}

	// Register subcommands
	usersCmd.AddCommand(newListCmd(getClient))
	usersCmd.AddCommand(newGetCmd(getClient))
	usersCmd.AddCommand(newGetByEmailCmd(getClient))
	usersCmd.AddCommand(newAddCmd(getClient))
	usersCmd.AddCommand(newUpdateCmd(getClient))

	root.AddCommand(usersCmd)
}
