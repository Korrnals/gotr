package snap

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Korrnals/gotr/internal/interactive"
	snaplib "github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [snapshot_id] [output_path]",
		Short: "Export a snapshot to a JSON file",
		Long: `Exports snapshot metadata and entity data into a single portable JSON file.

The output contains {"meta": {...}, "data": {...}}.
If snapshot_id is omitted, shows an interactive picker.
If output_path is omitted, prompts for filename and directory interactively.`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := snaplib.NewStore()
			if err != nil {
				return fmt.Errorf("newExportCmd.func: %w", err)
			}

			manifest, err := snaplib.LoadManifest(store)
			if err != nil {
				return fmt.Errorf("newExportCmd.func: %w", err)
			}

			snapID, err := resolveSnapshotID(cmd.Context(), args, manifest, "Select snapshot to export:")
			if err != nil {
				return fmt.Errorf("newExportCmd.func: %w", err)
			}

			entry := manifest.Find(snapID)
			if entry == nil {
				return fmt.Errorf("snapshot %q not found in manifest", snapID)
			}

			outPath := ""
			if len(args) >= 2 {
				outPath = args[1]
			}

			// Interactive prompt for output path if not provided.
			if outPath == "" && !interactive.IsNonInteractive(cmd.Context()) {
				outPath, err = promptExportPath(cmd, snapID)
				if err != nil {
					return fmt.Errorf("newExportCmd.func: %w", err)
				}
			}

			if outPath == "" {
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

	return cmd
}

// promptExportPath interactively asks for export filename and directory.
func promptExportPath(cmd *cobra.Command, snapID string) (string, error) {
	p := interactive.PrompterFromContext(cmd.Context())

	defaultName := "snapshot_" + sanitizeFilename(snapID) + ".json"
	name, err := p.Input("Export filename:", defaultName)
	if err != nil {
		return "", wrapInterrupt(err)
	}
	if name == "" {
		name = defaultName
	}

	dir, err := p.Input("Export directory:", ".")
	if err != nil {
		return "", wrapInterrupt(err)
	}
	if dir == "" {
		dir = "."
	}

	return filepath.Join(dir, name), nil
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
