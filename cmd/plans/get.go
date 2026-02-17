package plans

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду 'plans get'
// Эндпоинт: GET /get_plan/{plan_id}
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <plan_id>",
		Short: "Получить тест-план по ID",
		Long:  `Получает детальную информацию о тест-плане, включая записи (entries).`,
		Example: `  # Получить информацию о плане
  gotr plans get 12345

  # Сохранить в файл
  gotr plans get 12345 --save`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || planID <= 0 {
				return fmt.Errorf("invalid plan_id: %s", args[0])
			}

			cli := getClient(cmd)
			resp, err := cli.GetPlan(planID)
			if err != nil {
				return fmt.Errorf("failed to get plan: %w", err)
			}

			return outputResult(cmd, resp)
		},
	}

	save.AddFlag(cmd)

	return cmd
}
