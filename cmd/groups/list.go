package groups

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'groups list'
// Эндпоинт: GET /get_groups/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project_id>",
		Short: "Список групп проекта",
		Long: `Выводит список групп пользователей, доступных в указанном проекте.

Каждая группа содержит ID, название и информацию о пользователях,
входящих в группу. Используется для просмотра структуры команд
и управления правами доступа в рамках проекта.`,
		Example: `  # Получить список групп проекта
  gotr groups list 1

  # Сохранить в файл
  gotr groups list 5 -o groups.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := flags.ValidateRequiredID(args, 0, "project_id")
			if err != nil {
				return err
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.GetGroups(ctx, projectID)
			if err != nil {
				return fmt.Errorf("failed to get groups list: %w", err)
			}

			return output.OutputResult(cmd, resp, "groups")
		},
	}

	output.AddFlag(cmd)

	return cmd
}
