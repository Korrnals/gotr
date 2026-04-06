package roles

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd creates the 'roles get' command.
// Endpoint: GET /get_role/{role_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [role_id]",
		Short: "Получить роль по ID",
		Long: `Получает информацию о роли пользователя по её ID.

Возвращает ID и название роли, которая используется для управления
правами доступа в системе TestRail.`,
		Example: `  # Получить информацию о роли
  gotr roles get 1

  # Сохранить в файл
  gotr roles get 3 -o role.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var roleID int64
			var err error
			if len(args) > 0 {
				roleID, err = flags.ValidateRequiredID(args, 0, "role_id")
				if err != nil {
					return err
				}
			} else {
				if err := requireInteractiveRoleArg(cmd.Context(), "gotr roles get [role_id]"); err != nil {
					return err
				}
				roleID, err = resolveRoleIDInteractive(cmd.Context(), getClient(cmd))
				if err != nil {
					return err
				}
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.GetRole(ctx, roleID)
			if err != nil {
				return fmt.Errorf("failed to get role: %w", err)
			}

			return output.OutputResult(cmd, resp, "roles")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
