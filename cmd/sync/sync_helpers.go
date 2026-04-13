package sync

import (
	"context"
	"os"

	"github.com/Korrnals/gotr/internal/service/migration"
	"github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
)

// newMigration is a test seam; defaults to migration.NewMigration.
var newMigration = migration.NewMigration

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

// buildSyncDataFromMapping converts a migration mapping into SyncData for snap rollback.
// The mapping contains sourceID→targetID for entities created during sync.
func buildSyncDataFromMapping(mapping map[int64]int64, srcProject, dstProject, srcSuite, dstSuite int64) snap.SyncData {
	sd := snap.SyncData{
		SrcProject: srcProject,
		DstProject: dstProject,
		SrcSuite:   srcSuite,
		DstSuite:   dstSuite,
	}
	for sourceID, targetID := range mapping {
		sd.Created = append(sd.Created, snap.SyncCreatedEntity{
			Type:     "sync_entity",
			SourceID: sourceID,
			TargetID: targetID,
		})
	}
	return sd
}
