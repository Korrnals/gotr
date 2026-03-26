package users

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду 'users get'
// Эндпоинт: GET /get_user/{user_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [user_id]",
		Short: "Получить информацию о пользователе по ID",
		Long: `Получает детальную информацию о пользователе по его идентификатору.

Выводит полную информацию: ID, имя, email, статус активности, 
роль, ID роли, MFA статус и признак администратора.`,
		Example: `  # Получить информацию о пользователе
  gotr users get 12345

  # Сохранить результат в файл
  gotr users get 12345 -o user.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var userID int64
			var err error
			if len(args) > 0 {
				userID, err = flags.ValidateRequiredID(args, 0, "user_id")
				if err != nil {
					return err
				}
			} else {
				if err := requireInteractiveUserArg(cmd.Context(), "gotr users get [user_id]"); err != nil {
					return err
				}
				userID, err = resolveUserIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.GetUser(ctx, userID)
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}

			_, err = output.Output(cmd, resp, "users", "json")
			return err
		},
	}

	_ = interactive.HasPrompterInContext
	output.AddFlag(cmd)

	return cmd
}
