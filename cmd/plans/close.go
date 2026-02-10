package plans

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/spf13/cobra"
)

// newCloseCmd создаёт команду 'plans close'
// Эндпоинт: POST /close_plan/{plan_id}
func newCloseCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close <plan_id>",
		Short: "Закрыть тест-план",
		Long:  `Закрывает открытый тест-план (отмечает как завершённый).`,
		Example: `  # Закрыть план
  gotr plans close 12345

  # Проверить перед закрытием
  gotr plans close 12345 --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || planID <= 0 {
				return fmt.Errorf("invalid plan_id: %s", args[0])
			}

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("plans close")
				dr.PrintSimple("Close Plan", fmt.Sprintf("Plan ID: %d", planID))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.ClosePlan(planID)
			if err != nil {
				return fmt.Errorf("failed to close plan: %w", err)
			}

			fmt.Printf("✅ Plan %d closed\n", planID)
			return outputResult(cmd, resp)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без реального закрытия")
	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")

	return cmd
}
