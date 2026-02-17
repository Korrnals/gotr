package tests

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду для получения информации о тесте
func newGetCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <test_id>",
		Short: "Получить информацию о тесте",
		Long: `Получить детальную информацию о тесте по его ID.

Тест представляет собой конкретное выполнение тест-кейса
в рамках тестового прогона.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()
			client := getClient(cmd)

			testID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || testID <= 0 {
				return fmt.Errorf("test_id должен быть положительным числом")
			}

			// Проверяем dry-run режим
			if isDryRun, _ := cmd.Flags().GetBool("dry-run"); isDryRun {
				dr := output.NewDryRunPrinter("tests get")
				dr.PrintSimple("Получить информацию о тесте", fmt.Sprintf("Test ID: %d", testID))
				return nil
			}

			test, err := client.GetTest(testID)
			if err != nil {
				return err
			}

			return outputResult(cmd, test, start)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать, что будет сделано без изменений")
	output.AddFlag(cmd)

	return cmd
}
