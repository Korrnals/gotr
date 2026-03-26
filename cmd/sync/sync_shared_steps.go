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

var sharedStepsCmd = &cobra.Command{
	Use:   "shared-steps",
	Short: "Миграция общих шагов (shared steps)",
	Long: `Перенос общих шагов (shared steps) из source проекта в destination проект.

Особенности:
• Автоматический интерактивный выбор проектов и сьютов (если не указаны флаги)
• Генерация mapping для замены shared_step_id при миграции кейсов
• Подтверждение перед импортом

Примеры:
	# Полностью интерактивный режим
	gotr sync shared-steps

	# Частично интерактивный
	gotr sync shared-steps --src-project 30

	# Полностью через флаги
	gotr sync shared-steps --src-project 30 --src-suite 20069 --dst-project 31 --approve --save-mapping
`,

	RunE: func(cmd *cobra.Command, args []string) error {
		cli := getClientInterface(cmd)
		ctx := cmd.Context()

		srcProject, _ := cmd.Flags().GetInt64("src-project")
		srcSuite, _ := cmd.Flags().GetInt64("src-suite")
		dstProject, _ := cmd.Flags().GetInt64("dst-project")
		compareField, _ := cmd.Flags().GetString("compare-field")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		autoApprove, _ := cmd.Flags().GetBool("approve")
		quiet := isQuiet(cmd)
		autoSaveMapping, _ := cmd.Flags().GetBool("save-mapping")
		autoSaveFiltered, _ := cmd.Flags().GetBool("save-filtered")

		p := interactive.PrompterFromContext(ctx)
		var err error

		// Интерактивный выбор source проекта
		if srcProject == 0 {
			srcProject, err = interactive.SelectProject(ctx, p, cli, "Select SOURCE project (copy shared steps from):")
			if err != nil {
				return err
			}
		}

		// Интерактивный выбор source сьюта (опционально, можно 0)
		if srcSuite == 0 {
			// Спрашиваем нужен ли suite
			specifySuite, err := p.Confirm("Specify source suite?", false)
			if err != nil {
				return err
			}
			if specifySuite {
				srcSuite, err = interactive.SelectSuiteForProject(ctx, p, cli, srcProject, "Select SOURCE suite:")
				if err != nil {
					return err
				}
			}
		}

		// Интерактивный выбор destination проекта
		if dstProject == 0 {
			dstProject, err = interactive.SelectProject(ctx, p, cli, "Select DESTINATION project (copy shared steps to):")
			if err != nil {
				return err
			}
		}

		// Директория для логов и инициализация миграции
		logDir, err := paths.EnsureLogsDirPath()
		if err != nil {
			return err
		}
		// Шаг 1) Инициализация объекта миграции (логирование, client, параметры)
		m, err := newMigration(cli, srcProject, srcSuite, dstProject, 0, compareField, logDir)
		if err != nil {
			return err
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
			return err
		}
		sourceSteps := loadedSteps.Source
		targetSteps := loadedSteps.Target

		// Шаг 2) Получение кейсов source для определения использования shared steps
		op.Phase("Loading source cases")
		sourceCases, err := runSyncStatus(ctx, "Loading source cases...", quiet, func(ctx context.Context) (data.GetCasesResponse, error) {
			return m.Client.GetCases(ctx, srcProject, srcSuite, 0)
		})
		if err != nil {
			return err
		}
		caseIDsSet := make(map[int64]struct{})
		for _, c := range sourceCases {
			caseIDsSet[c.ID] = struct{}{}
		}

		// Шаг 3) Фильтрация кандидатов (исключаем используемые и дубликаты)
		filtered, err := m.FilterSharedSteps(sourceSteps, targetSteps, caseIDsSet)
		if err != nil {
			return err
		}

		ui.Infof(os.Stdout, "Ready to import: %d new shared steps", len(filtered))

		if dryRun {
			ui.Info(os.Stdout, "Dry-run: import skipped")
			return nil
		}

		if len(filtered) == 0 {
			ui.Info(os.Stdout, "No new shared steps")
			return nil
		}

		// Шаг 4) Confirm import of
		op.Phase("Awaiting confirmation")
		if !autoApprove {
			ui.Infof(os.Stdout, "Confirm import of %d shared steps...", len(filtered))
			ok, err := p.Confirm("Continue?", false)
			if err != nil {
				return err
			}
			if !ok {
				ui.Cancelled(os.Stdout)
				return nil
			}
		}

		// Шаг 5) Импорт
		op.Phase("Importing shared steps")
		_, err = runSyncStatus(ctx, fmt.Sprintf("Importing %d shared steps...", len(filtered)), quiet, func(ctx context.Context) (struct{}, error) {
			return struct{}{}, m.ImportSharedSteps(ctx, filtered, false)
		})
		if err != nil {
			return err
		}

		// Шаг 6) Сохранение mapping/filtered при запросе
		if autoSaveMapping {
			m.ExportMapping(logDir)
		} else if len(m.Mapping()) > 0 {
			ok, err := p.Confirm("Save mapping?", false)
			if err == nil && ok {
				m.ExportMapping(logDir)
			}
		}

		if autoSaveFiltered {
			// Сохранение filtered
		} else if len(filtered) > 0 {
			ok, err := p.Confirm("Save filtered shared steps?", false)
			if err == nil && ok {
				// Сохранение
			}
		}

		return nil
	},
}
