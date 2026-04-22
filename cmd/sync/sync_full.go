package sync

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/paths"
	"github.com/Korrnals/gotr/internal/snap"
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
		startedAt := time.Now()

		srcProject, _ := cmd.Flags().GetInt64("src-project")
		srcSuite, _ := cmd.Flags().GetInt64("src-suite")
		dstProject, _ := cmd.Flags().GetInt64("dst-project")
		dstSuite, _ := cmd.Flags().GetInt64("dst-suite")
		compareField, err := resolveMatchField(ctx, cmd, interactive.MatchFieldCases)
		if err != nil {
			return fmt.Errorf("fullCmd.func: %w", err)
		}
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		quiet, _ := cmd.Flags().GetBool("quiet")
		autoApprove, _ := cmd.Flags().GetBool("approve")
		autoSaveMapping, _ := cmd.Flags().GetBool("save-mapping")
		autoSaveFiltered, _ := cmd.Flags().GetBool("save-filtered")
		applySessionFallback(ctx, &srcProject, &dstProject, &srcSuite, &dstSuite)

		p := interactive.PrompterFromContext(ctx)

		// Interactive source project selection
		if srcProject == 0 {
			srcProject, err = interactive.SelectProject(ctx, p, cli, "Select SOURCE project:")
			if err != nil {
				return fmt.Errorf("fullCmd.func: %w", err)
			}
		}

		// Interactive source suite selection
		if srcSuite == 0 {
			srcSuite, err = interactive.SelectSuiteForProject(ctx, p, cli, srcProject, "Select SOURCE suite:")
			if err != nil {
				return fmt.Errorf("fullCmd.func: %w", err)
			}
		}

		// Interactive destination project selection
		if dstProject == 0 {
			dstProject, err = interactive.SelectProject(ctx, p, cli, "Select DESTINATION project:")
			if err != nil {
				return fmt.Errorf("fullCmd.func: %w", err)
			}
		}

		// Interactive destination suite selection
		if dstSuite == 0 {
			dstSuite, err = interactive.SelectSuiteForProject(ctx, p, cli, dstProject, "Select DESTINATION suite:")
			if err != nil {
				return fmt.Errorf("fullCmd.func: %w", err)
			}
		}

		logDir, err := paths.EnsureLogsDirPath()
		if err != nil {
			return fmt.Errorf("fullCmd.func: %w", err)
		}
		m, err := newMigration(cli, srcProject, srcSuite, dstProject, dstSuite, compareField, logDir)
		if err != nil {
			return fmt.Errorf("fullCmd.func: %w", err)
		}
		defer m.Close()

		// Pre-sync snapshot (meta only, data after sync).
		var hook *snap.Hook
		var sd snapshotDecision
		if !dryRun {
			sd = confirmSnapshot(ctx, cmd)
			printPreConfirmSummary(0, "full migration", sd)

			if !autoApprove {
				ok, err := p.Confirm("Continue?", false)
				if err != nil {
					return fmt.Errorf("fullCmd.func: %w", err)
				}
				if !ok {
					ui.Canceled(os.Stdout)
					return nil
				}
			}
		}

		if sd.Create {
			hook = snap.NewHook(cmd)
			meta := snap.BuildMeta(
				snap.OpSyncFull, "sync", nil,
				snap.Tier2, srcProject, srcSuite, snap.ResolveName(cmd), os.Args[1:],
				snap.CurrentServerURL(),
			)
			meta.Label = sd.Label
			hook.Before(ctx, meta, nil)
		} else {
			hook = &snap.Hook{Enabled: false}
		}

		op := newSyncOperation("Full migration", quiet)
		defer op.Finish()

		// Step 1) Migrate shared steps (Fetch → Filter → Import)
		op.Phase("Step 1/2: shared steps")
		_, err = runSyncStatus(ctx, "Migrating shared steps...", quiet, func(ctx context.Context) (struct{}, error) {
			return struct{}{}, m.MigrateSharedSteps(ctx, dryRun)
		})
		if err != nil { // if dry-run — no import
			return fmt.Errorf("fullCmd.func: %w", err)
		}
		sharedFiltered := m.FilteredSharedSteps()
		sharedFilterStats := m.LastFilterStats()
		sharedFailed := m.FailedCount()

		if dryRun {
			ui.Info(os.Stdout, "Dry-run complete")
			return nil
		}

		// Step 2) Migrate cases (Fetch → Filter → Import)
		op.Phase("Step 2/2: cases")
		caseImport, err := runSyncStatus(ctx, "Migrating cases...", quiet, func(ctx context.Context) (struct {
			IDs    []int64
			Errors []string
		}, error) {
			createdIDs, importErrors, cErr := m.MigrateCasesReport(ctx, dryRun)
			if cErr != nil {
				return struct {
					IDs    []int64
					Errors []string
				}{}, cErr
			}
			return struct {
				IDs    []int64
				Errors []string
			}{IDs: createdIDs, Errors: importErrors}, nil
		})
		if err != nil {
			return fmt.Errorf("fullCmd.func: %w", err)
		}
		if len(caseImport.Errors) > 0 {
			ui.Warningf(os.Stdout, "Cases with import errors: %d (see migration log for details)", len(caseImport.Errors))
		}
		caseFilterStats := m.LastFilterStats()
		// Cases-only failures = cumulative failures minus the ones that happened
		// during the shared-steps phase earlier.
		caseFailed := m.FailedCount() - sharedFailed

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

		// Save sync created entities to snapshot for rollback.
		created := make([]snap.SyncCreatedEntity, 0)
		mapping := m.Mapping()
		for _, s := range sharedFiltered {
			if targetID, ok := mapping[s.ID]; ok {
				created = append(created, snap.SyncCreatedEntity{
					Type:     "shared_step",
					SourceID: s.ID,
					TargetID: targetID,
				})
			}
		}
		for _, caseID := range caseImport.IDs {
			created = append(created, snap.SyncCreatedEntity{
				Type:     "case",
				SourceID: 0,
				TargetID: caseID,
			})
		}
		hook.FinalizeSyncData(buildSyncData(created, srcProject, dstProject, srcSuite, dstSuite))

		saveMigrationReport(ctx, cmd, "sync_full", srcProject, dstProject, startedAt, hook, []reportResourceStats{
			filterStatsToReport("shared_steps", sharedFilterStats, int64(len(sharedFiltered)-sharedFailed), int64(sharedFailed)),
			filterStatsToReport("cases", caseFilterStats, int64(len(caseImport.IDs)), int64(max(len(caseImport.Errors), caseFailed))),
		})

		if err := runCoverageGate(ctx, cmd, m, quiet); err != nil {
			return fmt.Errorf("fullCmd.func: %w", err)
		}

		ui.Success(os.Stdout, "Full migration complete!")
		syncPostAction(ctx, cmd, hook, cli)
		return nil
	},
}
