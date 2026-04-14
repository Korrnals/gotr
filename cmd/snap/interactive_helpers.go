package snap

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/interactive"
	snaplib "github.com/Korrnals/gotr/internal/snap"
)

// selectSnapshot shows an interactive picker for available snapshots.
// Entries are shown with their server URL context.
// Returns the selected snapshot ID. If no snapshots exist, returns an error.
func selectSnapshot(ctx context.Context, manifest *snaplib.Manifest, prompt string) (string, error) {
	entries := manifest.All()
	if len(entries) == 0 {
		return "", fmt.Errorf("no snapshots found")
	}

	p := interactive.PrompterFromContext(ctx)

	options := make([]string, 0, len(entries))
	for i, e := range entries {
		server := e.ServerURL
		if server == "" {
			server = "(unknown)"
		}
		name := ""
		if e.Name != "" {
			name = fmt.Sprintf(" %q", e.Name)
		}
		options = append(options, fmt.Sprintf("[%d] %s %s%s | %s | T%d | %s | %s",
			i+1, e.Operation, e.EntityType, name, e.Status,
			e.RollbackTier, e.Timestamp.Format("2006-01-02 15:04"), server))
	}

	idx, _, err := p.Select(prompt, options)
	if err != nil {
		return "", fmt.Errorf("failed to select snapshot: %w", err)
	}

	return entries[idx].ID, nil
}

// resolveSnapshotID returns the snapshot ID from args or interactive selection.
// If args has an element, returns args[0]. Otherwise, prompts interactively.
// In non-interactive mode without args, returns an error.
func resolveSnapshotID(ctx context.Context, args []string, manifest *snaplib.Manifest, prompt string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	if interactive.IsNonInteractive(ctx) {
		return "", fmt.Errorf("snapshot_id required in non-interactive mode")
	}

	return selectSnapshot(ctx, manifest, prompt)
}
