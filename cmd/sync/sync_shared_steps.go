package sync

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/paths"
	"github.com/Korrnals/gotr/internal/snap"
	"github.com/Korrnals/gotr/internal/ui"

	"github.com/spf13/cobra"
)

var sharedStepsCmd = &cobra.Command{
	Use:   "shared-steps",
	Short: "Migrate shared steps",
	Long: `Transfer shared steps from source project to destination project.

Features:
• Automatic interactive selection of projects and suites (if flags are not specified)
• Generates mapping for shared_step_id replacement during case migration
• Confirmation before import

Examples:
	# Fully interactive mode
	gotr sync shared-steps

	# Partially interactive
	gotr sync shared-steps --src-project 30

	# Fully via flags
	gotr sync shared-steps --src-project 30 --src-suite 20069 --dst-project 31 --approve --save-mapping
`,

	RunE: func(cmd *cobra.Command, args []string) error {
		cli := getClientInterface(cmd)
		ctx := cmd.Context()
		startedAt := time.Now()

		srcProject, _ := cmd.Flags().GetInt64("src-project")
		srcSuite, _ := cmd.Flags().GetInt64("src-suite")
		dstProject, _ := cmd.Flags().GetInt64("dst-project")
		compareField, _ := cmd.Flags().GetString("compare-field")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		autoApprove, _ := cmd.Flags().GetBool("approve")
		quiet, _ := cmd.Flags().GetBool("quiet")
		autoSaveMapping, _ := cmd.Flags().GetBool("save-mapping")
		autoSaveFiltered, _ := cmd.Flags().GetBool("save-filtered")
		applySessionFallback(ctx, &srcProject, &dstProject, &srcSuite, new(int64))

		p := interactive.PrompterFromContext(ctx)
		var err error

		// Interactive source project selection
		if srcProject == 0 {
			srcProject, err = interactive.SelectProject(ctx, p, cli, "Select SOURCE project (copy shared steps from):")
			if err != nil {
				return fmt.Errorf("sharedStepsCmd.func: %w", err)
			}
		}

		// Interactive source suite selection (optional, can be 0)
		if srcSuite == 0 {
			// Ask if suite is needed
			specifySuite, err := p.Confirm("Specify source suite?", false)
			if err != nil {
				return fmt.Errorf("sharedStepsCmd.func: %w", err)
			}
			if specifySuite {
				srcSuite, err = interactive.SelectSuiteForProject(ctx, p, cli, srcProject, "Select SOURCE suite:")
				if err != nil {
					return fmt.Errorf("sharedStepsCmd.func: %w", err)
				}
			}
		}

		// Interactive destination project selection
		if dstProject == 0 {
			dstProject, err = interactive.SelectProject(ctx, p, cli, "Select DESTINATION project (copy shared steps to):")
			if err != nil {
				return fmt.Errorf("sharedStepsCmd.func: %w", err)
			}
		}

		// Log directory and migration initialization
		logDir, err := paths.EnsureLogsDirPath()
		if err != nil {
			return fmt.Errorf("sharedStepsCmd.func: %w", err)
		}
		// Step 1) Initialize migration object (logging, client, parameters)
		m, err := newMigration(cli, srcProject, srcSuite, dstProject, 0, compareField, logDir)
		if err != nil {
			return fmt.Errorf("sharedStepsCmd.func: %w", err)
		}
		defer m.Close()

		op := newSyncOperation("Sync shared steps", quiet)
		defer op.Finish()

		op.Phase("Loading shared steps")
		loadedSteps, err := runSyncStatus(ctx, "Loading shared steps...", quiet, func(ctx context.Context) (struct {
			Source data.GetSharedStepsResponse
			Target data.GetSharedStepsResponse
		}, error) {
			sourceSteps, targetSteps, err := m.FetchSharedStepsData(ctx)
			if err != nil {
				return struct {
					Source data.GetSharedStepsResponse
					Target data.GetSharedStepsResponse
				}{}, err
			}
			return struct {
				Source data.GetSharedStepsResponse
				Target data.GetSharedStepsResponse
			}{Source: sourceSteps, Target: targetSteps}, nil
		})
		if err != nil {
			return fmt.Errorf("sharedStepsCmd.func: %w", err)
		}
		sourceSteps := loadedSteps.Source
		targetSteps := loadedSteps.Target

		// Step 2) Fetch source cases to determine shared steps usage.
		// When srcSuite==0 (user opted out of suite selection), load cases
		// from every suite so filtering works for multi-suite projects.
		op.Phase("Loading source cases")
		var sourceCases data.GetCasesResponse
		if srcSuite != 0 {
			sourceCases, err = runSyncStatus(ctx, "Loading source cases...", quiet, func(ctx context.Context) (data.GetCasesResponse, error) {
				return m.Client.GetCases(ctx, srcProject, srcSuite, 0)
			})
			if err != nil {
				return fmt.Errorf("sharedStepsCmd.func: %w", err)
			}
		} else {
			suites, sErr := m.Client.GetSuites(ctx, srcProject)
			if sErr != nil {
				return fmt.Errorf("failed to load suites for project %d: %w", srcProject, sErr)
			}
			for _, s := range suites {
				cases, cErr := runSyncStatus(ctx, fmt.Sprintf("Loading cases from suite %d...", s.ID), quiet, func(ctx context.Context) (data.GetCasesResponse, error) {
					return m.Client.GetCases(ctx, srcProject, s.ID, 0)
				})
				if cErr != nil {
					return fmt.Errorf("failed to load cases for suite %d: %w", s.ID, cErr)
				}
				sourceCases = append(sourceCases, cases...)
			}
		}
		usedStepIDs := make(map[int64]struct{})
		for _, c := range sourceCases {
			for _, step := range c.CustomStepsSeparated {
				if step.SharedStepID != 0 {
					usedStepIDs[step.SharedStepID] = struct{}{}
				}
			}
		}

		// Step 3) Filter candidates (exclude used and duplicates)
		filtered, err := m.FilterSharedSteps(sourceSteps, targetSteps, usedStepIDs)
		if err != nil {
			return fmt.Errorf("sharedStepsCmd.func: %w", err)
		}

		printFilterSummary("shared steps", m.LastFilterStats())

		if dryRun {
			ui.Info(os.Stdout, "Dry-run: import skipped")
			return nil
		}

		if len(filtered) == 0 {
			ui.Info(os.Stdout, "No new shared steps")
			return nil
		}

		// Snapshot decision + confirmation
		op.Finish() // stop spinner before interactive prompts
		sd := confirmSnapshot(ctx, cmd)
		printPreConfirmSummary(len(filtered), "shared steps", sd)

		// Step 4) Confirm import
		if !autoApprove {
			ok, err := p.Confirm("Continue?", false)
			if err != nil {
				return fmt.Errorf("sharedStepsCmd.func: %w", err)
			}
			if !ok {
				ui.Canceled(os.Stdout)
				return nil
			}
		}

		// Step 5) Import
		op = newSyncOperation("Importing shared steps", quiet)
		defer op.Finish()

		var snapHook *snap.Hook
		if sd.Create {
			snapHook = snap.HookMutation(ctx, snap.Mutation{Cmd: cmd, Op: snap.OpSyncSharedSteps, EntityType: "sync", Tier: snap.Tier2, Label: sd.Label})
		}

		_, err = runSyncStatus(ctx, fmt.Sprintf("Importing %d shared steps...", len(filtered)), quiet, func(ctx context.Context) (struct{}, error) {
			return struct{}{}, m.ImportSharedSteps(ctx, filtered, false)
		})
		if err != nil {
			return fmt.Errorf("sharedStepsCmd.func: %w", err)
		}

		// Step 6) Save mapping/filtered if requested
		if autoSaveMapping {
			_ = m.ExportMapping(logDir)
		} else if len(m.Mapping()) > 0 {
			ok, err := p.Confirm("Save mapping?", false)
			if err == nil && ok {
				_ = m.ExportMapping(logDir)
			}
		}

		// Step 7) Save filtered shared steps list if requested
		if autoSaveFiltered {
			if err := m.ExportSharedSteps(filtered, true, logDir); err != nil {
				ui.Warningf(os.Stdout, "Failed to save filtered list: %v", err)
			}
		} else if len(filtered) > 0 {
			ok, err := p.Confirm("Save filtered shared steps list?", false)
			if err == nil && ok {
				if err := m.ExportSharedSteps(filtered, true, logDir); err != nil {
					ui.Warningf(os.Stdout, "Failed to save filtered list: %v", err)
				}
			}
		}

		if snapHook != nil && snapHook.Enabled {
			created := make([]snap.SyncCreatedEntity, 0, len(filtered))
			mapping := m.Mapping()
			for _, s := range filtered {
				if targetID, ok := mapping[s.ID]; ok {
					created = append(created, snap.SyncCreatedEntity{
						Type:     "shared_step",
						SourceID: s.ID,
						TargetID: targetID,
					})
				}
			}
			snapHook.FinalizeSyncData(buildSyncData(created, srcProject, dstProject, srcSuite, 0))
		}

		saveMigrationReport(ctx, cmd, "sync_shared_steps", srcProject, dstProject, startedAt, snapHook, []reportResourceStats{
			filterStatsToReport("shared_steps", m.LastFilterStats(), int64(len(filtered)), 0),
		})

		syncPostAction(ctx, cmd, snapHook, cli)
		return nil
	},
}
