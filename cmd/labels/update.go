package labels

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newUpdateTestCmd создаёт команду 'labels update test'
// Эндпоинт: POST /update_test/{test_id}
func newUpdateTestCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test <test_id>",
		Short: "Обновить метки одного теста",
		Long:  `Обновляет метки для конкретного теста по его ID.`,
		Example: `  # Добавить метки smoke и critical
  gotr labels update test 12345 --labels="smoke,critical"

  # Проверить без изменений
  gotr labels update test 99999 --labels="regression" --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			testID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || testID <= 0 {
				return fmt.Errorf("invalid test_id: %s", args[0])
			}

			labelsFlag, _ := cmd.Flags().GetString("labels")
			if labelsFlag == "" {
				return fmt.Errorf("--labels flag is required")
			}
			labels := parseLabels(labelsFlag)

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("labels update test")
				dr.PrintSimple("Update Test Labels", fmt.Sprintf("Test ID: %d, Labels: %v", testID, labels))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.UpdateTestLabels(testID, labels); err != nil {
				return fmt.Errorf("failed to update labels: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✅ Labels updated for test %d: %v\n", testID, labels)
			return nil
		},
	}

	cmd.Flags().String("labels", "", "Список меток через запятую (обязательно)")
	cmd.MarkFlagRequired("labels")

	return cmd
}

// newUpdateTestsCmd создаёт команду 'labels update tests'
// Эндпоинт: POST /update_tests/{run_id}
func newUpdateTestsCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tests",
		Short: "Обновить метки нескольких тестов в прогоне",
		Long:  `Обновляет метки для нескольких тестов в рамках одного тестового прогона.`,
		Example: `  # Обновить метки для тестов 1,2,3 в прогоне 100
  gotr labels update tests --run-id=100 --test-ids=1,2,3 --labels="smoke,critical"

  # Проверить без изменений
  gotr labels update tests --run-id=200 --test-ids=10,20 --labels="regression" --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runID, _ := cmd.Flags().GetInt64("run-id")
			if runID <= 0 {
				return fmt.Errorf("--run-id is required and must be positive")
			}

			testIDsFlag, _ := cmd.Flags().GetString("test-ids")
			if testIDsFlag == "" {
				return fmt.Errorf("--test-ids is required")
			}
			testIDs := parseIntList(testIDsFlag)
			if len(testIDs) == 0 {
				return fmt.Errorf("invalid test-ids: %s", testIDsFlag)
			}

			labelsFlag, _ := cmd.Flags().GetString("labels")
			if labelsFlag == "" {
				return fmt.Errorf("--labels flag is required")
			}
			labels := parseLabels(labelsFlag)

			// Check dry-run
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("labels update tests")
				dr.PrintSimple("Update Tests Labels", fmt.Sprintf("Run ID: %d, Test IDs: %v, Labels: %v", runID, testIDs, labels))
				return nil
			}

			cli := getClient(cmd)
			if err := cli.UpdateTestsLabels(runID, testIDs, labels); err != nil {
				return fmt.Errorf("failed to update labels: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✅ Labels updated for %d tests in run %d: %v\n", len(testIDs), runID, labels)
			return nil
		},
	}

	cmd.Flags().Int64("run-id", 0, "ID тестового прогона (обязательно)")
	cmd.Flags().String("test-ids", "", "Список ID тестов через запятую (обязательно)")
	cmd.Flags().String("labels", "", "Список меток через запятую (обязательно)")

	cmd.MarkFlagRequired("run-id")
	cmd.MarkFlagRequired("test-ids")
	cmd.MarkFlagRequired("labels")

	return cmd
}

// parseLabels разбирает метки, разделённые запятыми
func parseLabels(s string) []string {
	var labels []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			labels = append(labels, part)
		}
	}
	return labels
}

// parseIntList разбирает список чисел, разделённых запятыми
func parseIntList(s string) []int64 {
	var ids []int64
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}
