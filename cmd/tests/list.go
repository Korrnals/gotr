package tests

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду для списка тестов
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <run_id>",
		Short: "Получить список тестов прогона",
		Long: `Получить список всех тестов для указанного тестового прогона.

Для фильтрации по статусу используйте флаг --status-id.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()
			client := getClient(cmd)

			runID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || runID <= 0 {
				return fmt.Errorf("run_id должен быть положительным числом")
			}

			statusID, _ := cmd.Flags().GetInt64("status-id")

			filters := make(map[string]string)
			if statusID > 0 {
				filters["status_id"] = fmt.Sprintf("%d", statusID)
			}

			// Проверяем dry-run режим
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := dryrun.New("tests list")
				details := fmt.Sprintf("Run ID: %d", runID)
				if statusID > 0 {
					details += fmt.Sprintf(", Status ID: %d", statusID)
				}
				dr.PrintSimple("Получить список тестов", details)
				return nil
			}

			tests, err := client.GetTests(runID, filters)
			if err != nil {
				return err
			}

			return outputResult(cmd, tests, start)
		},
	}

	cmd.Flags().Int64("status-id", 0, "ID статуса для фильтрации")
	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	save.AddFlag(cmd)

	return cmd
}
