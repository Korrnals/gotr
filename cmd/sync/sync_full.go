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
	Short: "Полная миграция (shared-steps + cases за один проход)",
	Long: `Выполняет полную миграцию: сначала переносит shared steps (формирует mapping), затем переносит cases.

Особенности:
• Автоматический интерактивный выбор проектов и сьютов
• Выполняет двухэтапную миграцию за один вызов
• Сохраняет mapping автоматически (с --save-mapping)

Примеры:
	# Полностью интерактивный режим
	gotr sync full

	# Через флаги
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
		quiet := isQuiet(cmd)
		autoApprove, _ := cmd.Flags().GetBool("approve")
		autoSaveMapping, _ := cmd.Flags().GetBool("save-mapping")

		p := interactive.PrompterFromContext(ctx)
		var err error

		// Интерактивный выбор source проекта
		if srcProject == 0 {
			srcProject, err = interactive.SelectProject(ctx, p, cli, "Select SOURCE project:")
			if err != nil {
				return err
			}
		}

		// Интерактивный выбор source сьюта
		if srcSuite == 0 {
			srcSuite, err = interactive.SelectSuiteForProject(ctx, p, cli, srcProject, "Select SOURCE suite:")
			if err != nil {
				return err
			}
		}

		// Интерактивный выбор destination проекта
		if dstProject == 0 {
			dstProject, err = interactive.SelectProject(ctx, p, cli, "Select DESTINATION project:")
			if err != nil {
				return err
			}
		}

		// Интерактивный выбор destination сьюта
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

		// Шаг 1) Миграция shared steps (Fetch → Filter → Import)
		op.Phase("Step 1/2: shared steps")
		_, err = runSyncStatus(ctx, "Migrating shared steps...", quiet, func(ctx context.Context) (struct{}, error) {
			return struct{}{}, m.MigrateSharedSteps(ctx, dryRun || !autoApprove)
		})
		if err != nil { // если dry-run — без импорта
			return err
		}

		if dryRun {
			ui.Info(os.Stdout, "Dry-run complete")
			return nil
		}

		// Шаг 2) Миграция cases (Fetch → Filter → Import)
		op.Phase("Step 2/2: cases")
		_, err = runSyncStatus(ctx, "Migrating cases...", quiet, func(ctx context.Context) (struct{}, error) {
			return struct{}{}, m.MigrateCases(ctx, dryRun)
		})
		if err != nil {
			return err
		}

		if autoSaveMapping {
			m.ExportMapping(logDir)
		}

		ui.Success(os.Stdout, "Full migration complete!")
		return nil
	},
}
