package snap

import (
	"fmt"
	"os"

	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <snapshot_id>",
		Short: "Export a snapshot to a JSON file",
		Long: `Exports snapshot metadata and entity data into a single portable JSON file.

The output contains {"meta": {...}, "data": {...}}.
Use -o/--output to specify the destination path (default: snapshot_<id>.json).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			snapID := args[0]

			store, err := snaplib.NewStore()
			if err != nil {
				return err
			}

			manifest, err := snaplib.LoadManifest(store)
			if err != nil {
				return err
			}

			entry := manifest.Find(snapID)
			if entry == nil {
				return fmt.Errorf("snapshot %q not found in manifest", snapID)
			}

			outPath, _ := cmd.Flags().GetString("output")
			if outPath == "" {
				// Default filename: replace "/" in snapID with "_"
				safe := sanitizeFilename(snapID)
				outPath = "snapshot_" + safe + ".json"
			}

			if err := store.Export(snapID, outPath); err != nil {
				return fmt.Errorf("export failed: %w", err)
			}

			ui.Successf(os.Stdout, "Exported snapshot %s → %s", snapID, outPath)
			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", "Output file path (default: snapshot_<id>.json)")

	return cmd
}

// sanitizeFilename replaces path separators with underscores.
func sanitizeFilename(id string) string {
	out := make([]byte, len(id))
	for i := range id {
		if id[i] == '/' || id[i] == '\\' {
			out[i] = '_'
		} else {
			out[i] = id[i]
		}
	}
	return string(out)
}
