package interactive

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// MutationPostAction shows a minimal post-action menu after a data mutation.
// Options: Exit (default), View snapshots.
// Silently returns when not in interactive mode or prompter is unavailable.
func MutationPostAction(ctx context.Context, cmd *cobra.Command) {
	if !HasPrompterInContext(ctx) || IsNonInteractive(ctx) {
		return
	}

	p := PrompterFromContext(ctx)
	key, err := ActionMenu(ctx, p, "What next?", []ActionOption{
		{Label: OptExit, Key: "exit"},
		{Label: "📦 View snapshots", Key: "snap"},
	})
	if err != nil || key == "exit" {
		return
	}
	if key == "snap" {
		fmt.Println()
		_ = RunSubcommand(ctx, cmd.Root(), "snap", "list")
		fmt.Println()
	}
}
