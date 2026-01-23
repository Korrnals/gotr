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
	Long: `Переносит кейсы из source suite в destination suite с заменой shared_step_id.
Требует mapping из sync shared-steps (укажите --mapping-file).
Поддерживает --dry-run и --output.

Пример:
	gotr sync cases --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --mapping-file mapping.json --dry-run

Параметры:
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
		logDir := ".testrail"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("не удалось создать директорию %s: %w", logDir, err)
		}
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		logFile := filepath.Join(logDir, fmt.Sprintf("sync_cases_%s.json", timestamp))
		// Если указан дополнительный файл вывода, используем его
		if outputFile != "" {
			logFile = outputFile
		}

		// Загрузка mapping
		sharedMapping := make(map[int64]int64)
		if mappingFile != "" {
			var err error
			sharedMapping, err = utils.LoadMapping(mappingFile)
			if err != nil {
				return fmt.Errorf("ошибка загрузки mapping: %w", err)
			}
			fmt.Printf("Загружен mapping: %d записей\n", len(sharedMapping))
		} else {
			fmt.Println("Warning: mapping не загружен — shared_step_id НЕ будут заменены")
		}

		// Основной бар
		mainBar := pb.StartNew(6)
		mainBar.SetTemplateString(`{{counters . }} {{bar . }} {{percent . }} | ETA: {{etime . }}`)
		defer mainBar.Finish()

		fmt.Println("Шаг 1/6: Получаем кейсы из source и target...")
		mainBar.Increment()
		sourceCases, err := client.GetCases(srcProject, srcSuite, 0)
		if err != nil {
			return err
		}
		targetCases, err := client.GetCases(dstProject, dstSuite, 0)
		if err != nil {
			return err
		}

		fmt.Println("Шаг 2/6: Проверяем дубликаты...")
		mainBar.Increment()
		targetMap := make(map[string]int64)
		for _, t := range targetCases {
			val := utils.GetFieldValue(t, compareField)
			if val != "" {
				targetMap[val] = t.ID
			}
		}

		fmt.Println("Шаг 3/6: Фильтруем новые кейсы...")
		mainBar.Increment()
		var matches, filtered data.GetCasesResponse
		for _, c := range sourceCases {
			val := utils.GetFieldValue(c, compareField)
			if _, exists := targetMap[val]; exists {
				matches = append(matches, c)
			} else {
				filtered = append(filtered, c)
			}
		}

		fmt.Printf("\nРезультат анализа:\n")
		fmt.Printf("  Совпадения: %d\n", len(matches))
		fmt.Printf("  Новые: %d\n", len(filtered))

		if dryRun {
			fmt.Println("\nDry-run: импорт НЕ выполнен (безопасно).")
			saveLog(logFile, matches, filtered, nil, sharedMapping)
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
			saveLog(logFile, matches, filtered, nil, sharedMapping)
			return nil
		}

		fmt.Println("Шаг 5/6: Импорт кейсов...")
		mainBar.Increment()
		importBar := pb.StartNew(len(filtered))
		importBar.SetTemplateString(`{{counters . }} {{bar . | green}} {{percent . }}`)
		defer importBar.Finish()

		var importErrors []string
		var createdIDs []int64
		for _, c := range filtered {
			newReq := &data.AddCaseRequest{
				Title:                c.Title,
				TypeID:               c.TypeID,
				PriorityID:           c.PriorityID,
				TemplateID:           c.TemplateID,
				MilestoneID:          c.MilestoneID,
				Refs:                 c.Refs,
				CustomPreconds:       c.CustomPreconds,
				CustomStepsSeparated: make([]data.Step, len(c.CustomStepsSeparated)),
			}

			for i, orig := range c.CustomStepsSeparated {
				newStep := data.Step{
					Content:        orig.Content,
					AdditionalInfo: orig.AdditionalInfo,
					Expected:       orig.Expected,
					Refs:           orig.Refs,
					SharedStepID:   orig.SharedStepID,
				}

				if orig.SharedStepID != 0 {
					if newID, exists := sharedMapping[orig.SharedStepID]; exists {
						newStep.SharedStepID = newID
					}
				}

				newReq.CustomStepsSeparated[i] = newStep
			}

			created, err := client.AddCase(dstSuite, newReq)
			if err != nil {
				importErrors = append(importErrors, fmt.Sprintf("кейс %q: %v", c.Title, err))
			} else {
				createdIDs = append(createdIDs, created.ID)
			}

			importBar.Increment()
		}

		mainBar.Increment()
		fmt.Printf("\nИмпорт завершён: %d новых кейсов\n", len(filtered)-len(importErrors))

		if len(importErrors) > 0 {
			fmt.Println("\nОшибки:")
			for _, e := range importErrors {
				fmt.Printf("  - %s\n", e)
			}
		}

		saveLog(logFile, matches, filtered, importErrors, sharedMapping)

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
	syncCasesCmd.Flags().Int64("src-project", 0, "Source project ID")
	syncCasesCmd.Flags().Int64("src-suite", 0, "Source suite ID")
	syncCasesCmd.Flags().Int64("dst-project", 0, "Destination project ID")
	syncCasesCmd.Flags().Int64("dst-suite", 0, "Destination suite ID")
	syncCasesCmd.Flags().String("compare-field", "title", "Поле для дубликатов")
	syncCasesCmd.Flags().Bool("dry-run", false, "Просмотр без импорта")
	syncCasesCmd.Flags().String("output", "", "Дополнительный JSON")
	syncCasesCmd.Flags().String("mapping-file", "", "Mapping файл для замены shared_step_id")
}
