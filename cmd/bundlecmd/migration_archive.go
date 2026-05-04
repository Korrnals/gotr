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
//
// In multi-snap mode (--all, --label, or multiple positional ids) it emits
// a migration_bundle archive that bundles every selected snap, plus the
// union of matching reports/logs (or, with --all, the entire reports/ and
// logs/ trees) — suitable for transferring full ~/.gotr state to another
// machine in one file.
//
//nolint:gocyclo // Command flow keeps flag parsing, mode resolution and pre-flight checks together.
func newExportMigrationArchiveCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migration-archive [snap_id|label ...]",
		Short: "Export a migration archive (snap + reports + logs) as a portable tar.gz",
		Long: `Export a migration archive to a tar.gz file. Single-snap
archives default to ~/.gotr/exports/snaps/; multi-snap and full-state
bundles default to ~/.gotr/exports/all/.

Default (no arguments, no flags): bundles every snapshot in the local
store plus the ENTIRE ~/.gotr/reports/ and ~/.gotr/logs/ trees into a
single migration_bundle archive. This is the recommended single-file
artifact for transferring a complete ~/.gotr state to another machine.
` + "`gotr import migration-archive <file>`" + ` restores everything in one shot
and auto-detects the archive kind — no flags needed.

Selectors:

  • <snap_id|label>         single-snap mode: bundle just that one snap,
                            its reports, and its ±1h log window
  • <id1> <id2> ...         multi-snap: bundle all listed snapshots and
                            the union of their reports/logs
  • --label <substr>        bundle every snapshot whose label contains
                            <substr> (case-insensitive)
  • --all                   explicit form of the default full-state mode

Archive layout:

  • snaps/<id>/             the full snapshot tree (meta.json + data.json)
  • reports/...             matching report files (or the entire tree
                            with --all)
  • logs/...                matching log files (or the entire tree
                            with --all)`,
		ValidArgsFunction: completeExportSnapArg,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := snap.NewStore()
			if err != nil {
				return fmt.Errorf("export migration-archive: %w", err)
			}
			storeManifest, err := snap.LoadManifest(store)
			if err != nil {
				return fmt.Errorf("export migration-archive: %w", err)
			}
			all, _ := cmd.Flags().GetBool("all")
			label, _ := cmd.Flags().GetString("label")
			outPath, _ := cmd.Flags().GetString("out")
			redact, _ := cmd.Flags().GetBool("redact")
			noReports, _ := cmd.Flags().GetBool("no-reports")
			noLogs, _ := cmd.Flags().GetBool("no-logs")

			// Validate mutually-exclusive selection flags.
			selectorCount := 0
			if all {
				selectorCount++
			}
			if label != "" {
				selectorCount++
			}
			if len(args) > 0 {
				selectorCount++
			}
			if selectorCount > 1 {
				return fmt.Errorf("export migration-archive: --all, --label and positional ids are mutually exclusive")
			}

			// No selector → default to full-state bundle so a bare
			// `gotr export migration-archive` produces one portable
			// archive containing everything.
			if selectorCount == 0 {
				all = true
			}

			// Multi-snap branches.
			if all {
				if outPath == "" {
					p, err := snapbundle.DefaultMigrationBundlePath("all", len(storeManifest.Entries))
					if err != nil {
						return fmt.Errorf("export migration-archive: %w", err)
					}
					outPath = p
				}
				res, err := snapbundle.ExportFull(store, outPath, snapbundle.ExportOptions{
					GotrVersion: version,
					Redact:      redact,
				})
				if err != nil {
					return err
				}
				reportMultiExport(res)
				return nil
			}
			if label != "" {
				ids := selectByLabel(storeManifest, label)
				if len(ids) == 0 {
					return fmt.Errorf("export migration-archive: no snapshots match label substring %q", label)
				}
				if outPath == "" {
					p, err := snapbundle.DefaultMigrationBundlePath("label-"+label, len(ids))
					if err != nil {
						return fmt.Errorf("export migration-archive: %w", err)
					}
					outPath = p
				}
				res, err := snapbundle.ExportMany(store, ids, outPath, snapbundle.ExportOptions{
					GotrVersion:    version,
					Redact:         redact,
					IncludeReports: !noReports,
					IncludeLogs:    !noLogs,
				})
				if err != nil {
					return err
				}
				reportMultiExport(res)
				return nil
			}
			if len(args) > 1 {
				ids, err := resolveSnapIDs(storeManifest, args)
				if err != nil {
					return fmt.Errorf("export migration-archive: %w", err)
				}
				if outPath == "" {
					p, err := snapbundle.DefaultMigrationBundlePath("multi", len(ids))
					if err != nil {
						return fmt.Errorf("export migration-archive: %w", err)
					}
					outPath = p
				}
				res, err := snapbundle.ExportMany(store, ids, outPath, snapbundle.ExportOptions{
					GotrVersion:    version,
					Redact:         redact,
					IncludeReports: !noReports,
					IncludeLogs:    !noLogs,
				})
				if err != nil {
					return err
				}
				reportMultiExport(res)
				return nil
			}

			// Single-snap path (existing behavior).
			target, err := resolveExportSnapTarget(cmd, args)
			if err != nil {
				return err
			}
			entry := storeManifest.Find(target)
			if entry == nil {
				return fmt.Errorf("export migration-archive: snapshot %q not found", target)
			}
			if outPath == "" {
				p, err := defaultMigrationArchivePath(entry.ID, entry.Label)
				if err != nil {
					return fmt.Errorf("export migration-archive: %w", err)
				}
				outPath = p
			}
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
	cmd.Flags().String("out", "", "Destination path (default: ~/.gotr/exports/snaps/migration_<id>_<date>.tar.gz for single-snap; ~/.gotr/exports/all/migration_bundle_*.tar.gz for multi-snap/--all)")
	cmd.Flags().Bool("redact", false, "Strip assignee emails, names, and other sensitive fields from meta.json")
	cmd.Flags().Bool("no-reports", false, "Do not embed migration reports whose filename contains the snap_id (default: embed)")
	cmd.Flags().Bool("no-logs", false, "Do not embed migration_/mapping_/shared_steps_filtered_/sync_cases_ logs (default: embed)")
	cmd.Flags().Bool("all", false, "Bundle ALL snapshots from the local store plus the entire reports/ and logs/ trees into one migration_bundle archive (full-state cross-machine transfer)")
	cmd.Flags().String("label", "", "Bundle every snapshot whose label contains this substring into one migration_bundle archive")
	return cmd
}

