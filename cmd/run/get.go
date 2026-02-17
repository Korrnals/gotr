package run

import (
	"fmt"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду 'run get'
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
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			svc := newRunServiceFromInterface(cli)
			runID, err := svc.ParseID(args, 0)
			if err != nil {
				return fmt.Errorf("некорректный ID test run: %w", err)
			}

			// Проверяем dry-run режим
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := dryrun.New("run get")
				dr.PrintOperation(
					fmt.Sprintf("Get Run %d", runID),
					"GET",
					fmt.Sprintf("/index.php?/api/v2/get_run/%d", runID),
					nil,
				)
				return nil
			}

			run, err := svc.Get(runID)
			if err != nil {
				return fmt.Errorf("ошибка получения test run: %w", err)
			}

			return svc.Output(cmd, run)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")

	return cmd
}

// getCmd — экспортированная команда
var getCmd = newGetCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClientSafe(cmd)
})
