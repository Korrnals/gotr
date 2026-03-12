package groups

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду 'groups get'
// Эндпоинт: GET /get_group/{group_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <group_id>",
		Short: "Получить группу по ID",
		Long: `Получает детальную информацию о группе пользователей по её ID.

Включает название группы и полный список пользователей,
входящих в эту группу с их ролями и контактной информацией.`,
		Example: `  # Получить информацию о группе
  gotr groups get 1

  # Сохранить в файл
  gotr groups get 5 -o group.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupID, err := flags.ValidateRequiredID(args, 0, "group_id")
			if err != nil {
				return err
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.GetGroup(ctx, groupID)
			if err != nil {
				return fmt.Errorf("failed to get group: %w", err)
			}

			return output.OutputResult(cmd, resp, "groups")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
