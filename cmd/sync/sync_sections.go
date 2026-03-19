package sync

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/paths"
	"github.com/Korrnals/gotr/internal/ui"

	"github.com/spf13/cobra"
)

var sectionsCmd = &cobra.Command{
	Use:   "sections",
	Short: "Миграция sections между сюитами",
	Long: `Миграция секций между сюитами в пределах проектов.

Особенности:
• Автоматический интерактивный выбор проектов и сьютов
• Фильтрация дубликатов по названию
• Подтверждение перед импортом

Примеры:
	# Полностью интерактивный режим
	gotr sync sections

	# Через флаги
	gotr sync sections --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --approve
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
		quiet := isQuiet(cmd)
		autoApprove, _ := cmd.Flags().GetBool("approve")

		var err error
		autoSaveMapping, _ := cmd.Flags().GetBool("save-mapping")

		// Интерактивный выбор source проекта
		if srcProject == 0 {
			srcProject, err = selectProjectInteractively(ctx, cli, "Select SOURCE project:")
			if err != nil {
				return err
			}
		}

		// Интерактивный выбор source сьюта
		if srcSuite == 0 {
			srcSuite, err = selectSuiteInteractively(ctx, cli, srcProject, "Select SOURCE suite:")
			if err != nil {
				return err
			}
		}

		// Интерактивный выбор destination проекта
		if dstProject == 0 {
			dstProject, err = selectProjectInteractively(ctx, cli, "Select DESTINATION project:")
			if err != nil {
				return err
			}
		}

		// Интерактивный выбор destination сьюта
		if dstSuite == 0 {
			dstSuite, err = selectSuiteInteractively(ctx, cli, dstProject, "Select DESTINATION suite:")
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

		op := newSyncOperation("Sync sections", quiet)
		defer op.Finish()

		// Шаг 1) Получение sections из source и target
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
			return err
		}
		sourceSections := loaded.Source
		targetSections := loaded.Target

		// Шаг 2) Фильтрация дубликатов
		filtered, err := m.FilterSections(sourceSections, targetSections)
		if err != nil {
			return err
		}

		ui.Infof(os.Stdout, "Ready to import: %d new sections", len(filtered))

		// Шаг 3) Обработка dry-run
		if dryRun {
			ui.Info(os.Stdout, "Dry-run: import skipped")
			return nil
		}

		if len(filtered) == 0 {
			ui.Info(os.Stdout, "No new sections")
			return nil
		}

		// Шаг 4) Подтверждение и импорт
		op.Phase("Awaiting confirmation")
		if !autoApprove {
			ui.Infof(os.Stdout, "Confirm import of %d sections...", len(filtered))
			fmt.Print("Continue? [y/N]: ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
				ui.Cancelled(os.Stdout)
				return nil
			}
		}

		op.Phase("Importing sections")
		_, err = runSyncStatus(ctx, fmt.Sprintf("Importing %d sections...", len(filtered)), quiet, func(ctx context.Context) (struct{}, error) {
			return struct{}{}, m.ImportSections(ctx, filtered, false)
		})
		if err != nil {
			return err
		}

		// Шаг 5) Сохранение mapping при запросе
		if autoSaveMapping {
			m.ExportMapping(logDir)
		} else if len(m.Mapping()) > 0 {
			fmt.Print("\nSave mapping? [y/N]: ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
				m.ExportMapping(logDir)
			}
		}

		return nil
	},
}
