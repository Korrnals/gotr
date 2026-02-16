// Package groups реализует CLI команды для работы с группами пользователей TestRail
package groups

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) client.ClientInterface

// Register регистрирует все команды для работы с группами
func Register(root *cobra.Command, getClient GetClientFunc) {
	groupsCmd := &cobra.Command{
		Use:   "groups",
		Short: "Управление группами пользователей",
		Long: `Управление группами (groups) пользователей TestRail.

Группы используются для организации пользователей по командам, отделам
или проектам для удобного управления правами доступа.

Каждая группа имеет уникальный ID, название и список пользователей,
входящих в эту группу. Группы используются при настройке прав доступа
к проектам и назначении ролей.

Доступные операции:
  • list — список групп проекта
  • get  — информация о конкретной группе по ID`,
	}

	// Добавление подкоманд
	groupsCmd.AddCommand(newListCmd(getClient))
	groupsCmd.AddCommand(newGetCmd(getClient))
	groupsCmd.AddCommand(newAddCmd(getClient))
	groupsCmd.AddCommand(newUpdateCmd(getClient))
	groupsCmd.AddCommand(newDeleteCmd(getClient))

	root.AddCommand(groupsCmd)
}
