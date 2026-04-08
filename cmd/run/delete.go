package run

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/service"
	"github.com/spf13/cobra"
)

// runServiceWrapper wraps the service for working with runs.
type runServiceWrapper struct {
	svc *service.RunService
}

// Delete delegates run deletion to the underlying run service.
func (w *runServiceWrapper) Delete(ctx context.Context, runID int64) error {
	return w.svc.Delete(ctx, runID)
}

// ParseID delegates run ID parsing to the underlying run service.
func (w *runServiceWrapper) ParseID(ctx context.Context, args []string, index int) (int64, error) {
	return w.svc.ParseID(ctx, args, index)
}

// PrintSuccess delegates success message formatting to the underlying run service.
func (w *runServiceWrapper) PrintSuccess(ctx context.Context, cmd *cobra.Command, format string, args ...interface{}) {
	w.svc.PrintSuccess(ctx, cmd, format, args...)
}

// Create delegates run creation to the underlying run service.
func (w *runServiceWrapper) Create(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error) {
	return w.svc.Create(ctx, projectID, req)
}

// Output delegates command output formatting to the underlying run service.
func (w *runServiceWrapper) Output(ctx context.Context, cmd *cobra.Command, v interface{}) error {
	return w.svc.Output(ctx, cmd, v)
}

// Close delegates run closing to the underlying run service.
func (w *runServiceWrapper) Close(ctx context.Context, runID int64) (*data.Run, error) {
	return w.svc.Close(ctx, runID)
}

// Update delegates run updates to the underlying run service.
func (w *runServiceWrapper) Update(ctx context.Context, runID int64, req *data.UpdateRunRequest) (*data.Run, error) {
	return w.svc.Update(ctx, runID, req)
}

// Get delegates run retrieval by ID to the underlying run service.
func (w *runServiceWrapper) Get(ctx context.Context, runID int64) (*data.Run, error) {
	return w.svc.Get(ctx, runID)
}

// GetByProject delegates project run listing to the underlying run service.
func (w *runServiceWrapper) GetByProject(ctx context.Context, projectID int64) (data.GetRunsResponse, error) {
	return w.svc.GetByProject(ctx, projectID)
}

// newRunServiceFromInterface creates a service from a client interface.
func newRunServiceFromInterface(cli client.ClientInterface) *runServiceWrapper {
	// Try to cast to *HTTPClient if not a mock
	if httpClient, ok := cli.(*client.HTTPClient); ok {
		return &runServiceWrapper{svc: service.NewRunService(httpClient)}
	}
	// For tests with mock — use a special constructor
	return &runServiceWrapper{svc: service.NewRunServiceFromInterface(cli)}
}

// newDeleteCmd creates the 'run delete' command.
// Endpoint: POST /delete_run/{run_id}
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
				dr := output.NewDryRunPrinter("run delete")
				dr.PrintOperation(
					fmt.Sprintf("Delete Run %d", runID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/delete_run/%d", runID),
					nil,
				)
				return nil
			}

			if err := svc.Delete(ctx, runID); err != nil {
				return fmt.Errorf("failed to delete test run: %w", err)
			}

			svc.PrintSuccess(ctx, cmd, "Test run %d удалён успешно", runID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")

	return cmd
}

// deleteCmd is used for registration in Register.
var deleteCmd = newDeleteCmd(func(cmd *cobra.Command) client.ClientInterface {
	return getClientSafe(cmd)
})
