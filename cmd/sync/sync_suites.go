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

var suitesCmd = &cobra.Command{
	Use:   "suites",
	Short: "Migrate suites between projects",
	Long: `Transfer suites between projects.

Process:
	1) Fetch suites (source/target)
	2) Filter duplicates (by --compare-field)
	3) Confirmation and import
	4) Save mapping (optional)

Example:
	gotr sync suites --src-project 30 --dst-project 31 --approve --save-mapping

Flags:
	--src-project    Source project ID (required)
	--dst-project    Destination project ID (required)
	--compare-field  Field for duplicate detection (default: title)
	--approve        Auto-approve confirmation
	--save-mapping   Save mapping
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cli := getClientInterface(cmd)
		ctx := cmd.Context()
		startedAt := time.Now()

		srcProject, _ := cmd.Flags().GetInt64("src-project")
		dstProject, _ := cmd.Flags().GetInt64("dst-project")
		compareField, err := resolveMatchField(ctx, cmd, interactive.MatchFieldSuites)
		if err != nil {
			return fmt.Errorf("suitesCmd.func: %w", err)
		}
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		quiet, _ := cmd.Flags().GetBool("quiet")
		autoApprove, _ := cmd.Flags().GetBool("approve")
		autoSaveMapping, _ := cmd.Flags().GetBool("save-mapping")
		applySessionFallback(ctx, &srcProject, &dstProject, new(int64), new(int64))

		p := interactive.PrompterFromContext(ctx)

		// Interactive source project selection
		if srcProject == 0 {
			srcProject, err = interactive.SelectProject(ctx, p, cli, "Select SOURCE project:")
			if err != nil {
				return fmt.Errorf("suitesCmd.func: %w", err)
			}
		}

		// Interactive destination project selection
		if dstProject == 0 {
			dstProject, err = interactive.SelectProject(ctx, p, cli, "Select DESTINATION project:")
			if err != nil {
				return fmt.Errorf("suitesCmd.func: %w", err)
			}
		}

		logDir, err := paths.EnsureLogsDirPath()
		if err != nil {
			return fmt.Errorf("suitesCmd.func: %w", err)
		}
		m, err := newMigration(cli, srcProject, 0, dstProject, 0, compareField, logDir)
		if err != nil {
			return fmt.Errorf("suitesCmd.func: %w", err)
		}
		defer m.Close()

		op := newSyncOperation("Sync suites", quiet)
		defer op.Finish()

		op.Phase("Loading suites")
		loaded, err := runSyncStatus(ctx, "Loading suites...", quiet, func(ctx context.Context) (struct {
			Source data.GetSuitesResponse
			Target data.GetSuitesResponse
		}, error) {
			sourceSuites, targetSuites, err := m.FetchSuitesData(ctx)
			if err != nil {
				return struct {
					Source data.GetSuitesResponse
					Target data.GetSuitesResponse
				}{}, err
			}
			return struct {
				Source data.GetSuitesResponse
				Target data.GetSuitesResponse
			}{Source: sourceSuites, Target: targetSuites}, nil
		})
		if err != nil {
			return fmt.Errorf("suitesCmd.func: %w", err)
		}
		sourceSuites := loaded.Source
		targetSuites := loaded.Target

		filtered, err := m.FilterSuites(sourceSuites, targetSuites)
		if err != nil {
			return fmt.Errorf("suitesCmd.func: %w", err)
		}

		printFilterSummary("suites", m.LastFilterStats())

		if dryRun {
			ui.Info(os.Stdout, "Dry-run: import skipped")
			return nil
		}

		if len(filtered) == 0 {
			ui.Info(os.Stdout, "No new suites")
			return nil
		}

		// Snapshot decision + confirmation
		op.Finish() // stop spinner before interactive prompts
		sd := confirmSnapshot(ctx, cmd)
		printPreConfirmSummary(len(filtered), "suites", sd)

		if !autoApprove {
			ok, err := p.Confirm("Continue?", false)
			if err != nil {
				return fmt.Errorf("suitesCmd.func: %w", err)
			}
			if !ok {
				ui.Canceled(os.Stdout)
				return nil
			}
		}

		// Step 3) Import
		op = newSyncOperation("Importing suites", quiet)
		defer op.Finish()

		var snapHook *snap.Hook
		if sd.Create {
			snapHook = snap.HookMutation(ctx, snap.Mutation{Cmd: cmd, Op: snap.OpSyncSuites, EntityType: "sync", Tier: snap.Tier2, Label: sd.Label})
		}

		_, err = runSyncStatus(ctx, fmt.Sprintf("Importing %d suites...", len(filtered)), quiet, func(ctx context.Context) (struct{}, error) {
			return struct{}{}, m.ImportSuites(ctx, filtered, false)
		})
		if err != nil {
			return fmt.Errorf("suitesCmd.func: %w", err)
		}

		// Step 4) Save mapping if requested
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
						Type:     "suite",
						SourceID: s.ID,
						TargetID: targetID,
					})
				}
			}
			snapHook.FinalizeSyncData(buildSyncData(created, srcProject, dstProject, 0, 0))
		}

		saveMigrationReport(ctx, cmd, "sync_suites", srcProject, dstProject, startedAt, snapHook, []reportResourceStats{
			filterStatsToReport("suites", m.LastFilterStats(), int64(len(filtered)), 0),
		})

		syncPostAction(ctx, cmd, snapHook, cli)
		return nil
	},
}
