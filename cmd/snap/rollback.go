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

Partial failures are resumable: re-run the same rollback to retry failed entities.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()

			store, err := snaplib.NewStore()
			if err != nil {
				return err
			}

			manifest, err := snaplib.LoadManifest(store)
			if err != nil {
				return err
			}

			snapID, err := resolveSnapshotID(ctx, args, manifest, "Select snapshot to rollback:")
			if err != nil {
				return err
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
				printDiffPreview(result)
				return nil
			}

			// Interactive mode: show preview + confirm.
			prompter := interactive.PrompterFromContext(ctx)
			if !interactive.IsNonInteractive(ctx) {
				previewOpts := opts
				previewOpts.DryRun = true
				preview, err := snaplib.Rollback(ctx, cli, store, manifest, snapID, previewOpts)
				if err != nil {
					return fmt.Errorf("preview failed: %w", err)
				}

				if len(preview.Preview) > 0 {
					printDiffPreview(preview)
					confirmed, err := prompter.Confirm("Apply this rollback?", true)
					if err != nil || !confirmed {
						fmt.Fprintln(os.Stdout, "Rollback cancelled.")
						return nil
					}
				}
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
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview changes without applying them")
	cmd.Flags().String("entity-ids", "", "Limit rollback to specific entity IDs (comma-separated)")

	return cmd
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

// printDiffPreview renders a diff table to stdout.
func printDiffPreview(result *snaplib.RollbackResult) {
	if result.DryRun {
		fmt.Fprintf(os.Stdout, "Dry-run preview: %s\n", result.Message)
	}

	if len(result.Preview) == 0 {
		fmt.Fprintln(os.Stdout, "No differences found — snapshot matches current state.")
		return
	}

	fmt.Fprintln(os.Stdout, "")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ENTITY ID\tFIELD\tCURRENT\tSNAPSHOT")
	for _, d := range result.Preview {
		cur := truncate(d.Current, 50)
		sav := truncate(d.Saved, 50)
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", d.EntityID, d.Field, cur, sav)
	}
	_ = w.Flush()
	fmt.Fprintln(os.Stdout, "")
}

// truncate shortens a string to maxLen, adding "..." if needed.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
