package cmd

import (
	"fmt"
	"gotr/internal/utils"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
)

var syncSectionsCmd = &cobra.Command{
	Use:   "sections",
	Short: "Миграция sections между сюитами",
	Long: `Миграция секций между сюитами в пределах проектов.

Процесс:
	1) Получение sections (source/target)
	2) Фильтрация дубликатов
	3) Подтверждение и импорт
	4) Сохранение mapping (опционально)

Пример:
	gotr sync sections --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --approve

Флаги:
	--src-project    ID source проекта
	--src-suite      ID source сюиты
	--dst-project    ID destination проекта
	--dst-suite      ID destination сюиты
	--approve        Автоматическое подтверждение
`,

	RunE: func(cmd *cobra.Command, args []string) error {
		client := GetClient(cmd)

		srcProject, _ := cmd.Flags().GetInt64("src-project")
		srcSuite, _ := cmd.Flags().GetInt64("src-suite")
		dstProject, _ := cmd.Flags().GetInt64("dst-project")
		dstSuite, _ := cmd.Flags().GetInt64("dst-suite")
		compareField, _ := cmd.Flags().GetString("compare-field")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		autoApprove, _ := cmd.Flags().GetBool("approve")
		autoSaveMapping, _ := cmd.Flags().GetBool("save-mapping")

		if srcProject == 0 || srcSuite == 0 || dstProject == 0 || dstSuite == 0 {
			return fmt.Errorf("укажите все обязательные IDs")
		}

		logDir := utils.LogDir()
		m, err := newMigration(client, srcProject, srcSuite, dstProject, dstSuite, compareField, logDir)
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

func init() {
	addSyncFlags(syncSectionsCmd)
	syncCmd.AddCommand(syncSectionsCmd)
}
