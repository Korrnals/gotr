package run

import (
	"fmt"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// newCloseCmd создаёт команду 'run close'
func newCloseCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close [run-id]",
		Short: "Закрыть test run",
		Long: `Закрывает test run (отмечает как завершённый).

Закрытый test run:
- Нельзя изменять (update вернёт ошибку)
- Нельзя добавлять результаты тестов
- Сохраняется в системе для истории и отчётности
- Поле is_completed становится true

Это действие обратимо — можно открыть run заново через веб-интерфейс TestRail.

Примеры:
	# Закрыть run после завершения тестирования
	gotr run close 12345

	# Закрыть и сохранить информацию о закрытом run
	gotr run close 12345 -o closed_run.json

	# Dry-run режим
	gotr run close 12345 --dry-run`,
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
				dr := dryrun.New("run close")
				dr.PrintOperation(
					fmt.Sprintf("Close Run %d", runID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/close_run/%d", runID),
					nil,
				)
				return nil
			}

			run, err := svc.Close(runID)
			if err != nil {
				return fmt.Errorf("ошибка закрытия test run: %w", err)
			}

			svc.PrintSuccess(cmd, "Test run закрыт успешно:")
			return svc.Output(cmd, run)
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")

	return cmd
}

// closeCmd — экспортированная команда
var closeCmd = newCloseCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClientSafe(cmd)
})
