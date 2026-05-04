package sync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/service/migration"
	"github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// snapshotDecision holds the result of the pre-migration snapshot prompt.
type snapshotDecision struct {
	Create bool
	Label  string
}

// confirmSnapshot implements smart confirm for snapshot creation before sync.
// Priority: explicit --snapshot flag > config snap.enabled > interactive prompt.
// Default answer is true (create snapshot).
// When a snapshot will be created, prompts for an optional label.
// Must be called BEFORE the final "Continue?" confirmation.
func confirmSnapshot(ctx context.Context, cmd *cobra.Command) snapshotDecision {
	p := interactive.PrompterFromContext(ctx)

	var create bool
	if cmd.Flags().Changed(snap.FlagSnapshot) || viper.IsSet("snap.enabled") {
		create = snap.ResolveDecision(cmd)
	} else {
		// Smart: ask user interactively.
		ok, err := p.Confirm("📦 Create snapshot before migration? (recommended)", true)
		if err != nil {
			// On error (e.g. non-interactive mode) default to creating snapshot.
			create = true
		} else {
			create = ok
		}
	}

	if !create {
		return snapshotDecision{}
	}

	// Label from explicit flag has highest priority.
	label := snap.ResolveLabel(cmd)
	if label != "" {
		return snapshotDecision{Create: true, Label: label}
	}

	// Auto-generate default label when user didn't provide one.
	mode := snap.ModeInteractive
	if interactive.IsNonInteractive(ctx) {
		mode = snap.ModeBatch
	}
	label = snap.GenerateDefaultLabel(mode, cmd.Name())

	return snapshotDecision{Create: true, Label: label}
}

// printFilterSummary prints a human-readable summary of the filter operation.
// After the multiset matcher (match.go + Bucket) the stats carry two extra
// fields compared to the legacy Duplicates counter: AlreadyInTarget (source
// rows matched 1:1 to a target row) and SrcCollapsed (source rows that share
// a MatchKey with another source row — a diagnostic signal, these will still
// be imported as separate rows in the target).
func printFilterSummary(entityName string, stats migration.FilterStats) {
	w := os.Stdout
	ui.Infof(w, "─── Filter result: %s ───", entityName)
	ui.Infof(w, "  Source:            %d (total in source)", stats.Source)
	ui.Infof(w, "  Target:            %d (total in destination)", stats.Target)
	ui.Infof(w, "  Already in target: %d (matched 1:1, will be skipped)", stats.AlreadyInTarget)
	if stats.SrcCollapsed > 0 {
		ui.Infof(w, "  Source duplicates: %d (same match key — imported as distinct rows)", stats.SrcCollapsed)
	}
	ui.Infof(w, "  New:               %d (will be imported)", stats.New)
}

// pluralizeEntity returns the singular form of an entity name when count==1,
// otherwise the input as-is. Handles the small set of entity nouns used in
// pre-confirm summaries: "cases" / "suites" / "sections" / "shared steps".
// Unknown nouns are returned unchanged.
func pluralizeEntity(count int, name string) string {
	if count != 1 {
		return name
	}
	switch name {
	case "cases":
		return "case"
	case "suites":
		return "suite"
	case "sections":
		return "section"
	case "shared steps":
		return "shared step"
	}
	return name
}

// reportOrphanSharedSteps prints a single concise note when source cases
// contained references to shared_step_ids that could not be resolved against
// the mapping. This is typically caused by upstream deletion of a shared step
// while the cases still carry the orphan reference (TestRail keeps the inline
// expansion in the case payload). Such steps are imported as inline content
// (graceful degradation) — the note is informational, not an error.
func reportOrphanSharedSteps(m *migration.Migration) {
	ids, hits := m.UnmappedSharedStepRefs()
	if hits == 0 {
		return
	}
	ui.Infof(os.Stdout,
		"Note: %d step occurrence(s) in %d unique source shared step(s) %v could not be mapped (likely deleted upstream); imported as inline content.",
		hits, len(ids), ids)
}

// printPreConfirmSummary prints a summary of migration parameters before final confirmation.
//
// `count` is the number of entities to import. Use a negative value when the
// count is not yet known at confirm time (e.g. `sync full` runs filtering for
// shared steps and cases inside its own phases, after the user has already
// approved the operation); in that case the helper prints an `Operation:`
// line instead of a misleading `Import: 0 ...` line.
func printPreConfirmSummary(count int, entityName string, sd snapshotDecision) {
	w := os.Stdout
	ui.Infof(w, "─── Migration summary ───")
	if count < 0 {
		ui.Infof(w, "  Operation: %s", entityName)
	} else {
		ui.Infof(w, "  Import:    %d %s", count, pluralizeEntity(count, entityName))
	}
	if sd.Create {
		if sd.Label != "" {
			ui.Infof(w, "  Snapshot:  ✓ enabled (🏷 %s)", sd.Label)
		} else {
			ui.Infof(w, "  Snapshot:  ✓ enabled")
		}
	} else {
		ui.Infof(w, "  Snapshot:  ✗ disabled")
	}
}

