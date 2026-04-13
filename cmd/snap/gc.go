package snap

import (
	"fmt"
	"os"

	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

func newGCCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gc",
		Short: "Clean up orphaned snapshots",
		Long:  "Removes snapshot directories on disk that are not tracked in the manifest index.",
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
			cleaned, err := store.CleanOrphans(ids)
			if err != nil {
				return fmt.Errorf("gc failed: %w", err)
			}

			if cleaned == 0 {
				fmt.Fprintln(os.Stdout, "No orphaned snapshots found.")
			} else {
				ui.Successf(os.Stdout, "Cleaned %d orphaned snapshot(s)", cleaned)
			}
			return nil
		},
	}
}
