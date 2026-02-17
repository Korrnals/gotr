// Package roles реализует CLI команды для работы с ролями пользователей TestRail
package roles

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register регистрирует все команды для работы с ролями
func Register(root *cobra.Command, getClient GetClientFunc) {
	rolesCmd := &cobra.Command{
		Use:   "roles",
		Short: "Управление ролями пользователей",
		Long: `Управление ролями (roles) пользователей TestRail.

Роли определяют набор прав доступа пользователей в системе TestRail.
Каждая роль имеет уникальный ID и название (например, Administrator, Tester, Guest).

Роли используются для контроля доступа к проектам, тест-кейсам,
тест-ранам и другим сущностям системы.

Доступные операции:
  • list — список всех ролей в системе
  • get  — информация о конкретной роли по ID`,
	}

	// Добавление подкоманд
	rolesCmd.AddCommand(newListCmd(getClient))
	rolesCmd.AddCommand(newGetCmd(getClient))

	root.AddCommand(rolesCmd)
}
