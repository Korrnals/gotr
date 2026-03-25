package tests

import (
	"fmt"
	"os"
	"time"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newUpdateCmd создаёт команду 'tests update'
// Эндпоинт: POST /update_test/{test_id}
func newUpdateCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [test_id]",
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
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			var testID int64
			var err error
			if len(args) > 0 {
				testID, err = flags.ValidateRequiredID(args, 0, "test_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("test_id is required in non-interactive mode: gotr tests update [test_id]")
				}
				if _, ok := interactive.PrompterFromContext(ctx).(*interactive.NonInteractivePrompter); ok {
					return fmt.Errorf("test_id is required in non-interactive mode: gotr tests update [test_id]")
				}
				testID, err = resolveTestIDInteractive(ctx, cli)
				if err != nil {
					return err
				}
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

			resp, err := cli.UpdateTest(ctx, testID, &req)
			if err != nil {
				return fmt.Errorf("failed to update test: %w", err)
			}

			ui.Successf(os.Stdout, "Test %d updated", testID)
			return printJSON(cmd, resp, time.Now())
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	output.AddFlag(cmd)
	cmd.Flags().Int64("status-id", 0, "ID статуса теста (1=passed, 5=failed, etc.)")
	cmd.Flags().Int64("assigned-to", 0, "ID пользователя для назначения")

	return cmd
}
