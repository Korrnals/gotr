package snap

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/Korrnals/gotr/internal/interactive"
	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

func newRollbackCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback [snapshot_id]",
		Short: "Rollback a mutation using a snapshot",
		Long: `Reverses a mutation by applying the saved pre-mutation state.

For update operations: restores original field values (Tier 1 — full rollback).
For delete operations: re-creates the entity with a new ID (Tier 2).
For add operations: deletes the created entity.

If snapshot_id is omitted, shows an interactive picker.

Flags:
  --dry-run       Preview changes without applying them (shows diff table)
  --entity-ids    Limit rollback to specific entity IDs (comma-separated)

Partial failures are resumable: re-run the same rollback to retry failed entities.

Subcommands:
  list   Browse rolled-back snapshots interactively
  undo   Undo a previous rollback (delete re-created entities)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			store, err := snaplib.NewStore()
			if err != nil {
				return fmt.Errorf("newRollbackCmd.func: %w", err)
			}

			manifest, err := snaplib.LoadManifest(store)
			if err != nil {
				return fmt.Errorf("newRollbackCmd.func: %w", err)
			}

			snapID, err := resolveSnapshotIDWith(ctx, args, manifest, "Select snapshot to rollback:", &pickerOpts{
				statusFilter: []snaplib.Status{snaplib.StatusAvailable, snaplib.StatusRollbackPartial},
			})
			if err != nil {
				return fmt.Errorf("newRollbackCmd.func: %w", err)
			}

			entry := manifest.Find(snapID)
			if entry == nil {
				return fmt.Errorf("snapshot %q not found in manifest", snapID)
			}

			// Parse options.
			opts := snaplib.RollbackOpts{}
			opts.DryRun, _ = cmd.Flags().GetBool("dry-run")

			if raw, _ := cmd.Flags().GetString("entity-ids"); raw != "" {
				ids, err := parseEntityIDs(raw)
				if err != nil {
					return fmt.Errorf("invalid --entity-ids: %w", err)
				}
				opts.EntityIDs = ids
			}

			// Dry-run: show diff preview and exit.
			if opts.DryRun {
				result, err := snaplib.Rollback(ctx, cli, store, manifest, snapID, opts)
				if err != nil {
					return fmt.Errorf("dry-run failed: %w", err)
				}
				printRollbackHeader(cmd, entry)
				printDiffPreview(cmd, result)
				return nil
			}

			// Interactive mode: show preview + confirm.
			yes, _ := cmd.Flags().GetBool("yes")
			if canceled, err := maybeConfirmRollback(ctx, cmd, cli, store, manifest, entry, snapID, opts, yes); err != nil {
				return err
			} else if canceled {
				return nil
			}

			// Execute rollback.
			quiet, _ := cmd.Flags().GetBool("quiet")
			result, err := ui.RunWithStatus(ctx, ui.StatusConfig{
				Title:  "Rolling back",
				Writer: os.Stderr,
				Quiet:  quiet,
			}, func(ctx context.Context) (*snaplib.RollbackResult, error) {
				return snaplib.Rollback(ctx, cli, store, manifest, snapID, opts)
			})
			if err != nil {
				return fmt.Errorf("rollback failed: %w", err)
			}

			ui.Successf(os.Stdout, "Rollback complete: %s", result.Message)
			printSyncRollbackDetails(os.Stdout, result)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview changes without applying them")
	cmd.Flags().String("entity-ids", "", "Limit rollback to specific entity IDs (comma-separated)")
	cmd.Flags().BoolP("yes", "y", false, "Skip the interactive confirmation prompt")

	cmd.AddCommand(newRollbackListCmd(getClient))
	cmd.AddCommand(newRollbackUndoCmd(getClient))

	return cmd
}

// maybeConfirmRollback renders the rollback preview and asks for
// confirmation when stdin is a TTY and --yes is not set. Returns canceled=true
// when the user declined the prompt so the caller can exit without applying.
func maybeConfirmRollback(
	ctx context.Context,
	cmd *cobra.Command,
	cli snaplib.RollbackAPI,
	store *snaplib.Store,
	manifest *snaplib.Manifest,
	entry *snaplib.ManifestEntry,
	snapID string,
	opts snaplib.RollbackOpts,
	yes bool,
) (canceled bool, err error) {
	if interactive.IsNonInteractive(ctx) || yes {
		return false, nil
	}
	printRollbackHeader(cmd, entry)
	previewOpts := opts
	previewOpts.DryRun = true
	preview, pErr := snaplib.Rollback(ctx, cli, store, manifest, snapID, previewOpts)
	if pErr != nil {
		return false, fmt.Errorf("preview failed: %w", pErr)
	}
	if len(preview.Preview) == 0 {
		return false, nil
	}
	printDiffPreview(cmd, preview)
	prompter := interactive.PrompterFromContext(ctx)
	confirmed, cErr := prompter.Confirm("Apply this rollback?", true)
	if cErr != nil {
		return false, wrapInterrupt(cErr)
	}
	if !confirmed {
		fmt.Fprintln(os.Stdout, "Rollback canceled.")
		return true, nil
	}
	return false, nil
}

// parseEntityIDs parses comma-separated int64 IDs.
func parseEntityIDs(raw string) ([]int64, error) {
	parts := strings.Split(raw, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ID %q: %w", p, err)
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("empty ID list")
	}
	return ids, nil
}

// printRollbackHeader shows server and snapshot context before rollback.
func printRollbackHeader(cmd *cobra.Command, entry *snaplib.ManifestEntry) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "\nServer:    %s\n", serverLabel(entry.ServerURL))
	fmt.Fprintf(out, "Snapshot:  %s (%s %s, T%d)\n\n", entry.ID, entry.Operation, entry.EntityType, entry.RollbackTier)
}

// printDiffPreview renders a diff table to stdout.
func printDiffPreview(cmd *cobra.Command, result *snaplib.RollbackResult) {
	out := cmd.OutOrStdout()
	if result.DryRun {
		fmt.Fprintf(out, "The following changes will be applied:\n")
	}

	if len(result.Preview) == 0 {
		fmt.Fprintln(out, "No differences found — snapshot matches current state.")
		return
	}

	fmt.Fprintln(out, "")
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ENTITY ID\tFIELD\tCURRENT\tSNAPSHOT")
	for _, d := range result.Preview {
		cur := truncate(d.Current, 50)
		sav := truncate(d.Saved, 50)
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", d.EntityID, d.Field, cur, sav)
	}
	_ = w.Flush()
	fmt.Fprintln(out, "")
}

// truncate shortens a string to maxLen, adding "..." if needed.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// printSyncRollbackDetails prints a per-type breakdown plus failure samples for
// sync rollbacks. No-op for non-sync rollbacks.
func printSyncRollbackDetails(out *os.File, result *snaplib.RollbackResult) {
	if result == nil || result.Stats == nil {
		return
	}
	stats := result.Stats
	typeOrder := []string{"case", "section", "shared_step", "suite"}
	hasDetail := false
	for _, t := range typeOrder {
		if ts, ok := stats.ByType[t]; ok && ts.Total > 0 {
			hasDetail = true
			break
		}
	}
	if hasDetail {
		fmt.Fprintln(out, "  Per-type breakdown:")
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "    TYPE\tTOTAL\tDELETED\tSKIPPED\tFAILED\tPRIOR")
		for _, t := range typeOrder {
			ts, ok := stats.ByType[t]
			if !ok || ts.Total == 0 {
				continue
			}
			fmt.Fprintf(w, "    %s\t%d\t%d\t%d\t%d\t%d\n",
				t, ts.Total, ts.Deleted, ts.Skipped, ts.Failed, ts.PreRestored)
		}
		_ = w.Flush()
	}
	if len(stats.Failures) > 0 {
		limit := len(stats.Failures)
		if limit > 5 {
			limit = 5
		}
		fmt.Fprintf(out, "  Failure samples (showing %d of %d):\n", limit, len(stats.Failures))
		for i := 0; i < limit; i++ {
			f := stats.Failures[i]
			fmt.Fprintf(out, "    - %s id=%d: %s\n", f.Type, f.TargetID, f.Error)
		}
	}
}
