package sync

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/service/migration"
	"github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
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

// buildSyncData builds SyncData for snap rollback from created entities.
func buildSyncData(created []snap.SyncCreatedEntity, srcProject, dstProject, srcSuite, dstSuite int64) snap.SyncData {
	sd := snap.SyncData{
		SrcProject: srcProject,
		DstProject: dstProject,
		SrcSuite:   srcSuite,
		DstSuite:   dstSuite,
		Created:    created,
	}
	return sd
}

// resolveMatchField returns the compare field to feed into the migration
// layer. Priority: (1) explicit --compare-field flag from the user, (2)
// interactive SelectMatchField prompt when running in an interactive session,
// (3) the kind's default ("Title" / "Name").
//
// The value is normalized to the canonical case-insensitive form expected
// by migration.fieldValue's reflection.
func resolveMatchField(ctx context.Context, cmd *cobra.Command, kind interactive.MatchFieldKind) (string, error) {
	raw, _ := cmd.Flags().GetString("compare-field")
	// User set the flag explicitly — honor it as-is (normalized).
	if cmd.Flags().Changed("compare-field") {
		return interactive.NormalizeMatchField(kind, raw), nil
	}

	// Flag is at its default value — try interactive selection.
	p := interactive.PrompterFromContext(ctx)
	defaultField := interactive.NormalizeMatchField(kind, raw)
	selected, err := interactive.SelectMatchField(ctx, p, kind, defaultField)
	if err != nil {
		return "", fmt.Errorf("resolveMatchField: %w", err)
	}
	return selected, nil
}

