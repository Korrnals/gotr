package roles

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'roles list'
// Эндпоинт: GET /get_roles
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
			resp, err := cli.GetRoles()
			if err != nil {
				return fmt.Errorf("не удалось получить список ролей: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")

	return cmd
}

// outputResult выводит результат в JSON или сохраняет в файл
func outputResult(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if output != "" {
		return os.WriteFile(output, jsonBytes, 0644)
	}

	fmt.Println(string(jsonBytes))
	return nil
}
