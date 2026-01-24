package cmd

import (
	"encoding/json"
	"fmt"
	"gotr/internal/models/data"
	"gotr/internal/utils"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
)

var syncCasesCmd = &cobra.Command{
	Use:   "cases",
	Short: "Синхронизация тест-кейсов между сюитами",
	Long: `Полная процедура переноса тест-кейсов из одной сюиты в другую с поддержкой:
	- замены ` + "`shared_step_id`" + ` по mapping-файлу,
	- интерактивного подтверждения (или --approve),
	- dry-run режима (без создания объектов),
	- сохранения JSON-лога результата.

Процесс команды (внутренние этапы):
	1) Инициализация Migration (логирование, client, параметры)
	2) Загрузка mapping (если указан --mapping-file)
	3) Получение данных (Fetch) из source и target
	4) Фильтрация дубликатов (по --compare-field)
	5) Подтверждение (интерактивно или --approve)
	6) Импорт (параллельно)
	7) Сохранение логов и mapping

Пример:
	gotr sync cases --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --mapping-file mapping.json --dry-run

Флаги:
	--src-project      ID source проекта (обязательный)
	--src-suite        ID source сюиты (обязательный)
	--dst-project      ID destination проекта (обязательный)
	--dst-suite        ID destination сюиты (обязательный)
	--compare-field    Поле для поиска дубликатов (по умолчанию: title)
	--mapping-file     Файл mapping для замены shared_step_id
	--dry-run          Просмотр без импорта (по умолчанию: false)
	--output           Дополнительный JSON файл с результатами
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := GetClient(cmd)

		srcProject, _ := cmd.Flags().GetInt64("src-project")
		srcSuite, _ := cmd.Flags().GetInt64("src-suite")
		dstProject, _ := cmd.Flags().GetInt64("dst-project")
		dstSuite, _ := cmd.Flags().GetInt64("dst-suite")
		compareField, _ := cmd.Flags().GetString("compare-field")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		outputFile, _ := cmd.Flags().GetString("output")
		mappingFile, _ := cmd.Flags().GetString("mapping-file")

		if srcProject == 0 || srcSuite == 0 || dstProject == 0 || dstSuite == 0 {
			return fmt.Errorf("укажите все обязательные IDs")
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
		m, err := newMigration(client, srcProject, srcSuite, dstProject, dstSuite, compareField, logDir)
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

func init() {
	addSyncFlags(syncCasesCmd)
	syncCmd.AddCommand(syncCasesCmd)
}
