package users

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду 'users get'
// Эндпоинт: GET /get_user/{user_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <user_id>",
		Short: "Получить информацию о пользователе по ID",
		Long: `Получает детальную информацию о пользователе по его идентификатору.

Выводит полную информацию: ID, имя, email, статус активности, 
роль, ID роли, MFA статус и признак администратора.`,
		Example: `  # Получить информацию о пользователе
  gotr users get 12345

  # Сохранить результат в файл
  gotr users get 12345 -o user.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || userID <= 0 {
				return fmt.Errorf("invalid user_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetUser(userID)
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}

			_, err = save.Output(cmd, resp, "users", "json")
			return err
		},
	}

	save.AddFlag(cmd)

	return cmd
}
