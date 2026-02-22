package milestones

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду 'milestones get'
// Эндпоинт: GET /get_milestone/{milestone_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <milestone_id>",
		Short: "Получить информацию о майлстоне по ID",
		Long: `Получает детальную информацию о майлстоне по его идентификатору.

Выводит полную информацию: название, описание, даты, статус завершения,
количество связанных тестовых прогонов и т.д.`,
		Example: `  # Получить информацию о майлстоне
  gotr milestones get 12345

  # Сохранить результат в файл
  gotr milestones get 12345 -o milestone.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			milestoneID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || milestoneID <= 0 {
				return fmt.Errorf("invalid milestone_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetMilestone(milestoneID)
			if err != nil {
				return fmt.Errorf("failed to get milestone: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}

	output.AddFlag(cmd)

	return cmd
}
