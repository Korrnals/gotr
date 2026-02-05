package sync

import (
	"fmt"
	"github.com/Korrnals/gotr/internal/utils"
	"strings"

	"github.com/cheggaaa/pb/v3"
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

		srcProject, _ := cmd.Flags().GetInt64("src-project")
		dstProject, _ := cmd.Flags().GetInt64("dst-project")
		compareField, _ := cmd.Flags().GetString("compare-field")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		autoApprove, _ := cmd.Flags().GetBool("approve")
		autoSaveMapping, _ := cmd.Flags().GetBool("save-mapping")

		if srcProject == 0 || dstProject == 0 {
			return fmt.Errorf("укажите обязательные IDs: --src-project и --dst-project")
		}

		logDir := utils.LogDir()
		m, err := newMigration(cli, srcProject, 0, dstProject, 0, compareField, logDir)
		if err != nil {
			return err
		}
		defer m.Close()

		mainBar := pb.StartNew(3)
		mainBar.SetTemplateString(`{{counters . }} {{bar . }} {{percent . }}`)
		defer mainBar.Finish()

		mainBar.Increment()
		sourceSuites, targetSuites, err := m.FetchSuitesData()
		if err != nil {
			return err
		}

		mainBar.Increment()
		filtered, err := m.FilterSuites(sourceSuites, targetSuites)
		if err != nil {
			return err
		}

		fmt.Printf("\nГотово к импорту: %d новых suites\n", len(filtered))

		if dryRun {
			fmt.Println("Dry-run: импорт не выполнен")
			return nil
		}

		if len(filtered) == 0 {
			fmt.Println("Нет новых suites")
			return nil
		}

		if !autoApprove {
			fmt.Printf("Подтверждение импорта %d suites...\n", len(filtered))
			fmt.Print("Продолжить? [y/N]: ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
				fmt.Println("Отменено")
				return nil
			}
		}

		// Шаг 3) Подтверждение и импорт
		mainBar.Increment()
		if err := m.ImportSuites(filtered, false); err != nil {
			return err
		}

		// Шаг 4) Сохранение mapping при запросе
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
