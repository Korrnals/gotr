package snap

import (
	"fmt"
	"os"

	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

func newGCCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gc",
		Short: "Clean up orphaned snapshots from local storage",
		Long: `Removes snapshot directories from ~/.gotr/snaps/ that exist on disk
but are not tracked in the manifest index.

Orphans can appear when:
  • a snapshot was partially saved (e.g. interrupted add/sync)
  • the manifest was manually edited or corrupted
  • a bug left stale directories behind

This command only affects local storage — it does not touch TestRail.
To see what would be cleaned without deleting, use --dry-run.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := snaplib.NewStore()
			if err != nil {
				return err
			}

			manifest, err := snaplib.LoadManifest(store)
			if err != nil {
				return err
			}

			ids := manifest.ManifestIDs()

			// First collect orphans to show details.
			orphans, err := store.CollectOrphans(ids)
			if err != nil {
				return fmt.Errorf("gc failed: %w", err)
			}

			if len(orphans) == 0 {
				fmt.Fprintf(os.Stdout, "No orphaned snapshots found (%d tracked in manifest).\n", len(ids))
				return nil
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")

			// Show what will be (or would be) cleaned.
			verb := "Cleaning"
			if dryRun {
				verb = "Would clean"
			}
			fmt.Fprintf(os.Stdout, "%s %d orphaned snapshot(s) (of %d on disk, %d tracked):\n",
				verb, len(orphans), len(orphans)+len(ids), len(ids))
			for _, id := range orphans {
				fmt.Fprintf(os.Stdout, "  • %s\n", id)
			}

			if dryRun {
				return nil
			}

			// Actually clean.
			cleaned, err := store.CleanOrphans(ids)
			if err != nil {
				return fmt.Errorf("gc failed: %w", err)
			}

			ui.Successf(os.Stdout, "Cleaned %d orphaned snapshot(s)", cleaned)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show orphans without deleting them")
	return cmd
}

