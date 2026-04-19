package snap

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/interactive"
	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

func newRollbackListCmd(getClient GetClientFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Browse rolled-back snapshots",
		Long: `Shows only rolled-back snapshots in an interactive browser.
From the snapshot card you can undo the rollback (delete re-created entities).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cli := getClient(cmd)
			api, ok := cli.(snaplib.RollbackAPI)
			if !ok {
				return fmt.Errorf("client does not support undo operations")
			}

			store, err := snaplib.NewStore()
			if err != nil {
				return fmt.Errorf("newRollbackListCmd.func: %w", err)
			}

			manifest, err := snaplib.LoadManifest(store)
			if err != nil {
				return fmt.Errorf("newRollbackListCmd.func: %w", err)
			}

			return browseUndoSnapshots(cmd, api, store, manifest)
		},
	}
}

func newRollbackUndoCmd(getClient GetClientFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "undo [snapshot_id]",
		Short: "Undo a previous rollback",
		Long: `Reverses a previous rollback by deleting re-created entities.

Undo reverses the rollback effect by deleting re-created entities:
  • delete rollbacks: deletes the re-created entity (case, section, project)
  • After undo the snapshot status resets to "available"

Not all rollbacks are undoable:
  • add/copy rollbacks (entity was deleted) — cannot restore
  • update rollbacks (post-mutation values not saved) — cannot reverse
  • sync rollbacks (entities deleted) — re-sync required

If snapshot_id is omitted, shows an interactive picker with only rolled-back snapshots.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			api, ok := cli.(snaplib.RollbackAPI)
			if !ok {
				return fmt.Errorf("client does not support undo operations")
			}
			ctx := cmd.Context()

			store, err := snaplib.NewStore()
			if err != nil {
				return fmt.Errorf("newRollbackUndoCmd.func: %w", err)
			}

			manifest, err := snaplib.LoadManifest(store)
			if err != nil {
				return fmt.Errorf("newRollbackUndoCmd.func: %w", err)
			}

			// Non-interactive or explicit ID: undo directly.
			if len(args) > 0 || interactive.IsNonInteractive(ctx) {
				snapID, err := resolveSnapshotIDWith(ctx, args, manifest, "Select snapshot to undo:", &pickerOpts{
					statusFilter: []snaplib.Status{snaplib.StatusRolledBack},
				})
				if err != nil {
					return fmt.Errorf("newRollbackUndoCmd.func: %w", err)
				}
				return executeUndo(cmd, api, store, manifest, snapID)
			}

			// Interactive browse: only rolled-back snapshots.
			return browseUndoSnapshots(cmd, api, store, manifest)
		},
	}
}

// browseUndoSnapshots opens an interactive browser showing only rolled-back snapshots.
func browseUndoSnapshots(cmd *cobra.Command, api snaplib.RollbackAPI, store *snaplib.Store, manifest *snaplib.Manifest) error {
	entries := manifest.ListByStatus(snaplib.StatusRolledBack)
	if len(entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No rolled-back snapshots found.")
		return nil
	}

	p := interactive.PrompterFromContext(cmd.Context())
	groups := groupByServer(entries)

	for { // Server-level loop.
		var selected serverGroup

		if len(groups) == 1 {
			selected = groups[0]
		} else {
			options := make([]string, 0, len(groups)+1)
			options = append(options, exitOption)
			for _, g := range groups {
				options = append(options, fmt.Sprintf("%s — %d rolled back", serverLabel(g.URL), len(g.Entries)))
			}

			sPrompt := fmt.Sprintf("Rolled-back snapshots (%d servers, %d total):", len(groups), len(entries))
			idx, _, err := p.Select(sPrompt, options)
			if err != nil {
				return nil
			}
			if idx == 0 {
				return nil
			}
			selected = groups[idx-1]
			fmt.Fprintf(cmd.OutOrStdout(), "  ✓ %s\n", serverLabel(selected.URL))
		}

		err := browseUndoList(cmd, api, store, manifest, p, selected.Entries, len(groups) > 1)
		if err == errGoBack {
			if len(groups) == 1 {
				return nil
			}
			continue
		}
		if err == errExit {
			return nil
		}
		if err != nil {
			return nil
		}
	}
}

