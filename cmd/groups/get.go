package groups

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
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
			groupID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || groupID <= 0 {
				return fmt.Errorf("некорректный group_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetGroup(groupID)
			if err != nil {
				return fmt.Errorf("не удалось получить группу: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}

	save.AddFlag(cmd)

	return cmd
}
