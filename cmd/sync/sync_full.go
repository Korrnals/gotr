package sync

import (
	"context"
	"os"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/paths"
	"github.com/Korrnals/gotr/internal/ui"

	"github.com/spf13/cobra"
)

var fullCmd = &cobra.Command{
	Use:   "full",
	Short: "Full migration (shared-steps + cases in one pass)",
	Long: `Performs a full migration: first transfers shared steps (generates mapping), then transfers cases.

Features:
• Automatic interactive selection of projects and suites
• Executes two-stage migration in a single call
• Saves mapping automatically (with --save-mapping)

Examples:
	# Fully interactive mode
	gotr sync full

	# Using flags
	gotr sync full --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --approve --save-mapping
`,

	RunE: func(cmd *cobra.Command, args []string) error {
		cli := getClientInterface(cmd)
		ctx := cmd.Context()

		srcProject, _ := cmd.Flags().GetInt64("src-project")
		srcSuite, _ := cmd.Flags().GetInt64("src-suite")
		dstProject, _ := cmd.Flags().GetInt64("dst-project")
		dstSuite, _ := cmd.Flags().GetInt64("dst-suite")
		compareField, _ := cmd.Flags().GetString("compare-field")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		quiet, _ := cmd.Flags().GetBool("quiet")
		autoApprove, _ := cmd.Flags().GetBool("approve")
		autoSaveMapping, _ := cmd.Flags().GetBool("save-mapping")
		autoSaveFiltered, _ := cmd.Flags().GetBool("save-filtered")

		p := interactive.PrompterFromContext(ctx)
		var err error

		// Interactive source project selection
		if srcProject == 0 {
			srcProject, err = interactive.SelectProject(ctx, p, cli, "Select SOURCE project:")
			if err != nil {
				return err
			}
		}

		// Interactive source suite selection
		if srcSuite == 0 {
			srcSuite, err = interactive.SelectSuiteForProject(ctx, p, cli, srcProject, "Select SOURCE suite:")
			if err != nil {
				return err
			}
		}

		// Interactive destination project selection
		if dstProject == 0 {
			dstProject, err = interactive.SelectProject(ctx, p, cli, "Select DESTINATION project:")
			if err != nil {
				return err
			}
		}

		// Interactive destination suite selection
		if dstSuite == 0 {
			dstSuite, err = interactive.SelectSuiteForProject(ctx, p, cli, dstProject, "Select DESTINATION suite:")
			if err != nil {
				return err
			}
		}

		logDir, err := paths.EnsureLogsDirPath()
		if err != nil {
			return err
		}
		m, err := newMigration(cli, srcProject, srcSuite, dstProject, dstSuite, compareField, logDir)
		if err != nil {
			return err
		}
		defer m.Close()

		op := newSyncOperation("Full migration", quiet)
defer op.Finish()

		// Step 1) Migrate shared steps (Fetch → Filter → Import)
		op.Phase("Step 1/2: shared steps")
		_, err = runSyncStatus(ctx, "Migrating shared steps...", quiet, func(ctx context.Context) (struct{}, error) {
			return struct{}{}, m.MigrateSharedSteps(ctx, dryRun || !autoApprove)
		})
		if err != nil { // if dry-run — no import
			return err
		}

		if dryRun {
			ui.Info(os.Stdout, "Dry-run complete")
			return nil
		}

		// Step 2) Migrate cases (Fetch → Filter → Import)
		op.Phase("Step 2/2: cases")
		_, err = runSyncStatus(ctx, "Migrating cases...", quiet, func(ctx context.Context) (struct{}, error) {
			return struct{}{}, m.MigrateCases(ctx, dryRun)
		})
		if err != nil {
			return err
		}

		if autoSaveMapping {
			_ = m.ExportMapping(logDir)
		}

		if autoSaveFiltered {
			if filtered := m.FilteredSharedSteps(); len(filtered) > 0 {
				if err := m.ExportSharedSteps(filtered, true, logDir); err != nil {
					ui.Warningf(os.Stdout, "Failed to save filtered list: %v", err)
				}
			}
		}

		ui.Success(os.Stdout, "Full migration complete!")
		return nil
	},
}
