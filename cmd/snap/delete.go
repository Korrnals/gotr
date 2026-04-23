package snap

import (
	"fmt"
	"os"

	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [snapshot_id]",
		Short: "Delete a snapshot",
		Long:  "Removes a snapshot from disk and the manifest index.\nIf snapshot_id is omitted, shows an interactive picker.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := snaplib.NewStore()
			if err != nil {
				return fmt.Errorf("newDeleteCmd.func: %w", err)
			}

			manifest, err := snaplib.LoadManifest(store)
			if err != nil {
				return fmt.Errorf("newDeleteCmd.func: %w", err)
			}

			snapID, err := resolveSnapshotID(cmd.Context(), args, manifest, "Select snapshot to delete:")
			if err != nil {
				return fmt.Errorf("newDeleteCmd.func: %w", err)
			}

			if err := store.Delete(snapID); err != nil {
				return fmt.Errorf("delete snapshot files: %w", err)
			}

			if err := manifest.Remove(snapID); err != nil {
				ui.Warning(os.Stderr, fmt.Sprintf("snap: manifest remove: %v", err))
			}

			ui.Successf(os.Stdout, "Snapshot %s deleted", snapID)
			return nil
		},
	}
}
