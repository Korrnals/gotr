package bundlecmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Korrnals/gotr/internal/paths"
	intreport "github.com/Korrnals/gotr/internal/report"
	"github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/snapbundle"
	"github.com/spf13/cobra"
)

// newExportMigrationArchiveCmd builds `gotr export migration-archive`. The
// archive is a tar.gz that combines the snapshot tree, all migration reports
// whose filename references the snap, and the migration JSON / mapping logs
// written within ±1h of the snap. It is intended as a single-file bundle of
// everything produced by a single `gotr sync` run.
func newExportMigrationArchiveCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migration-archive [snap_id|label]",
		Short: "Export a full migration archive (snap + reports + logs) as a portable tar.gz",
		Long: `Export a complete migration archive to a tar.gz file under
~/.gotr/exports/snaps/.

The archive bundles together everything a single migration run produces:

  • snaps/<id>/         the full snapshot tree (meta.json + data.json)
  • reports/...         all markdown/PDF reports whose filename references
                        the snap id (same scope as ` + "`gotr export snap`" + `)
  • logs/...            migration_*.json, mapping_*.json,
                        shared_steps_filtered_*.json and sync_cases_*.json
                        whose mtime is within ~1h of the snap's timestamp

It is the recommended single-file artifact for archival, transfer between
machines, or sharing a reproducible migration with another operator.

Use ` + "`gotr import migration-archive`" + ` to restore everything in one shot.`,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeExportSnapArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := snap.NewStore()
			if err != nil {
				return fmt.Errorf("export migration-archive: %w", err)
			}
			manifest, err := snap.LoadManifest(store)
			if err != nil {
				return fmt.Errorf("export migration-archive: %w", err)
			}
			target, err := resolveExportSnapTarget(cmd, args)
			if err != nil {
				return err
			}
			entry := manifest.Find(target)
			if entry == nil {
				return fmt.Errorf("export migration-archive: snapshot %q not found", target)
			}
			outPath, _ := cmd.Flags().GetString("out")
			if outPath == "" {
				p, err := defaultMigrationArchivePath(entry.ID, entry.Label)
				if err != nil {
					return fmt.Errorf("export migration-archive: %w", err)
				}
				outPath = p
			}
			redact, _ := cmd.Flags().GetBool("redact")
			noReports, _ := cmd.Flags().GetBool("no-reports")
			noLogs, _ := cmd.Flags().GetBool("no-logs")

			res, err := snapbundle.ExportOne(store, entry.ID, outPath, snapbundle.ExportOptions{
				GotrVersion:    version,
				Redact:         redact,
				IncludeReports: !noReports,
				IncludeLogs:    !noLogs,
			})
			if err != nil {
				return err
			}
			successf("Exported migration archive %s → %s", res.SnapID, res.ArchivePath)
			if n := len(res.IncludedReports); n > 0 {
				infof("Embedded %d matching report(s) under reports/", n)
			}
			if n := len(res.IncludedLogs); n > 0 {
				infof("Embedded %d migration log(s) under logs/", n)
			}
			if len(res.Redacted) > 0 {
				warnf("redacted fields: %s", strings.Join(res.Redacted, ", "))
			}
			return nil
		},
	}
	cmd.Flags().String("out", "", "Destination path (default: ~/.gotr/exports/snaps/migration_<id>_<date>.tar.gz)")
	cmd.Flags().Bool("redact", false, "Strip assignee emails, names, and other sensitive fields from meta.json")
	cmd.Flags().Bool("no-reports", false, "Do not embed migration reports whose filename contains the snap_id (default: embed)")
	cmd.Flags().Bool("no-logs", false, "Do not embed migration_/mapping_/shared_steps_filtered_/sync_cases_ logs (default: embed)")
	return cmd
}

