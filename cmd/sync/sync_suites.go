package sync

import (
	"fmt"
	"os"
	"strings"

	"github.com/Korrnals/gotr/internal/paths"
	"github.com/Korrnals/gotr/internal/progress"
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
		autoApprove, _ := cmd.Flags().GetBool("approve")
		autoSaveMapping, _ := cmd.Flags().GetBool("save-mapping")

		if srcProject == 0 || dstProject == 0 {
			return fmt.Errorf("required IDs: --src-project and --dst-project")
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

		// Create progress manager
		pm := progress.NewManager()

		progress.Describe(pm.NewSpinner(""), "Загрузка suites...")
		sourceSuites, targetSuites, err := m.FetchSuitesData(ctx)
		if err != nil {
			return err
		}

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

		if !autoApprove {
			ui.Infof(os.Stdout, "Confirm import of %d suites...", len(filtered))
			fmt.Print("Continue? [y/N]: ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
				ui.Cancelled(os.Stdout)
				return nil
			}
		}

		// Шаг 3) Подтверждение и импорт
		progress.Describe(pm.NewSpinner(""), fmt.Sprintf("Импорт %d suites...", len(filtered)))
		if err := m.ImportSuites(ctx, filtered, false); err != nil {
			return err
		}

		// Шаг 4) Сохранение mapping при запросе
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