// reportMultiExport prints a uniform summary for ExportMany/ExportFull.
func reportMultiExport(res *snapbundle.Result) {
	successf("Exported migration_bundle (%d snap(s), %d file(s)) → %s",
		len(res.SnapIDs), len(res.Files), res.ArchivePath)
	if n := len(res.IncludedReports); n > 0 {
		infof("Embedded %d report(s) under reports/", n)
	}
	if n := len(res.IncludedLogs); n > 0 {
		infof("Embedded %d migration log(s) under logs/", n)
	}
	if len(res.Redacted) > 0 {
		warnf("redacted fields: %s", strings.Join(res.Redacted, ", "))
	}
}

// selectByLabel returns the IDs of every manifest entry whose Label
// contains the given substring (case-insensitive).
func selectByLabel(m *snap.Manifest, substr string) []string {
	needle := strings.ToLower(substr)
	var ids []string
	for _, e := range m.Entries {
		if strings.Contains(strings.ToLower(e.Label), needle) {
			ids = append(ids, e.ID)
		}
	}
	return ids
}

// resolveSnapIDs maps each user-supplied token (snap_id or label) to a
// concrete snap ID via storeManifest.Find. Errors out on unknown tokens.
func resolveSnapIDs(m *snap.Manifest, tokens []string) ([]string, error) {
	ids := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		e := m.Find(tok)
		if e == nil {
			return nil, fmt.Errorf("snapshot %q not found", tok)
		}
		ids = append(ids, e.ID)
	}
	return ids, nil
}

// newImportMigrationArchiveCmd builds `gotr import migration-archive`,
// the inverse of the export command above. It restores the snap, embedded
// reports and embedded logs in a single invocation.
//
//nolint:gocyclo // Command flow keeps interactive picker, auto-detection and import dispatch together.
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
			target, err := resolveImportMigrationArchiveTarget(cmd, args)
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
				if len(res.SnapIDs) > 0 {
					infof("Would import migration_bundle (%d snap(s), %d files) from %s",
						len(res.SnapIDs), len(res.Files), res.ArchivePath)
				} else {
					infof("Would import migration archive %s (%d files) from %s", res.SnapID, len(res.Files), res.ArchivePath)
				}
				return nil
			}
			if len(res.SnapIDs) > 0 {
				successf("Imported migration_bundle (%d snap(s), %d files) from %s",
					len(res.SnapIDs), len(res.Files), res.ArchivePath)
			} else {
				successf("Imported migration archive %s (%d files) from %s", res.SnapID, len(res.Files), res.ArchivePath)
			}
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
			if res.SnapID != "" {
				if err := reindexImported(store, res.SnapID); err != nil {
					warnf("manifest refresh: %v", err)
				}
			}
			for _, id := range res.SnapIDs {
				if err := reindexImported(store, id); err != nil {
					warnf("manifest refresh %s: %v", id, err)
				}
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