// newImportMigrationArchiveCmd builds `gotr import migration-archive`,
// the inverse of the export command above. It restores the snap, embedded
// reports and embedded logs in a single invocation.
func newImportMigrationArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "migration-archive [path.tar.gz]",
		Short:             "Import a full migration archive (snap + reports + logs)",
		Long:              "Import a tar.gz produced by `gotr export migration-archive`. Restores the snapshot, any embedded reports under ~/.gotr/reports/, and any embedded migration logs under ~/.gotr/logs/. Existing files are never overwritten.",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeImportSnapArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := snap.NewStore()
			if err != nil {
				return fmt.Errorf("import migration-archive: %w", err)
			}
			target, err := resolveImportSnapTarget(cmd, args)
			if err != nil {
				return err
			}
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			rename, _ := cmd.Flags().GetString("rename-id")
			dry, _ := cmd.Flags().GetBool("dry-run")
			skipReports, _ := cmd.Flags().GetBool("skip-reports")
			skipLogs, _ := cmd.Flags().GetBool("skip-logs")

			res, err := snapbundle.Import(store, target, snapbundle.ImportOptions{
				Overwrite:   overwrite,
				RenameID:    rename,
				DryRun:      dry,
				SkipReports: skipReports,
				SkipLogs:    skipLogs,
			})
			if err != nil {
				return err
			}
			if dry {
				infof("Would import migration archive %s (%d files) from %s", res.SnapID, len(res.Files), res.ArchivePath)
				return nil
			}
			successf("Imported migration archive %s (%d files) from %s", res.SnapID, len(res.Files), res.ArchivePath)
			if n := len(res.IncludedReports); n > 0 {
				infof("Restored %d bundled report(s) into ~/.gotr/reports/", n)
				if dir, dErr := paths.ReportsDirPath(); dErr == nil {
					if rErr := intreport.Reindex(dir); rErr != nil {
						warnf("reindex reports: %v", rErr)
					}
				}
			}
			if n := len(res.SkippedReports); n > 0 {
				warnf("skipped %d bundled report(s) (already exist on disk): %s",
					n, strings.Join(res.SkippedReports, ", "))
			}
			if n := len(res.IncludedLogs); n > 0 {
				infof("Restored %d bundled log(s) into ~/.gotr/logs/", n)
			}
			if n := len(res.SkippedLogs); n > 0 {
				warnf("skipped %d bundled log(s) (already exist on disk): %s",
					n, strings.Join(res.SkippedLogs, ", "))
			}
			if err := reindexImported(store, res.SnapID); err != nil {
				warnf("manifest refresh: %v", err)
			}
			return nil
		},
	}
	cmd.Flags().Bool("overwrite", false, "Replace an existing snapshot; the old one is moved to ~/.gotr/snaps/.trash/")
	cmd.Flags().String("rename-id", "", "Import under this snapshot id instead of the one recorded in the bundle")
	cmd.Flags().Bool("dry-run", false, "Validate the bundle and print what would be imported; no changes")
	cmd.Flags().Bool("skip-reports", false, "Do not restore bundled reports into ~/.gotr/reports/ (default: restore)")
	cmd.Flags().Bool("skip-logs", false, "Do not restore bundled logs into ~/.gotr/logs/ (default: restore)")
	return cmd
}

// defaultMigrationArchivePath builds the conventional output path under
// ~/.gotr/exports/snaps/ using a `migration_<id>_<ts>.tar.gz` filename.
func defaultMigrationArchivePath(snapID, label string) (string, error) {
	// Reuse snapbundle's destination directory and timestamp shape, but
	// swap the leading `snap_` prefix for `migration_` so the file is
	// recognizable at a glance. We piggyback on DefaultExportPath to
	// avoid duplicating the directory-resolution logic.
	p, err := snapbundle.DefaultExportPath(snapID, label)
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(p)
	base := filepath.Base(p)
	if strings.HasPrefix(base, "snap_") {
		base = "migration_" + strings.TrimPrefix(base, "snap_")
	}
	return filepath.Join(dir, base), nil
}
