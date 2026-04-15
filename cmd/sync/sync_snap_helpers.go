package sync

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// confirmSnapshot implements smart confirm for snapshot creation before sync.
// Priority: explicit --snapshot flag > config snap.enabled > interactive prompt.
// Default answer is true (create snapshot).
func confirmSnapshot(ctx context.Context, cmd *cobra.Command) bool {
	// If explicitly set via flag or config — use existing resolution, no prompt.
	if cmd.Flags().Changed(snap.FlagSnapshot) || viper.IsSet("snap.enabled") {
		return snap.ResolveDecision(cmd)
	}

	// Smart: ask user interactively.
	p := interactive.PrompterFromContext(ctx)
	ok, err := p.Confirm("📦 Create snapshot before migration? (recommended)", true)
	if err != nil {
		// On error (e.g. non-interactive mode) default to creating snapshot.
		return true
	}
	return ok
}

// syncPostAction shows a post-migration action menu when a snapshot was taken.
// hook is the Hook returned by HookMutation (may be nil or disabled).
// cli is the API client for rollback operations.
func syncPostAction(ctx context.Context, cmd *cobra.Command, hook *snap.Hook, cli client.ClientInterface) {
	if hook == nil || !hook.Enabled || hook.Snap == nil {
		return
	}

	snapID := hook.Snap.Meta.ID
	ui.Infof(os.Stdout, "📦 Snapshot saved: %s", snapID)

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
