// Package cleanup provides the `gotr cleanup` top-level command.
//
// It dispatches to per-artifact retention executors: reports (via
// internal/retention), exports (via internal/retention) and snaps (delegated
// to `gotr snap gc` for backwards compatibility).
package cleanup

import (
	"fmt"
	"time"

	"github.com/Korrnals/gotr/internal/paths"
	"github.com/Korrnals/gotr/internal/retention"
	"github.com/spf13/cobra"
)

// Register attaches the cleanup command tree to the given root command.
func Register(root *cobra.Command) {
	root.AddCommand(newCleanupCmd())
}

func newCleanupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Apply retention policies to reports, snapshots and exports",
		Long: `Run the retention executor for gotr artifacts.

Retention is driven by the 'retention:' section of ~/.gotr/config/default.yaml.
All sub-commands accept --dry-run to preview the effect without touching disk.`,
	}
	cmd.AddCommand(newReportsCmd())
	cmd.AddCommand(newExportsCmd())
	cmd.AddCommand(newSnapsCmd())
	cmd.AddCommand(newAllCmd())
	return cmd
}

func newReportsCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "reports",
		Short: "Prune ~/.gotr/reports/ according to retention.reports",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := paths.ReportsDirPath()
			if err != nil {
				return fmt.Errorf("cleanup reports: %w", err)
			}
			policy := retention.LoadReports()
			if dryRun {
				policy.DryRun = true
			}
			if !policy.Enabled {
				fmt.Fprintln(cmd.OutOrStdout(), "retention.reports is disabled (set retention.reports.enabled: true to apply)")
				return nil
			}
			res, err := retention.CleanupReports(dir, policy, time.Now())
			if err != nil {
				return fmt.Errorf("cleanup reports: %w", err)
			}
			fmt.Fprint(cmd.OutOrStdout(), res.Summary())
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview deletions without touching disk")
	return cmd
}

func newExportsCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "exports",
		Short: "Prune ~/.gotr/exports/ according to retention.exports",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := paths.ExportsDirPath()
			if err != nil {
				return fmt.Errorf("cleanup exports: %w", err)
			}
			policy := retention.LoadExports()
			if dryRun {
				policy.DryRun = true
			}
			if !policy.Enabled {
				fmt.Fprintln(cmd.OutOrStdout(), "retention.exports is disabled (set retention.exports.enabled: true to apply)")
				return nil
			}
			res, err := retention.CleanupExports(dir, policy, time.Now())
			if err != nil {
				return fmt.Errorf("cleanup exports: %w", err)
			}
			fmt.Fprint(cmd.OutOrStdout(), res.Summary())
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview deletions without touching disk")
	return cmd
}

func newSnapsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snaps",
		Short: "Delegate to `gotr snap gc` (retention.snaps controls TTL)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			policy := retention.LoadSnaps()
			if !policy.Enabled {
				fmt.Fprintln(cmd.OutOrStdout(),
					"retention.snaps is disabled. Use `gotr snap gc` directly, or set "+
						"retention.snaps.enabled: true and re-run `gotr cleanup snaps`.")
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(),
				"Snap retention is managed by `gotr snap gc`. Run that command with --confirm to apply; "+
					"this subcommand is a placeholder so `gotr cleanup all` can summarize all three artifact families.")
			return nil
		},
	}
	return cmd
}

func newAllCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Run retention for reports and exports (snaps → see `gotr snap gc`)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			now := time.Now()
			out := cmd.OutOrStdout()

			// Reports
			rdir, err := paths.ReportsDirPath()
			if err != nil {
				return fmt.Errorf("cleanup all: reports dir: %w", err)
			}
			rpolicy := retention.LoadReports()
			if dryRun {
				rpolicy.DryRun = true
			}
			if rpolicy.Enabled {
				res, err := retention.CleanupReports(rdir, rpolicy, now)
				if err != nil {
					return fmt.Errorf("cleanup all: reports: %w", err)
				}
				fmt.Fprint(out, "[reports] ", res.Summary())
			} else {
				fmt.Fprintln(out, "[reports] disabled")
			}

			// Exports
			edir, err := paths.ExportsDirPath()
			if err != nil {
				return fmt.Errorf("cleanup all: exports dir: %w", err)
			}
			epolicy := retention.LoadExports()
			if dryRun {
				epolicy.DryRun = true
			}
			if epolicy.Enabled {
				res, err := retention.CleanupExports(edir, epolicy, now)
				if err != nil {
					return fmt.Errorf("cleanup all: exports: %w", err)
				}
				fmt.Fprint(out, "[exports] ", res.Summary())
			} else {
				fmt.Fprintln(out, "[exports] disabled")
			}

			// Snaps — delegate
			spolicy := retention.LoadSnaps()
			if spolicy.Enabled {
				fmt.Fprintln(out, "[snaps]   use `gotr snap gc --confirm` to apply snapshot retention")
			} else {
				fmt.Fprintln(out, "[snaps]   disabled")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview deletions without touching disk")
	return cmd
}
