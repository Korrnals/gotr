package run

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd creates the 'run get' command.
func newGetCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [run-id]",
		Short: "Получить информацию о test run",
		Long: `Получает детальную информацию о test run по его ID.

Test run — это экземпляр тест-сюиты, запущенный для выполнения тестов.
В ответе содержится: название, описание, статистика прохождения,
даты создания/обновления, assignedto_id и другие поля.

Примеры:
	# Получить информацию о run
	gotr run get 12345

	# Сохранить результат в файл
	gotr run get 12345 -o run_info.json

	# Dry-run режим
	gotr run get 12345 --dry-run
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			runID, err := resolveRunID(ctx, cli, args)
			if err != nil {
				return fmt.Errorf("invalid test run ID: %w", err)
			}

			svc := newRunServiceFromInterface(cli)

			// Check dry-run mode
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := output.NewDryRunPrinter("run get")
				dr.PrintOperation(
					fmt.Sprintf("Get Run %d", runID),
					"GET",
					fmt.Sprintf("/index.php?/api/v2/get_run/%d", runID),
					nil,
				)
				return nil
			}

			run, err := svc.Get(ctx, runID)
			if err != nil {
				return fmt.Errorf("failed to get test run: %w", err)
			}

			return output.OutputResultWithFlags(cmd, run)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")

	return cmd
}

// getCmd is the exported command.
var getCmd = newGetCmd(getClientSafe)
