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

var sectionsCmd = &cobra.Command{
	Use:   "sections",
	Short: "Migrate sections between suites",
	Long: `Migrate sections between suites within projects.

Features:
• Automatic interactive selection of projects and suites
• Duplicate filtering by name
• Confirmation before import

Examples:
	# Fully interactive mode
	gotr sync sections

	# Using flags
	gotr sync sections --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --approve
`,

	RunE: func(cmd *cobra.Command, args []string) error {
		cli := getClientInterface(cmd)
		ctx := cmd.Context()
		startedAt := time.Now()

		srcProject, _ := cmd.Flags().GetInt64("src-project")
		srcSuite, _ := cmd.Flags().GetInt64("src-suite")
		dstProject, _ := cmd.Flags().GetInt64("dst-project")
		dstSuite, _ := cmd.Flags().GetInt64("dst-suite")
		compareField, err := resolveMatchField(ctx, cmd, interactive.MatchFieldSections)
		if err != nil {
			return fmt.Errorf("sectionsCmd.func: %w", err)
		}
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		quiet, _ := cmd.Flags().GetBool("quiet")
		autoApprove, _ := cmd.Flags().GetBool("approve")

		autoSaveMapping, _ := cmd.Flags().GetBool("save-mapping")
		applySessionFallback(ctx, &srcProject, &dstProject, &srcSuite, &dstSuite)

		p := interactive.PrompterFromContext(ctx)

		// Interactive source project selection
		if srcProject == 0 {
			srcProject, err = interactive.SelectProject(ctx, p, cli, "Select SOURCE project:")
			if err != nil {
				return fmt.Errorf("sectionsCmd.func: %w", err)
			}
		}

		// Interactive source suite selection
		if srcSuite == 0 {
			srcSuite, err = interactive.SelectSuiteForProject(ctx, p, cli, srcProject, "Select SOURCE suite:")
			if err != nil {
				return fmt.Errorf("sectionsCmd.func: %w", err)
			}
		}

		// Interactive destination project selection
		if dstProject == 0 {
			dstProject, err = interactive.SelectProject(ctx, p, cli, "Select DESTINATION project:")
			if err != nil {
				return fmt.Errorf("sectionsCmd.func: %w", err)
			}
		}

		// Interactive destination suite selection
		if dstSuite == 0 {
			dstSuite, err = interactive.SelectSuiteForProject(ctx, p, cli, dstProject, "Select DESTINATION suite:")
			if err != nil {
				return fmt.Errorf("sectionsCmd.func: %w", err)
			}
		}

		logDir, err := paths.EnsureLogsDirPath()
		if err != nil {
			return fmt.Errorf("sectionsCmd.func: %w", err)
		}
		m, err := newMigration(cli, srcProject, srcSuite, dstProject, dstSuite, compareField, logDir)
		if err != nil {
			return fmt.Errorf("sectionsCmd.func: %w", err)
		}
		defer m.Close()

		op := newSyncOperation("Sync sections", quiet)
		defer op.Finish()

		// Step 1) Fetch sections from source and target
		op.Phase("Loading sections")
		loaded, err := runSyncStatus(ctx, "Loading sections...", quiet, func(ctx context.Context) (struct {
			Source data.GetSectionsResponse
			Target data.GetSectionsResponse
		}, error) {
			sourceSections, targetSections, err := m.FetchSectionsData(ctx)
			if err != nil {
				return struct {
					Source data.GetSectionsResponse
					Target data.GetSectionsResponse
				}{}, err
			}
			return struct {
				Source data.GetSectionsResponse
				Target data.GetSectionsResponse
			}{Source: sourceSections, Target: targetSections}, nil
		})
		if err != nil {
			return fmt.Errorf("sectionsCmd.func: %w", err)
		}
		sourceSections := loaded.Source
		targetSections := loaded.Target

		// Step 2) Filter duplicates
		filtered, err := m.FilterSections(sourceSections, targetSections)
		if err != nil {
			return fmt.Errorf("sectionsCmd.func: %w", err)
		}

		printFilterSummary("sections", m.LastFilterStats())

		// Step 3) Handle dry-run
		if dryRun {
			ui.Info(os.Stdout, "Dry-run: import skipped")
			return nil
		}

		if len(filtered) == 0 {
			ui.Info(os.Stdout, "No new sections")
			return nil
		}

		// Step 4) Snapshot decision + confirmation
		op.Finish() // stop spinner before interactive prompts
		sd := confirmSnapshot(ctx, cmd)
		printPreConfirmSummary(len(filtered), "sections", sd)

		if !autoApprove {
			ok, err := p.Confirm("Continue?", false)
			if err != nil {
				return fmt.Errorf("sectionsCmd.func: %w", err)
			}
			if !ok {
				ui.Canceled(os.Stdout)
				return nil
			}
		}

		op = newSyncOperation("Importing sections", quiet)
		defer op.Finish()

		var snapHook *snap.Hook
		if sd.Create {
			snapHook = snap.HookMutation(ctx, snap.Mutation{Cmd: cmd, Op: snap.OpSyncSections, EntityType: "sync", Tier: snap.Tier2, Label: sd.Label})
		}

		_, err = runSyncStatus(ctx, fmt.Sprintf("Importing %d sections...", len(filtered)), quiet, func(ctx context.Context) (struct{}, error) {
			return struct{}{}, m.ImportSections(ctx, filtered, false)
		})
		if err != nil {
			return fmt.Errorf("sectionsCmd.func: %w", err)
		}

		// Step 5) Save mapping if requested
		if autoSaveMapping {
			_ = m.ExportMapping(logDir)
		} else if len(m.Mapping()) > 0 {
			ok, err := p.Confirm("Save mapping?", false)
			if err == nil && ok {
				_ = m.ExportMapping(logDir)
			}
		}

		if snapHook != nil && snapHook.Enabled {
			created := make([]snap.SyncCreatedEntity, 0, len(filtered))
			mapping := m.Mapping()
			for _, s := range filtered {
				if targetID, ok := mapping[s.ID]; ok {
					created = append(created, snap.SyncCreatedEntity{
						Type:     "section",
						SourceID: s.ID,
						TargetID: targetID,
					})
				}
			}
			snapHook.FinalizeSyncData(buildSyncData(created, srcProject, dstProject, srcSuite, dstSuite))
		}

		saveMigrationReport(ctx, cmd, "sync_sections", srcProject, dstProject, startedAt, snapHook, []reportResourceStats{
			filterStatsToReport("sections", m.LastFilterStats(), int64(len(filtered)), 0),
		})

		syncPostAction(ctx, cmd, snapHook, cli)
		return nil
	},
}
