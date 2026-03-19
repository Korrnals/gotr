package sync

import (
	"context"
	"os"

	"github.com/Korrnals/gotr/internal/service/migration"
	"github.com/Korrnals/gotr/internal/ui"
)

// newMigration — seam для тестов; по умолчанию указывает на migration.NewMigration
var newMigration = migration.NewMigration

func newSyncOperation(title string) ui.Operation {
	return ui.NewOperation(ui.StatusConfig{
		Title:  title,
		Writer: os.Stderr,
	})
}

func runSyncStatus[T any](ctx context.Context, title string, fn func(context.Context) (T, error)) (T, error) {
	return ui.RunWithStatus(ctx, ui.StatusConfig{
		Title:  title,
		Writer: os.Stderr,
	}, fn)
}
