package snap

import (
	"fmt"
	"os"

	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

func newManifestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manifest",
		Short: "Inspect and repair the snapshot manifest index",
		Long: `Manage the snapshot manifest (~/.gotr/snaps/manifest.json).

The manifest is the index used by 'snap list', 'snap rollback' and other
subcommands to find snapshots. If the manifest drifts from the on-disk
snapshot directories (manual edits, partial writes, tests writing into
the real home directory, etc.) commands like 'snap rollback <id>' may
fail with "not found in manifest" even though the snapshot data is intact.`,
	}
	cmd.AddCommand(newManifestRepairCmd())
	return cmd
}

func newManifestRepairCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "repair",
		Short: "Reconcile the manifest with on-disk snapshot directories",
		Long: `Scan ~/.gotr/snaps/<category>/<id>/meta.json and reconcile the manifest.

Actions performed (idempotent):
  • re-index snapshot directories whose meta.json is intact but the manifest
    has no entry for them;
  • remove manifest entries whose snapshot directory no longer exists on
    disk (orphan entries).

Snapshot directories with an unreadable meta.json are reported but left
untouched — investigate them manually.

Use --dry-run to preview the plan without modifying the manifest.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := snaplib.NewStore()
			if err != nil {
				return fmt.Errorf("snap manifest repair: %w", err)
			}
			manifest, err := snaplib.LoadManifest(store)
			if err != nil {
				return fmt.Errorf("snap manifest repair: %w", err)
			}

			result, err := snaplib.RepairManifest(store, manifest, dryRun)
			if err != nil {
				return fmt.Errorf("snap manifest repair: %w", err)
			}

			printRepairReport(cmd.OutOrStdout(), result)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show planned changes without modifying the manifest")
	return cmd
}

func printRepairReport(w interface{ Write([]byte) (int, error) }, r *snaplib.RepairResult) {
	mode := "applied"
	if r.DryRun {
		mode = "dry-run"
	}

	if !r.HasChanges() && len(r.MetaErrors) == 0 {
		ui.Successf(os.Stdout, "Manifest is consistent with snapshot directories (%s)", mode)
		return
	}

	fmt.Fprintf(os.Stdout, "Snap manifest repair (%s):\n", mode)

	if len(r.Added) > 0 {
		fmt.Fprintf(os.Stdout, "  Re-indexed %d snapshot(s):\n", len(r.Added))
		for _, a := range r.Added {
			fmt.Fprintf(os.Stdout, "    + %s — %s\n", a.SnapID, a.Reason)
		}
	}
	if len(r.Removed) > 0 {
		fmt.Fprintf(os.Stdout, "  Removed %d orphan entry(ies):\n", len(r.Removed))
		for _, a := range r.Removed {
			fmt.Fprintf(os.Stdout, "    - %s — %s\n", a.SnapID, a.Reason)
		}
	}
	if len(r.MetaErrors) > 0 {
		fmt.Fprintf(os.Stdout, "  Skipped %d directory(ies) with unreadable meta.json:\n", len(r.MetaErrors))
		for _, a := range r.MetaErrors {
			fmt.Fprintf(os.Stdout, "    ? %s — %s\n", a.SnapID, a.Reason)
		}
	}

	if r.DryRun {
		fmt.Fprintln(os.Stdout, "\nNo changes written. Re-run without --dry-run to apply.")
	}
	_ = w // unused; reserved for future test-friendly output redirection
}
