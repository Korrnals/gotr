package users

import (
	"fmt"

	"github.com/Korrnals/gotr/cmd/internal/output"
	"github.com/spf13/cobra"
)

// newGetByEmailCmd создаёт команду 'users get-by-email'
// Эндпоинт: GET /get_user_by_email
func newGetByEmailCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-by-email <email>",
		Short: "Получить пользователя по email",
		Long: `Получает информацию о пользователе по его email адресу.

Выводит полную информацию: ID, имя, email, статус активности, 
роль, ID роли, MFA статус и признак администратора.

Полезно для поиска пользователя, когда известен email, но не ID.`,
		Example: `  # Получить пользователя по email
  gotr users get-by-email user@example.com

  # Сохранить результат в файл
  gotr users get-by-email user@example.com -o user.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			email := args[0]
			if email == "" {
				return fmt.Errorf("email cannot be empty")
			}

			cli := getClient(cmd)
			resp, err := cli.GetUserByEmail(email)
			if err != nil {
				return fmt.Errorf("failed to get user by email: %w", err)
			}

			return output.Result(cmd, resp)
		},
	}

	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")

	return cmd
}
