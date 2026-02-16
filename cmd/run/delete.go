package run

import (
	"fmt"

	"github.com/Korrnals/gotr/cmd/common/dryrun"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/service"
	"github.com/spf13/cobra"
)

// runServiceWrapper оборачивает сервис для работы с runs
type runServiceWrapper struct {
	svc *service.RunService
}

func (w *runServiceWrapper) Delete(runID int64) error {
	return w.svc.Delete(runID)
}

func (w *runServiceWrapper) ParseID(args []string, index int) (int64, error) {
	return w.svc.ParseID(args, index)
}

func (w *runServiceWrapper) PrintSuccess(cmd *cobra.Command, format string, args ...interface{}) {
	w.svc.PrintSuccess(cmd, format, args...)
}

func (w *runServiceWrapper) Create(projectID int64, req *data.AddRunRequest) (*data.Run, error) {
	return w.svc.Create(projectID, req)
}

func (w *runServiceWrapper) Output(cmd *cobra.Command, data interface{}) error {
	return w.svc.Output(cmd, data)
}

func (w *runServiceWrapper) Close(runID int64) (*data.Run, error) {
	return w.svc.Close(runID)
}

func (w *runServiceWrapper) Update(runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
	return w.svc.Update(runID, req)
}

func (w *runServiceWrapper) Get(runID int64) (*data.Run, error) {
	return w.svc.Get(runID)
}

func (w *runServiceWrapper) GetByProject(projectID int64) (data.GetRunsResponse, error) {
	return w.svc.GetByProject(projectID)
}

// newRunServiceFromInterface создаёт сервис из клиента-интерфейса
func newRunServiceFromInterface(cli client.ClientInterface) *runServiceWrapper {
	// Пытаемся привести к *HTTPClient, если это не mock
	if httpClient, ok := cli.(*client.HTTPClient); ok {
		return &runServiceWrapper{svc: service.NewRunService(httpClient)}
	}
	// Для тестов с mock - используем специальный конструктор
	return &runServiceWrapper{svc: service.NewRunServiceFromInterface(cli)}
}

// newDeleteCmd создаёт команду 'run delete'
// Эндпоинт: POST /delete_run/{run_id}
func newDeleteCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [run-id]",
		Short: "Удалить test run",
		Long: `Удаляет test run по его ID.

⚠️ ВНИМАНИЕ: Это действие необратимо!

При удалении run:
- Все результаты тестов будут удалены
- Все тесты (tests) будут удалены
- Сама структура run будет удалена
- Кейсы в сьюте останутся нетронутыми

Рекомендуется сначала закрыть run (gotr run close), а не удалять.

Примеры:
	# Удалить run (без подтверждения — осторожно!)
	gotr run delete 12345

	# Удалить в тихом режиме (для скриптов)
	gotr run delete 12345 -q

	# Dry-run режим
	gotr run delete 12345 --dry-run`,
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
				dr := dryrun.New("run delete")
				dr.PrintOperation(
					fmt.Sprintf("Delete Run %d", runID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/delete_run/%d", runID),
					nil,
				)
				return nil
			}

			if err := svc.Delete(runID); err != nil {
				return fmt.Errorf("ошибка удаления test run: %w", err)
			}

			svc.PrintSuccess(cmd, "Test run %d удалён успешно", runID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")

	return cmd
}

// deleteCmd используется для регистрации в Register
var deleteCmd = newDeleteCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClientSafe(cmd)
})
