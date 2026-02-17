// Package users реализует CLI команды для работы с пользователями TestRail
package users

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register регистрирует все команды для работы с пользователями
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

	// Добавление подкоманд
	usersCmd.AddCommand(newListCmd(getClient))
	usersCmd.AddCommand(newGetCmd(getClient))
	usersCmd.AddCommand(newGetByEmailCmd(getClient))
	usersCmd.AddCommand(newAddCmd(getClient))
	usersCmd.AddCommand(newUpdateCmd(getClient))

	root.AddCommand(usersCmd)
}
