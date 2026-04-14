package snap

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/interactive"
	snaplib "github.com/Korrnals/gotr/internal/snap"
)

// selectSnapshot shows an interactive picker for available snapshots.
// Returns the selected snapshot ID. If no snapshots exist, returns an error.
func selectSnapshot(ctx context.Context, manifest *snaplib.Manifest, prompt string) (string, error) {
	entries := manifest.All()
	if len(entries) == 0 {
		return "", fmt.Errorf("no snapshots found")
	}

	p := interactive.PrompterFromContext(ctx)

	options := make([]string, 0, len(entries))
	for i, e := range entries {
		options = append(options, fmt.Sprintf("[%d] %s | %s %s | %s | %s",
			i+1, e.ID, e.Operation, e.EntityType, e.Status, e.Timestamp.Format("2006-01-02 15:04:05")))
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
