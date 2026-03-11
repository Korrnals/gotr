package tests

import (
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newUpdateCmd создаёт команду 'tests update'
// Эндпоинт: POST /update_test/{test_id}
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <test_id>",
		Short: "Обновить тест",
		Long: `Обновляет тест (результат выполнения тест-кейса).

Можно изменить статус теста (passed, failed, blocked, etc.) и
назначить исполнителя.`,
		Example: `  # Обновить статус теста
  gotr tests update 12345 --status-id=1

  # Назначить исполнителя
  gotr tests update 12345 --assigned-to=5

  # Проверить перед обновлением
  gotr tests update 12345 --status-id=5 --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			testID, err := flags.ValidateRequiredID(args, 0, "test_id")
			if err != nil {
				return err
			}

			req := data.UpdateTestRequest{}

			if v, _ := cmd.Flags().GetInt64("status-id"); v > 0 {
				req.StatusID = v
			}
			if v, _ := cmd.Flags().GetInt64("assigned-to"); v > 0 {
				req.AssignedTo = v
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("tests update")
				dr.PrintSimple("Обновить тест", fmt.Sprintf("Test ID: %d", testID))
				return nil
			}

			cli := getClient(cmd)
			ctx := cmd.Context()
			resp, err := cli.UpdateTest(ctx, testID, &req)
			if err != nil {
				return fmt.Errorf("failed to update test: %w", err)
			}

			fmt.Printf("✅ Тест %d обновлён\n", testID)
			return printJSON(cmd, resp, time.Now())
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	output.AddFlag(cmd)
	cmd.Flags().Int64("status-id", 0, "ID статуса теста (1=passed, 5=failed, etc.)")
	cmd.Flags().Int64("assigned-to", 0, "ID пользователя для назначения")

	return cmd
}
