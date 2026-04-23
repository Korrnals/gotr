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

// runCoverageGate re-fetches cases from the source and destination suites and
// reports any source case that has no counterpart in target (per MatchKey).
// The --verify-coverage flag opts in to this behavior; when not set the gate
// is a no-op and the function returns nil so the caller may proceed.
//
// On gaps the function prints the missing cases (id + title), returns a
// non-nil error that wraps migration.CoverageReport, and the sync command is
// expected to propagate it to Cobra so the process exits non-zero — this is
// the difference between "0 errors, 1547 skipped" and a loud stop.
func runCoverageGate(ctx context.Context, cmd *cobra.Command, m *migration.Migration, quiet bool) error {
	enabled, _ := cmd.Flags().GetBool("verify-coverage")
	if !enabled {
		return nil
	}

	report, err := runSyncStatus(ctx, "Verifying coverage (post-import)...", quiet, func(ctx context.Context) (migration.CoverageReport, error) {
		srcCases, tgtCases, ferr := m.FetchCasesData(ctx)
		if ferr != nil {
			return migration.CoverageReport{}, ferr
		}
		return m.VerifyCasesCoverage(srcCases, tgtCases), nil
	})
	if err != nil {
		return fmt.Errorf("runCoverageGate: %w", err)
	}

	if len(report.Missing) == 0 {
		ui.Successf(os.Stdout, "Coverage OK: %d/%d source cases matched in target", report.Matched, report.Source)
		return nil
	}

	ui.Error(os.Stdout, fmt.Sprintf("Coverage gap: %d/%d source cases missing in target (%.1f%% coverage)",
		len(report.Missing), report.Source, report.Coverage*100))
	limit := len(report.Missing)
	if limit > 50 {
		limit = 50
	}
	for _, miss := range report.Missing[:limit] {
		fmt.Fprintf(os.Stdout, "  - [%d] %q (src_section=%d, dst_section=%d)\n",
			miss.SrcID, miss.Title, miss.SrcSectionID, miss.DstSectionID)
	}
	if len(report.Missing) > limit {
		fmt.Fprintf(os.Stdout, "  ... and %d more (see migration log)\n", len(report.Missing)-limit)
	}
	return fmt.Errorf("coverage gap: %d source cases missing in target", len(report.Missing))
}


