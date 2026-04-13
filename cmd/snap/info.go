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
		Use:   "info <snapshot_id>",
		Short: "Show snapshot details",
		Long:  "Displays the full metadata of a snapshot as JSON.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			snapID := args[0]

			store, err := snaplib.NewStore()
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