// browseUndoList shows the snapshot picker → card → undo menu for rolled-back snapshots.
//nolint:gocyclo // Interactive browser flow keeps explicit menu transitions.
func browseUndoList(cmd *cobra.Command, api snaplib.RollbackAPI, store *snaplib.Store, manifest *snaplib.Manifest, p interactive.Prompter, entries []snaplib.ManifestEntry, allowBack bool) error {
	labels := undoPickerLabels(store, entries)

	for {
		options := make([]string, 0, len(labels)+3)
		if allowBack {
			options = append(options, backOption)
		}
		options = append(options, exitOption)
		options = append(options, labels...)
		if allowBack {
			options = append(options, backOption)
		}

		snapPrompt := fmt.Sprintf("Select rolled-back snapshot (%d snapshots):", len(entries))

		idx, _, err := p.Select(snapPrompt, options)
		if err != nil {
			return errExit
		}

		if allowBack && (idx == 0 || idx == len(options)-1) {
			return errGoBack
		}

		exitIdx := 0
		if allowBack {
			exitIdx = 1
		}
		if idx == exitIdx {
			return errExit
		}

		offset := exitIdx + 1
		entry := entries[idx-offset]

		meta, err := store.LoadMeta(entry.ID)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Error loading snapshot: %v\n\n", err)
			continue
		}

		renderInfoCard(cmd, meta)
		fmt.Fprintln(cmd.OutOrStdout())

		action, err := postUndoCardAction(cmd, p, meta)
		if err != nil {
			return errExit
		}
		switch action {
		case postActionBack:
			continue
		case postActionExit:
			return errExit
		case postActionRollback: // reused as "undo" action
			executeUndoFromBrowser(cmd, api, store, manifest, entry.ID)
			// Re-fetch entries — status may have changed.
			entries = manifest.ListByStatus(snaplib.StatusRolledBack)
			if len(entries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "  No more rolled-back snapshots.")
				return errExit
			}
			labels = undoPickerLabels(store, entries)
			continue
		}
	}
}

// postUndoCardAction shows a mini-menu after viewing a rolled-back snapshot card.
// Loops on unavailable-undo selection so the user stays on the card.
func postUndoCardAction(cmd *cobra.Command, p interactive.Prompter, meta *snaplib.Meta) (postAction, error) {
	out := cmd.OutOrStdout()
	canUndo := snaplib.CanUndo(meta)

	for {
		options := []string{backOption, exitOption}
		if canUndo {
			options = append(options, "↩ Undo rollback")
		} else {
			fmt.Fprintf(out, "  ⚠ %s\n", undoHint(meta))
			options = append(options, "↩ Undo rollback (unavailable)")
		}

		idx, _, err := p.Select("Action:", options)
		if err != nil {
			return postActionExit, err
		}

		switch idx {
		case 0:
			return postActionBack, nil
		case 1:
			return postActionExit, nil
		case 2:
			if !canUndo {
				fmt.Fprintf(out, "\n  ⚠ %s\n", undoHint(meta))
				fmt.Fprintln(out, "  💡 Undo is only available for delete-operation rollbacks that re-created entities.")
				fmt.Fprintf(out, "     To re-rollback, use: gotr snap rollback %s\n\n", shortID(meta.ID))
				continue // Re-show action menu — user stays on card.
			}
			return postActionRollback, nil // reused as undo
		default:
			return postActionBack, nil
		}
	}
}

