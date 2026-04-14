package snap

import (
	"encoding/json"
	"fmt"
	"os"

	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info [snapshot_id]",
		Short: "Show snapshot details",
		Long:  "Displays the full metadata of a snapshot as JSON.\nIf snapshot_id is omitted, shows an interactive picker.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := snaplib.NewStore()
			if err != nil {
				return err
			}

			manifest, err := snaplib.LoadManifest(store)
			if err != nil {
				return err
			}

			snapID, err := resolveSnapshotID(cmd.Context(), args, manifest, "Select snapshot to inspect:")
			if err != nil {
				return err
			}

			meta, err := store.LoadMeta(snapID)
			if err != nil {
				return fmt.Errorf("snapshot %q not found: %w", snapID, err)
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(meta)
		},
	}
}
