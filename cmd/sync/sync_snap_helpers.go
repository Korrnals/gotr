package sync

import (
	"context"
	"fmt"
	"os"

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

// printPreConfirmSummary prints a summary of migration parameters before final confirmation.
func printPreConfirmSummary(count int, entityName string, sd snapshotDecision) {
	w := os.Stdout
	ui.Infof(w, "─── Migration summary ───")
	ui.Infof(w, "  Import:    %d %s", count, entityName)
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
