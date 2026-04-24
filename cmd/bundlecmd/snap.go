package bundlecmd

import (
	"fmt"
	"strings"

	"github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/snapbundle"
	"github.com/spf13/cobra"
)

func newExportSnapCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snap <snap_id|label>",
		Short: "Export a snapshot as a portable tar.gz bundle",
		Long: `Export a single snapshot to a tar.gz archive under ~/.gotr/exports/snaps/.
The archive contains manifest.json (with schema_version, gotr version,
file list with SHA-256), SHA256SUMS, README.txt and the full
~/.gotr/snaps/<id>/ tree.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := snap.NewStore()
			if err != nil {
				return fmt.Errorf("export snap: %w", err)
			}
			manifest, err := snap.LoadManifest(store)
			if err != nil {
				return fmt.Errorf("export snap: %w", err)
			}
			entry := manifest.Find(args[0])
			if entry == nil {
				return fmt.Errorf("export snap: snapshot %q not found", args[0])
			}
			outPath, _ := cmd.Flags().GetString("out")
			if outPath == "" {
				p, err := snapbundle.DefaultExportPath(entry.ID, entry.Label)
				if err != nil {
					return fmt.Errorf("export snap: %w", err)
				}
				outPath = p
			}
			redact, _ := cmd.Flags().GetBool("redact")
			noReports, _ := cmd.Flags().GetBool("no-reports")

			res, err := snapbundle.ExportOne(store, entry.ID, outPath, snapbundle.ExportOptions{
				GotrVersion:    version,
				Redact:         redact,
				IncludeReports: !noReports,
			})
			if err != nil {
				return err
			}
			successf("Exported snapshot %s → %s", res.SnapID, res.ArchivePath)
			if n := len(res.IncludedReports); n > 0 {
				infof("Embedded %d matching report(s) under reports/", n)
			}
			if len(res.Redacted) > 0 {
				warnf("redacted fields: %s", strings.Join(res.Redacted, ", "))
			}
			return nil
		},
	}
	cmd.Flags().String("out", "", "Destination path (default: ~/.gotr/exports/snaps/snap_<id>_<date>.tar.gz)")
	cmd.Flags().Bool("redact", false, "Strip assignee emails, names, and other sensitive fields from meta.json")
	cmd.Flags().Bool("no-reports", false, "Do not embed migration reports whose filename contains the snap_id (default: embed)")
	return cmd
}

func newImportSnapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snap <path.tar.gz>",
		Short: "Import a snapshot bundle into ~/.gotr/snaps/",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := snap.NewStore()
			if err != nil {
				return fmt.Errorf("import snap: %w", err)
			}
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			rename, _ := cmd.Flags().GetString("rename-id")
			dry, _ := cmd.Flags().GetBool("dry-run")

			res, err := snapbundle.Import(store, args[0], snapbundle.ImportOptions{
				Overwrite: overwrite,
				RenameID:  rename,
				DryRun:    dry,
			})
			if err != nil {
				return err
			}
			if dry {
				infof("Would import snapshot %s (%d files) from %s", res.SnapID, len(res.Files), res.ArchivePath)
				return nil
			}
			successf("Imported snapshot %s (%d files) from %s", res.SnapID, len(res.Files), res.ArchivePath)
			// Keep manifest in sync if target was a fresh id.
			if err := reindexImported(store, res.SnapID); err != nil {
				warnf("manifest refresh: %v", err)
			}
			return nil
		},
	}
	cmd.Flags().Bool("overwrite", false, "Replace an existing snapshot; the old one is moved to ~/.gotr/snaps/.trash/")
	cmd.Flags().String("rename-id", "", "Import under this snapshot id instead of the one recorded in the bundle")
	cmd.Flags().Bool("dry-run", false, "Validate the bundle and print what would be imported; no changes")
	return cmd
}

// reindexImported adds a manifest entry for a freshly imported snapshot
// when one is missing. It is best-effort: failures are logged and ignored.
func reindexImported(store *snap.Store, snapID string) error {
	meta, err := store.LoadMeta(snapID)
	if err != nil {
		return err
	}
	manifest, err := snap.LoadManifest(store)
	if err != nil {
		return err
	}
	if manifest.Find(snapID) != nil {
		return nil
	}
	return manifest.Add(meta)
}