// artifactSet bundles paths of artifacts produced by a migration command, so
// that the operator can locate them without scrolling the log.
//
// Empty fields are skipped during printing.
type artifactSet struct {
	MigrationLog    string // ~/.gotr/logs/migration_*.json
	MappingFile     string // mapping_*.json (if --save-mapping)
	CasesLog        string // sync_cases JSON log file (if any)
	MigrationReport string // markdown report path returned by saveMigrationReport
	SnapshotDir     string // snapshot directory on disk (if a snapshot was created)
	SnapshotID      string // snapshot identifier (e.g. "sync/20260428T001036_full_p30_to_p34")
}

// printArtifacts emits a uniform "Artifacts" footer listing every produced
// path plus a "Hints" block with ready-to-run gotr commands. Silently skips
// empty entries; if the whole set is empty, prints nothing.
func printArtifacts(a artifactSet) {
	type row struct {
		label string
		value string
	}
	rows := []row{
		{"Migration log", a.MigrationLog},
		{"Mapping", a.MappingFile},
		{"Cases log", a.CasesLog},
		{"Report", a.MigrationReport},
		{"Snapshot", a.SnapshotDir},
	}
	hasAny := false
	for _, r := range rows {
		if r.value != "" {
			hasAny = true
			break
		}
	}
	if !hasAny {
		return
	}
	w := os.Stdout
	ui.Infof(w, "─── Artifacts ───")
	for _, r := range rows {
		if r.value == "" {
			continue
		}
		ui.Infof(w, "  %-14s %s", r.label+":", r.value)
	}

	hints := buildArtifactHints(a)
	if len(hints) == 0 {
		return
	}
	ui.Infof(w, "─── Hints ───")
	for _, h := range hints {
		ui.Infof(w, "  %s", h)
	}
}

// buildArtifactHints produces a list of ready-to-copy gotr commands for the
// produced artifacts. Returns an empty slice when nothing is actionable.
func buildArtifactHints(a artifactSet) []string {
	var hints []string
	if a.MigrationReport != "" {
		name := filepath.Base(a.MigrationReport)
		hints = append(hints,
			fmt.Sprintf("View report:    gotr report show %s", name),
			fmt.Sprintf("Print report:   gotr report show %s --print", name),
			"List reports:   gotr report list",
		)
	}
	if a.SnapshotID != "" {
		hints = append(hints,
			fmt.Sprintf("Snap details:   gotr snap info %s", a.SnapshotID),
			fmt.Sprintf("Rollback:       gotr snap rollback %s", a.SnapshotID),
			fmt.Sprintf("Bundle:         gotr export migration-archive %s", a.SnapshotID),
		)
	}
	return hints
}

// snapshotDirFromHook returns the absolute filesystem path of the snapshot
// captured by hook (if any), or "" if the hook is nil/disabled or no snapshot
// was actually written.
func snapshotDirFromHook(hook *snap.Hook) string {
	if hook == nil || !hook.Enabled || hook.Snap == nil || hook.Store == nil {
		return ""
	}
	id := hook.Snap.Meta.ID
	if id == "" {
		return ""
	}
	return hook.Store.SnapDir(id)
}

// snapshotIDFromHook returns the snapshot identifier (category/id) suitable
// for `gotr snap info|rollback` arguments, or "" when the hook produced no
// snapshot.
func snapshotIDFromHook(hook *snap.Hook) string {
	if hook == nil || !hook.Enabled || hook.Snap == nil {
		return ""
	}
	return hook.Snap.Meta.ID
}

// syncPostAction shows a post-migration action menu.
// hook is the Hook returned by HookMutation (may be nil or disabled).
// cli is the API client for rollback operations.
func syncPostAction(ctx context.Context, cmd *cobra.Command, hook *snap.Hook, cli client.ClientInterface) {
	if !interactive.HasPrompterInContext(ctx) || interactive.IsNonInteractive(ctx) {
		return
	}

	p := interactive.PrompterFromContext(ctx)

	hasSnap := hook != nil && hook.Enabled && hook.Snap != nil
	if hasSnap {
		snapID := hook.Snap.Meta.ID
		label := hook.Snap.Meta.Label
		if label != "" {
			ui.Infof(os.Stdout, "📦 Snapshot saved: %s (🏷 %s)", snapID, label)
		} else {
			ui.Infof(os.Stdout, "📦 Snapshot saved: %s", snapID)
		}
	}

	const (
		optExit     = "exit"
		optRollback = "rollback"
	)

	for {
		options := []interactive.ActionOption{
			{Label: interactive.OptExit, Key: optExit},
		}
		if hasSnap {
			options = append(options, interactive.ActionOption{Label: "↻ Rollback this migration", Key: optRollback})
		}
		options = append(options, interactive.CrossNavOptions()...)

		key, err := interactive.ActionMenu(ctx, p, "Post-migration:", options)
		if err != nil {
			return
		}

		switch key {
		case optExit:
			return
		case optRollback:
			ok, err := p.Confirm("⚠ Are you sure you want to rollback?", false)
			if err != nil || !ok {
				continue
			}

			result, err := snap.Rollback(ctx, cli, hook.Store, hook.Manifest, hook.Snap.Meta.ID)
			if err != nil {
				ui.Error(os.Stdout, fmt.Sprintf("Rollback failed: %v", err))
				return
			}
			ui.Successf(os.Stdout, "✓ Rollback complete: %s", result.Message)
			return
		default:
			interactive.HandleCrossNav(ctx, cmd, key)
			continue
		}
	}
}
