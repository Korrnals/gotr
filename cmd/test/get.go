package test

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/service"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду для получения информации о тесте
func newGetCmd(getClient func(cmd *cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [test-id]",
		Short: "Получить информацию о тесте",
		Long: `Получает детальную информацию о тесте по его ID.

Примеры:
	# Получить тест по ID
	gotr test get 12345

	# Получить и сохранить в файл
	gotr test get 12345 -o test.json
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			httpClient := getClient(cmd)
			ctx := cmd.Context()
			if httpClient == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			svc := service.NewTestService(httpClient)

			var testID int64
			var err error

			if len(args) > 0 {
				testID, err = svc.ParseID(ctx, args, 0)
				if err != nil {
					return fmt.Errorf("invalid test ID: %w", err)
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("test_id is required in non-interactive mode: gotr test get [test_id]")
				}
				if _, ok := interactive.PrompterFromContext(ctx).(*interactive.NonInteractivePrompter); ok {
					return fmt.Errorf("test_id is required in non-interactive mode: gotr test get [test_id]")
				}
				testID, err = resolveTestIDInteractive(ctx, httpClient)
				if err != nil {
					return err
				}
			}

			test, err := svc.Get(ctx, testID)
			if err != nil {
				return fmt.Errorf("failed to get test: %w", err)
			}

			// Проверяем нужно ли сохранить в файл
			saveFlag, _ := cmd.Flags().GetBool("save")
			if saveFlag {
				filepath, err := output.Output(cmd, test, "test", "json")
				if err != nil {
					return fmt.Errorf("save error: %w", err)
				}
				if filepath != "" {
					svc.PrintSuccess(ctx, cmd, "Тест сохранён в %s", filepath)
				}
				return nil
			}

			return svc.Output(ctx, cmd, test)
		},
	}

	output.AddFlag(cmd)
	cmd.Flags().BoolP("quiet", "q", false, "Тихий режим")

	return cmd
}
