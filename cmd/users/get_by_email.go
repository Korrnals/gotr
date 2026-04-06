package users

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetByEmailCmd creates the 'users get-by-email' command.
// Endpoint: GET /get_user_by_email
func newGetByEmailCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-by-email [email]",
		Short: "Получить пользователя по email",
		Long: `Получает информацию о пользователе по его email адресу.

Выводит полную информацию: ID, имя, email, статус активности, 
роль, ID роли, MFA статус и признак администратора.

Полезно для поиска пользователя, когда известен email, но не ID.`,
		Example: `  # Получить пользователя по email
  gotr users get-by-email user@example.com

  # Сохранить результат в файл
  gotr users get-by-email user@example.com -o user.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var email string
			if len(args) > 0 {
				email = args[0]
			} else {
				if err := requireInteractiveUserArg(cmd.Context(), "gotr users get-by-email [email]"); err != nil {
					return err
				}
				var err error
				email, err = resolveEmailInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
			}
			if email == "" {
				return fmt.Errorf("email cannot be empty")
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.GetUserByEmail(ctx, email)
			if err != nil {
				return fmt.Errorf("failed to get user by email: %w", err)
			}

			_, err = output.Output(cmd, resp, "users", "json")
			return err
		},
	}

	output.AddFlag(cmd)

	return cmd
}
