package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/internal/models/data"
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
			testID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || testID <= 0 {
				return fmt.Errorf("некорректный test_id: %s", args[0])
			}

			req := data.UpdateTestRequest{}

			if v, _ := cmd.Flags().GetInt64("status-id"); v > 0 {
				req.StatusID = v
			}
			if v, _ := cmd.Flags().GetInt64("assigned-to"); v > 0 {
				req.AssignedTo = v
			}

			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("tests update")
				dr.PrintSimple("Обновить тест", fmt.Sprintf("Test ID: %d", testID))
				return nil
			}

			cli := getClient(cmd)
			resp, err := cli.UpdateTest(testID, &req)
			if err != nil {
				return fmt.Errorf("не удалось обновить тест: %w", err)
			}

			fmt.Printf("✅ Тест %d обновлён\n", testID)
			return printJSON(cmd, resp, time.Now())
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (JSON)")
	cmd.Flags().Int64("status-id", 0, "ID статуса теста (1=passed, 5=failed, etc.)")
	cmd.Flags().Int64("assigned-to", 0, "ID пользователя для назначения")

	return cmd
}

func _unused_outputResult(cmd *cobra.Command, data interface{}) error {
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
