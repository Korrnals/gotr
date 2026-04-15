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

	// Ask for optional label (from flag or interactively).
	label := snap.ResolveLabel(cmd)
	if label == "" {
		val, err := p.Input("🏷  Snapshot label (optional, press Enter to skip):", "")
		if err == nil {
			label = val
		}
	}

	return snapshotDecision{Create: true, Label: label}
}

// printFilterSummary prints a human-readable summary of the filter operation.
func printFilterSummary(entityName string, stats migration.FilterStats) {
	w := os.Stdout
	ui.Infof(w, "─── Filter result: %s ───", entityName)
	ui.Infof(w, "  Source:     %d (total in source)", stats.Source)
	ui.Infof(w, "  Target:     %d (total in destination)", stats.Target)
	ui.Infof(w, "  Matched:   %d (already exist in destination)", stats.Duplicates)
	ui.Infof(w, "  New:        %d (will be imported)", stats.New)
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

// syncPostAction shows a post-migration action menu when a snapshot was taken.
// hook is the Hook returned by HookMutation (may be nil or disabled).
// cli is the API client for rollback operations.
func syncPostAction(ctx context.Context, cmd *cobra.Command, hook *snap.Hook, cli client.ClientInterface) {
	if hook == nil || !hook.Enabled || hook.Snap == nil {
		return
	}

	snapID := hook.Snap.Meta.ID
	label := hook.Snap.Meta.Label
	if label != "" {
		ui.Infof(os.Stdout, "📦 Snapshot saved: %s (🏷 %s)", snapID, label)
	} else {
		ui.Infof(os.Stdout, "📦 Snapshot saved: %s", snapID)
	}

	p := interactive.PrompterFromContext(ctx)

	const (
		optExit     = "✕ Exit"
		optRollback = "↻ Rollback this migration"
	)

	for {
		_, choice, err := p.Select("Post-migration:", []string{optExit, optRollback})
		if err != nil {
			return
		}

		switch choice {
		case optRollback:
			ok, err := p.Confirm("⚠ Are you sure you want to rollback?", false)
			if err != nil || !ok {
				continue
			}

			result, err := snap.Rollback(ctx, cli, hook.Store, hook.Manifest, snapID)
			if err != nil {
				ui.Error(os.Stdout, fmt.Sprintf("Rollback failed: %v", err))
				return
			}
			ui.Successf(os.Stdout, "✓ Rollback complete: %s", result.Message)
			return

		default: // Exit
			return
		}
	}
}
