package interactive

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// NavTarget identifies a cross-navigation destination.
type NavTarget struct {
	Label string   // Display label in the menu.
	Path  []string // Cobra command path (e.g. ["compare", "all"]).
}

// Standard navigation targets used across post-action menus.
var (
	NavCompareAll = NavTarget{Label: "📊 Compare: verify current state", Path: []string{"compare", "all"}}
	NavSyncFull   = NavTarget{Label: "🔄 Sync: migrate data", Path: []string{"sync", "full"}}
	NavSnapList   = NavTarget{Label: "📦 Snap: manage snapshots", Path: []string{"snap", "list"}}
)

// NavigateMenu shows a navigation sub-menu with the given targets and
// dispatches the selected command via RunSubcommand.
// Returns nil on successful dispatch or exit; returns ErrGoBack when the user
// chooses "← Back".
func NavigateMenu(ctx context.Context, cmd *cobra.Command, prompt string, targets []NavTarget) error {
	p := PrompterFromContext(ctx)

	labels := make([]string, len(targets))
	for i, t := range targets {
		labels[i] = t.label()
	}

	idx, err := Browse(ctx, p, BrowseConfig{
		Prompt:    prompt,
		Items:     labels,
		AllowBack: true,
	})
	if err != nil {
		return err // ErrGoBack or ErrExit
	}

	target := targets[idx]
	fmt.Println()
	if runErr := RunSubcommand(ctx, cmd.Root(), target.Path...); runErr != nil {
		return runErr
	}
	fmt.Println()
	return nil
}

func (t NavTarget) label() string {
	return t.Label
}
