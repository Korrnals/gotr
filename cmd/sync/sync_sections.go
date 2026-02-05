package sync

import (
	"fmt"
	"github.com/Korrnals/gotr/internal/utils"
	"strings"

	"github.com/cheggaaa/pb/v3"
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

		srcProject, _ := cmd.Flags().GetInt64("src-project")
		srcSuite, _ := cmd.Flags().GetInt64("src-suite")
		dstProject, _ := cmd.Flags().GetInt64("dst-project")
		dstSuite, _ := cmd.Flags().GetInt64("dst-suite")
		compareField, _ := cmd.Flags().GetString("compare-field")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		autoApprove, _ := cmd.Flags().GetBool("approve")

		var err error
		autoSaveMapping, _ := cmd.Flags().GetBool("save-mapping")

		// Интерактивный выбор source проекта
		if srcProject == 0 {
			srcProject, err = selectProjectInteractively(cli, "Выберите SOURCE проект:")
			if err != nil {
				return err
			}
		}

		// Интерактивный выбор source сьюта
		if srcSuite == 0 {
			srcSuite, err = selectSuiteInteractively(cli, srcProject, "Выберите SOURCE сьют:")
			if err != nil {
				return err
			}
		}

		// Интерактивный выбор destination проекта
		if dstProject == 0 {
			dstProject, err = selectProjectInteractively(cli, "Выберите DESTINATION проект:")
			if err != nil {
				return err
			}
		}

		// Интерактивный выбор destination сьюта
		if dstSuite == 0 {
			dstSuite, err = selectSuiteInteractively(cli, dstProject, "Выберите DESTINATION сьют:")
			if err != nil {
				return err
			}
		}

		logDir := utils.LogDir()
		m, err := newMigration(cli, srcProject, srcSuite, dstProject, dstSuite, compareField, logDir)
		if err != nil {
			return err
		}
		defer m.Close()

		mainBar := pb.StartNew(6)
		mainBar.SetTemplateString(`{{counters . }} {{bar . }} {{percent . }}`)
		defer mainBar.Finish()

		// Шаг 1) Получение sections из source и target
		mainBar.Increment()
		sourceSections, targetSections, err := m.FetchSectionsData()
		if err != nil {
			return err
		}

		// Шаг 2) Фильтрация дубликатов
		mainBar.Increment()
		filtered, err := m.FilterSections(sourceSections, targetSections)
		if err != nil {
			return err
		}

		fmt.Printf("\nГотово к импорту: %d новых sections\n", len(filtered))

		// Шаг 3) Обработка dry-run
		if dryRun {
			fmt.Println("Dry-run: импорт не выполнен")
			return nil
		}

		if len(filtered) == 0 {
			fmt.Println("Нет новых sections")
			return nil
		}

		// Шаг 4) Подтверждение и импорт
		if !autoApprove {
			fmt.Printf("Подтверждение импорта %d sections...\n", len(filtered))
			fmt.Print("Продолжить? [y/N]: ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
				fmt.Println("Отменено")
				return nil
			}
		}

		mainBar.Increment()
		if err := m.ImportSections(filtered, false); err != nil {
			return err
		}

		// Шаг 5) Сохранение mapping при запросе
		mainBar.Increment()
		if autoSaveMapping {
			m.ExportMapping(logDir)
		} else if len(m.Mapping()) > 0 {
			fmt.Print("\nСохранить mapping? [y/N]: ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
				m.ExportMapping(logDir)
			}
		}

		return nil
	},
}
