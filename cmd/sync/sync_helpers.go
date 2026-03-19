package sync

import (
	"context"
	"os"

	"github.com/Korrnals/gotr/internal/service/migration"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

// newMigration — seam для тестов; по умолчанию указывает на migration.NewMigration
var newMigration = migration.NewMigration

func isQuiet(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	quiet, _ := cmd.Flags().GetBool("quiet")
	return quiet
}

func newSyncOperation(title string, quiet bool) ui.Operation {
	return ui.NewOperation(ui.StatusConfig{
		Title:  title,
		Writer: os.Stderr,
		Quiet:  quiet,
	})
}

func runSyncStatus[T any](ctx context.Context, title string, quiet bool, fn func(context.Context) (T, error)) (T, error) {
	return ui.RunWithStatus(ctx, ui.StatusConfig{
		Title:  title,
		Writer: os.Stderr,
		Quiet:  quiet,
	}, fn)
}
