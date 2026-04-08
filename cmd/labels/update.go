package labels

import (
	"fmt"
	"strings"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newUpdateTestCmd creates the 'labels update test' command.
// Endpoint: POST /update_test/{test_id}
func newUpdateTestCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test [test_id]",
		Short: "Обновить метки одного теста",
		Long:  `Обновляет метки для конкретного теста по его ID.`,
		Example: `  # Добавить метки smoke и critical
  gotr labels update test 12345 --labels="smoke,critical"

  # Проверить без изменений
  gotr labels update test 99999 --labels="regression" --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var testID int64
			var err error
			if len(args) > 0 {
				testID, err = flags.ValidateRequiredID(args, 0, "test_id")
				if err != nil {
					return err
				}
			} else {
				if !interactive.HasPrompterInContext(cmd.Context()) {
					return fmt.Errorf("test_id is required in non-interactive mode: gotr labels update test [test_id]")
				}
				if _, ok := interactive.PrompterFromContext(cmd.Context()).(*interactive.NonInteractivePrompter); ok {
					return fmt.Errorf("test_id is required in non-interactive mode: gotr labels update test [test_id]")
				}
				if testID, err = resolveTestIDInteractive(cmd.Context(), getClient(cmd)); err != nil {
					return err
				}
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
			ctx := cmd.Context()
			if err := cli.UpdateTestLabels(ctx, testID, labels); err != nil {
				return fmt.Errorf("failed to update labels: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✅ Labels updated for test %d: %v\n", testID, labels)
			return nil
		},
	}

	cmd.Flags().String("labels", "", "Список меток через запятую (обязательно)")
	_ = cmd.MarkFlagRequired("labels")

	return cmd
}

// newUpdateTestsCmd creates the 'labels update tests' command.
// Endpoint: POST /update_tests/{run_id}
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
			ctx := cmd.Context()
			if err := cli.UpdateTestsLabels(ctx, runID, testIDs, labels); err != nil {
				return fmt.Errorf("failed to update labels: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✅ Labels updated for %d tests in run %d: %v\n", len(testIDs), runID, labels)
			return nil
		},
	}

	cmd.Flags().Int64("run-id", 0, "ID тестового прогона (обязательно)")
	cmd.Flags().String("test-ids", "", "Список ID тестов через запятую (обязательно)")
	cmd.Flags().String("labels", "", "Список меток через запятую (обязательно)")

	_ = cmd.MarkFlagRequired("run-id")
	_ = cmd.MarkFlagRequired("test-ids")
	_ = cmd.MarkFlagRequired("labels")

	return cmd
}

// parseLabels splits a comma-separated string into a list of label names.
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

// parseIntList splits a comma-separated string into a list of int64 IDs.
func parseIntList(s string) []int64 {
	var ids []int64
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := flags.ParseID(part)
		if err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}
