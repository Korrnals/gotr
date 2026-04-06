package sync

import (
	"context"
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/paths"
	"github.com/Korrnals/gotr/internal/ui"

	"github.com/spf13/cobra"
)

var suitesCmd = &cobra.Command{
	Use:   "suites",
	Short: "Миграция suites между проектами",
	Long: `Перенос suites между проектами.

Процесс:
	1) Получение suites (source/target)
	2) Фильтрация дубликатов (по --compare-field)
	3) Подтверждение и импорт
	4) Сохранение mapping (опционально)

Пример:
	gotr sync suites --src-project 30 --dst-project 31 --approve --save-mapping

Флаги:
	--src-project    ID source проекта (обязательный)
	--dst-project    ID destination проекта (обязательный)
	--compare-field  Поле для поиска дубликатов (по умолчанию: title)
	--approve        Автоматическое подтверждение
	--save-mapping   Сохранить mapping
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

		if srcProject == 0 || dstProject == 0 {
			return fmt.Errorf("required IDs: --src-project and --dst-project")
		}

		p := interactive.PrompterFromContext(ctx)

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

		ui.Infof(os.Stdout, "Ready to import: %d new suites", len(filtered))

		if dryRun {
			ui.Info(os.Stdout, "Dry-run: import skipped")
			return nil
		}

		if len(filtered) == 0 {
			ui.Info(os.Stdout, "No new suites")
			return nil
		}

		op.Phase("Awaiting confirmation")
		if !autoApprove {
			ui.Infof(os.Stdout, "Confirm import of %d suites...", len(filtered))
			ok, err := p.Confirm("Continue?", false)
			if err != nil {
				return err
			}
			if !ok {
				ui.Cancelled(os.Stdout)
				return nil
			}
		}

		// Step 3) Confirmation and import
		op.Phase("Importing suites")
		_, err = runSyncStatus(ctx, fmt.Sprintf("Importing %d suites...", len(filtered)), quiet, func(ctx context.Context) (struct{}, error) {
			return struct{}{}, m.ImportSuites(ctx, filtered, false)
		})
		if err != nil {
			return err
		}

		// Step 4) Save mapping if requested
		if autoSaveMapping {
			m.ExportMapping(logDir)
		} else if len(m.Mapping()) > 0 {
			ok, err := p.Confirm("Save mapping?", false)
			if err == nil && ok {
				m.ExportMapping(logDir)
			}
		}

		return nil
	},
}
