package tests

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду для получения информации о тесте
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [test_id]",
		Short: "Получить информацию о тесте",
		Long: `Получить детальную информацию о тесте по его ID.

Тест представляет собой конкретное выполнение тест-кейса
в рамках тестового прогона.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := getClient(cmd)
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
					return fmt.Errorf("test_id is required in non-interactive mode: gotr tests get [test_id]")
				}
				if _, ok := interactive.PrompterFromContext(ctx).(*interactive.NonInteractivePrompter); ok {
					return fmt.Errorf("test_id is required in non-interactive mode: gotr tests get [test_id]")
				}
				testID, err = resolveTestIDInteractive(ctx, client)
				if err != nil {
					return err
				}
			}

			// Проверяем dry-run режим
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("tests get")
				dr.PrintSimple("Получить информацию о тесте", fmt.Sprintf("Test ID: %d", testID))
				return nil
			}

			test, err := client.GetTest(ctx, testID)
			if err != nil {
				return err
			}

			return output.OutputResult(cmd, test, "tests")
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	output.AddFlag(cmd)

	return cmd
}
