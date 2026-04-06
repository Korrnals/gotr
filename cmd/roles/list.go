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
		Short: "Список ролей системы",
		Long: `Выводит список всех ролей пользователей, доступных в системе TestRail.

Каждая роль содержит ID и название. Роли используются для управления
правами доступа пользователей к различным функциям системы.`,
		Example: `  # Получить список всех ролей
  gotr roles list

  # Сохранить в файл
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