// undoHint returns a short user-facing explanation of why undo is unavailable.
func undoHint(meta *snaplib.Meta) string {
	switch {
	case meta.Operation == snaplib.OpAdd || meta.Operation == snaplib.OpCopy:
		return "Undo unavailable — entity was deleted during rollback."
	case meta.Operation == snaplib.OpUpdate:
		return "Undo unavailable — post-mutation values were not saved."
	case meta.IsSyncOp():
		return "Undo unavailable — re-sync required."
	default:
		return "Undo unavailable for this snapshot."
	}
}

// undoPickerLabels returns aligned picker labels enriched with an UNDO column.
func undoPickerLabels(store *snaplib.Store, entries []snaplib.ManifestEntry) []string {
	_, base := alignedPickerLabels(entries)
	labels := make([]string, len(entries))
	for i, e := range entries {
		tag := "✗ no undo"
		meta, err := store.LoadMeta(e.ID)
		if err == nil && snaplib.CanUndo(meta) {
			tag = "↩ undoable"
		}
		labels[i] = base[i] + " │ " + tag
	}
	return labels
}

// executeUndoFromBrowser runs undo from the browse flow.
func executeUndoFromBrowser(cmd *cobra.Command, api snaplib.RollbackAPI, store *snaplib.Store, manifest *snaplib.Manifest, snapID string) {
	p := interactive.PrompterFromContext(cmd.Context())
	confirmed, err := p.Confirm("Undo this rollback? Re-created entities will be deleted.", false)
	if err != nil || !confirmed {
		fmt.Fprintln(cmd.OutOrStdout(), "  Canceled.")
		return
	}

	quiet, _ := cmd.Flags().GetBool("quiet")
	result, err := ui.RunWithStatus(cmd.Context(), ui.StatusConfig{
		Title:  "Undoing rollback",
		Writer: os.Stderr,
		Quiet:  quiet,
	}, func(ctx context.Context) (*snaplib.UndoResult, error) {
		return snaplib.UndoRollback(ctx, api, store, manifest, snapID)
	})
	if err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "  ⚠ Undo failed: %v\n\n", err)
		return
	}

	ui.Successf(cmd.OutOrStdout(), "%s", result.Message)
	fmt.Fprintln(cmd.OutOrStdout())
}

// executeUndo runs undo in non-interactive mode or for a direct ID.
func executeUndo(cmd *cobra.Command, api snaplib.RollbackAPI, store *snaplib.Store, manifest *snaplib.Manifest, snapID string) error {
	entry := manifest.Find(snapID)
	if entry == nil {
		return fmt.Errorf("snapshot %q not found in manifest", snapID)
	}

	meta, err := store.LoadMeta(snapID)
	if err != nil {
		return fmt.Errorf("load snapshot: %w", err)
	}

	if !snaplib.CanUndo(meta) {
		return fmt.Errorf("undo not available: %s", undoHint(meta))
	}

	// Interactive: confirm before proceeding.
	prompter := interactive.PrompterFromContext(cmd.Context())
	if !interactive.IsNonInteractive(cmd.Context()) {
		printRollbackHeader(cmd, entry)
		fmt.Fprintf(cmd.OutOrStdout(), "This will delete re-created entities and reset snapshot to available.\n\n")

		confirmed, err := prompter.Confirm("Apply undo?", false)
		if err != nil {
			return wrapInterrupt(err)
		}
		if !confirmed {
			fmt.Fprintln(cmd.OutOrStdout(), "Undo canceled.")
			return nil
		}
	}

	quiet, _ := cmd.Flags().GetBool("quiet")
	result, err := ui.RunWithStatus(cmd.Context(), ui.StatusConfig{
		Title:  "Undoing rollback",
		Writer: os.Stderr,
		Quiet:  quiet,
	}, func(ctx context.Context) (*snaplib.UndoResult, error) {
		return snaplib.UndoRollback(ctx, api, store, manifest, snapID)
	})
	if err != nil {
		return fmt.Errorf("undo failed: %w", err)
	}

	ui.Successf(cmd.OutOrStdout(), "%s", result.Message)
	return nil
}
