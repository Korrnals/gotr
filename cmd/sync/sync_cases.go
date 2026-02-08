package sync

import (
	"encoding/json"
	"fmt"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/utils"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
)

var casesCmd = &cobra.Command{
	Use:   "cases",
	Short: "Синхронизация тест-кейсов между сюитами",
	Long: `Полная процедура переноса тест-кейсов из одной сюиты в другую.

Особенности:
• Автоматический интерактивный выбор проектов и сьютов (если не указаны флаги)
• Поддержка замены shared_step_id по mapping-файлу
• Интерактивное подтверждение перед импортом
• Dry-run режим (без создания объектов)
• Сохранение JSON-лога результата

Если ID проектов/сьютов не указаны, будет предложено выбрать их из списка.

Примеры:
	# Полностью интерактивный режим (выбор всех параметров)
	gotr sync cases

	# Частично интерактивный (указан только source проект)
	gotr sync cases --src-project 30

	# Полностью через флаги
	gotr sync cases --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859

	# С mapping-файлом и dry-run
	gotr sync cases --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --mapping-file mapping.json --dry-run
`,

	RunE: func(cmd *cobra.Command, args []string) error {
		cli := getClientInterface(cmd)

		srcProject, _ := cmd.Flags().GetInt64("src-project")
		srcSuite, _ := cmd.Flags().GetInt64("src-suite")
		dstProject, _ := cmd.Flags().GetInt64("dst-project")
		dstSuite, _ := cmd.Flags().GetInt64("dst-suite")
		compareField, _ := cmd.Flags().GetString("compare-field")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		outputFile, _ := cmd.Flags().GetString("output")
		mappingFile, _ := cmd.Flags().GetString("mapping-file")

		var err error

		// Интерактивный выбор source проекта
		if srcProject == 0 {
			srcProject, err = selectProjectInteractively(cli, "Выберите SOURCE проект (откуда копировать):")
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
			dstProject, err = selectProjectInteractively(cli, "Выберите DESTINATION проект (куда копировать):")
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

		// Директория для логов
		logDir := utils.LogDir()
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		logFile := filepath.Join(logDir, fmt.Sprintf("sync_cases_%s.json", timestamp))
		// Если указан дополнительный файл вывода, используем его
		if outputFile != "" {
			logFile = outputFile
		}

		// Создаём объект миграции
		m, err := newMigration(cli, srcProject, srcSuite, dstProject, dstSuite, compareField, logDir)
		if err != nil {
			return err
		}
		defer m.Close()

		// Если указан mapping-файл — загрузим в m.mapping
		if mappingFile != "" {
			if err := m.LoadMappingFromFile(mappingFile); err != nil {
				return fmt.Errorf("ошибка загрузки mapping: %w", err)
			}
			fmt.Printf("Загружен mapping: %d записей\n", len(m.Mapping()))
		} else {
			fmt.Println("Warning: mapping не загружен — shared_step_id НЕ будут заменены")
		}

		// Основной бар
		mainBar := pb.StartNew(6)
		mainBar.SetTemplateString(`{{counters . }} {{bar . }} {{percent . }} | ETA: {{etime . }}`)
		defer mainBar.Finish()

		fmt.Println("Шаг 1/6: Получаем кейсы из source и target...")
		mainBar.Increment()
		sourceCases, targetCases, err := m.FetchCasesData()
		if err != nil {
			return err
		}

		fmt.Println("Шаг 2/6: Проверяем дубликаты...")
		mainBar.Increment()

		fmt.Println("Шаг 3/6: Фильтруем новые кейсы...")
		mainBar.Increment()
		filtered, err := m.FilterCases(sourceCases, targetCases)
		if err != nil {
			return err
		}

		// Считаем совпадения (matches)
		var matches data.GetCasesResponse
		filteredIDs := make(map[int64]struct{})
		for _, f := range filtered {
			filteredIDs[f.ID] = struct{}{}
		}
		for _, s := range sourceCases {
			if _, ok := filteredIDs[s.ID]; !ok {
				matches = append(matches, s)
			}
		}

		fmt.Printf("\nРезультат анализа:\n")
		fmt.Printf("  Совпадения: %d\n", len(matches))
		fmt.Printf("  Новые: %d\n", len(filtered))

		if dryRun {
			fmt.Println("\nDry-run: импорт НЕ выполнен (безопасно).")
			saveLog(logFile, matches, filtered, nil, m.Mapping())
			return nil
		}

		fmt.Printf("\nШаг 4/6: Подтверждение импорта %d новых кейсов...\n", len(filtered))
		mainBar.Increment()
		fmt.Print("Продолжить? [y/N]: ")
		var confirm string
		fmt.Scanln(&confirm)
		confirm = strings.ToLower(strings.TrimSpace(confirm))
		if confirm != "y" && confirm != "yes" {
			fmt.Println("Отменено.")
			saveLog(logFile, matches, filtered, nil, m.Mapping())
			return nil
		}

		fmt.Println("Шаг 5/6: Импорт кейсов...")
		mainBar.Increment()

		createdIDs, importErrors, err := m.ImportCasesReport(filtered, false)
		if err != nil {
			return err
		}

		mainBar.Increment()
		fmt.Printf("\nИмпорт завершён: %d новых кейсов\n", len(createdIDs))

		if len(importErrors) > 0 {
			fmt.Println("\nОшибки:")
			for _, e := range importErrors {
				fmt.Printf("  - %s\n", e)
			}
		}

		// Сохраняем лог и mapping
		saveLog(logFile, matches, filtered, importErrors, m.Mapping())

		return nil
	},
}

func saveLog(file string, matches, filtered data.GetCasesResponse, errors []string, mapping map[int64]int64) {
	result := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"matches":   len(matches),
		"filtered":  len(filtered),
		"errors":    errors,
		"mapping":   mapping,
	}
	jsonData, _ := json.MarshalIndent(result, "", "  ")
	os.WriteFile(file, jsonData, 0644)
	fmt.Printf("Лог сохранён: %s\n", file)
}
