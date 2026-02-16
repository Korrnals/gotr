package milestones

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'milestones list'
// Эндпоинт: GET /get_milestones/{project_id}
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project_id>",
		Short: "Список майлстонов проекта",
		Long: `Выводит список всех майлстонов указанного проекта.

Показывает ID, название, статус завершения и дедлайны всех майлстонов.
Поддерживает вывод в JSON для автоматизации.`,
		Example: `  # Список майлстонов проекта
  gotr milestones list 1

  # Сохранить список в файл
  gotr milestones list 1 -o milestones.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || projectID <= 0 {
				return fmt.Errorf("invalid project_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetMilestones(projectID)
			if err != nil {
				return fmt.Errorf("failed to list milestones: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}

	save.AddFlag(cmd)

	return cmd
}
