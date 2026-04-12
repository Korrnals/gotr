package run

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/service"
	"github.com/Korrnals/gotr/internal/ui"
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

// Create delegates run creation to the underlying run service.
func (w *runServiceWrapper) Create(ctx context.Context, projectID int64, req *data.AddRunRequest) (*data.Run, error) {
	return w.svc.Create(ctx, projectID, req)
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
	return &runServiceWrapper{svc: service.NewRunService(cli)}
}

// newDeleteCmd creates the 'run delete' command.
// Endpoint: POST /delete_run/{run_id}
func newDeleteCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [run-id]",
		Short: "Delete a test run",
		Long: `Deletes a test run by its ID.

WARNING: This action is irreversible!

When a run is deleted:
- All test results will be deleted
- All tests will be deleted
- The run structure itself will be deleted
- Cases in the suite will remain untouched

It is recommended to close a run (gotr run close) rather than delete it.

Examples:
	# Delete a run (no confirmation — use with caution!)
	gotr run delete 12345

	# Delete in quiet mode (for scripts)
	gotr run delete 12345 -q

	# Dry-run mode
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

			quiet, _ := cmd.Flags().GetBool("quiet")
			_, err = ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Deleting run",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (struct{}, error) {
				return struct{}{}, svc.Delete(ctx, runID)
			})
			if err != nil {
				return fmt.Errorf("failed to delete test run: %w", err)
			}

			output.PrintSuccess(cmd, "Test run %d deleted successfully", runID)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be executed without making actual changes")

	return cmd
}

// deleteCmd is used for registration in Register.
var deleteCmd = newDeleteCmd(getClientSafe)
