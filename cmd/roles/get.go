package roles

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду 'roles get'
// Эндпоинт: GET /get_role/{role_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <role_id>",
		Short: "Получить роль по ID",
		Long: `Получает информацию о роли пользователя по её ID.

Возвращает ID и название роли, которая используется для управления
правами доступа в системе TestRail.`,
		Example: `  # Получить информацию о роли
  gotr roles get 1

  # Сохранить в файл
  gotr roles get 3 -o role.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			roleID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || roleID <= 0 {
				return fmt.Errorf("некорректный role_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetRole(roleID)
			if err != nil {
				return fmt.Errorf("не удалось получить роль: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}

	output.AddFlag(cmd)

	return cmd
}
