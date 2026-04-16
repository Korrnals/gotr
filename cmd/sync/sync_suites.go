package sync

import (
	"context"
	"fmt"
	"os"

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

		srcProject, _ := cmd.Flags().GetInt64("src-project")
		dstProject, _ := cmd.Flags().GetInt64("dst-project")
		compareField, _ := cmd.Flags().GetString("compare-field")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		quiet, _ := cmd.Flags().GetBool("quiet")
		autoApprove, _ := cmd.Flags().GetBool("approve")
		autoSaveMapping, _ := cmd.Flags().GetBool("save-mapping")

		p := interactive.PrompterFromContext(ctx)
		var err error

		// Interactive source project selection
		if srcProject == 0 {
			srcProject, err = interactive.SelectProject(ctx, p, cli, "Select SOURCE project:")
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

		logDir, err := paths.EnsureLogsDirPath()
		if err != nil {
			return err
		}
		m, err := newMigration(cli, srcProject, 0, dstProject, 0, compareField, logDir)
		if err != nil {
			return err
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
			return err
		}
		sourceSuites := loaded.Source
		targetSuites := loaded.Target

		filtered, err := m.FilterSuites(sourceSuites, targetSuites)
		if err != nil {
			return err
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
				return err
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
			snapHook = snap.HookMutation(ctx, snap.Mutation{Cmd: cmd, Op: snap.OpSyncSuites, EntityType: "sync_entity", Tier: snap.Tier2, Label: sd.Label})
		}

		_, err = runSyncStatus(ctx, fmt.Sprintf("Importing %d suites...", len(filtered)), quiet, func(ctx context.Context) (struct{}, error) {
			return struct{}{}, m.ImportSuites(ctx, filtered, false)
		})
		if err != nil {
			return err
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

		syncPostAction(ctx, cmd, snapHook, cli)
		return nil
	},
}
