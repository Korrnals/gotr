package sync

import (
	"fmt"
	"github.com/Korrnals/gotr/internal/progress"
	"github.com/Korrnals/gotr/internal/utils"
	"strings"

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

		srcProject, _ := cmd.Flags().GetInt64("src-project")
		srcSuite, _ := cmd.Flags().GetInt64("src-suite")
		dstProject, _ := cmd.Flags().GetInt64("dst-project")
		compareField, _ := cmd.Flags().GetString("compare-field")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		autoApprove, _ := cmd.Flags().GetBool("approve")
		autoSaveMapping, _ := cmd.Flags().GetBool("save-mapping")
		autoSaveFiltered, _ := cmd.Flags().GetBool("save-filtered")

		var err error

		// Интерактивный выбор source проекта
		if srcProject == 0 {
			srcProject, err = selectProjectInteractively(cli, "Выберите SOURCE проект (откуда копировать shared steps):")
			if err != nil {
				return err
			}
		}

		// Интерактивный выбор source сьюта (опционально, можно 0)
		if srcSuite == 0 {
			// Спрашиваем нужен ли suite
			fmt.Print("\nУказать source suite? [y/N]: ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
				srcSuite, err = selectSuiteInteractively(cli, srcProject, "Выберите SOURCE сьют:")
				if err != nil {
					return err
				}
			}
		}

		// Интерактивный выбор destination проекта
		if dstProject == 0 {
			dstProject, err = selectProjectInteractively(cli, "Выберите DESTINATION проект (куда копировать shared steps):")
			if err != nil {
				return err
			}
		}

		// Директория для логов и инициализация миграции
		logDir := utils.LogDir()
		// Шаг 1) Инициализация объекта миграции (логирование, client, параметры)
		m, err := newMigration(cli, srcProject, srcSuite, dstProject, 0, compareField, logDir)
		if err != nil {
			return err
		}
		defer m.Close()

		// Create progress manager
		pm := progress.NewManager()

		progress.Describe(pm.NewSpinner(""), "Загрузка shared steps...")
		sourceSteps, targetSteps, err := m.FetchSharedStepsData()
		if err != nil {
			return err
		}

		// Шаг 2) Получение кейсов source для определения использования shared steps
		sourceCases, err := m.Client.GetCases(srcProject, srcSuite, 0)
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

		fmt.Printf("\nГотово к импорту: %d новых shared steps\n", len(filtered))

		if dryRun {
			fmt.Println("Dry-run: импорт не выполнен")
			return nil
		}

		if len(filtered) == 0 {
			fmt.Println("Нет новых shared steps")
			return nil
		}

		// Шаг 4) Подтверждение импорта
		if !autoApprove {
			fmt.Printf("Подтверждение импорта %d shared steps...\n", len(filtered))
			fmt.Print("Продолжить? [y/N]: ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
				fmt.Println("Отменено")
				return nil
			}
		}

		// Шаг 5) Импорт
		progress.Describe(pm.NewSpinner(""), fmt.Sprintf("Импорт %d shared steps...", len(filtered)))
		err = m.ImportSharedSteps(filtered, false)
		if err != nil {
			return err
		}

		// Шаг 6) Сохранение mapping/filtered при запросе
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

		if autoSaveFiltered {
			// Сохранение filtered
		} else if len(filtered) > 0 {
			fmt.Print("\nСохранить filtered shared steps? [y/N]: ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
				// Сохранение
			}
		}

		return nil
	},
}
