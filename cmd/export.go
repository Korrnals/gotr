package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/Korrnals/gotr/internal/utils"
	"github.com/spf13/cobra"
)

// exportCmd — команда для экспорта данных
var exportCmd = &cobra.Command{
	Use:   "export <resource> <endpoint> [id]",
	Short: "Экспорт данных из TestRail в JSON-файл",
	Long: `Экспортирует данные из TestRail в JSON-файл.

Имя файла для сохранения:
    • Через флаг --output (-o): gotr export cases get_cases 30 -o my_cases.json
    • Без флага: сохраняется в директорию .testrail с именем <resource>_[id]_<timestamp>.json

Пример:
    gotr export projects get_projects
    gotr export cases get_cases 1 --suite-id 5 -o cases_suite5.json`,

	Args: cobra.MinimumNArgs(2), // resource и endpoint обязательны
	RunE: func(cmd *cobra.Command, args []string) error {
		client := GetClient(cmd)

		resource := args[0]
		endpoint := args[1]

		// Определяем основной ID
		var mainID string
		if pid, _ := cmd.Flags().GetString("project-id"); pid != "" {
			mainID = pid
		} else if len(args) > 2 {
			mainID = args[2]
		}

		// Формируем путь и query-параметры — одна функция делает всё
		fullEndpoint, queryParams, err := buildRequestParams(endpoint, mainID, cmd)
		if err != nil {
			return err
		}

		utils.DebugPrint("{exportCmd} - Финальный эндпоинт: %s", fullEndpoint)
		utils.DebugPrint("{exportCmd} - Query-параметры: %v", queryParams)

		// Запрос
		start := time.Now()
		resp, err := client.Get(fullEndpoint, queryParams)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		data, err := client.ReadResponse(resp, time.Since(start), "json")
		if err != nil {
			return fmt.Errorf("response reading error: %w", err)
		}

		// Флаги
		quiet, _ := cmd.Flags().GetBool("quiet")
		outputFile, _ := cmd.Flags().GetString("save")

		// Имя файла
		filename := outputFile
		if filename == "" {
			exportDir := ".testrail"
			// Создаём директорию (MkdirAll — создаёт вложенные и игнорирует "exists")
			if err := os.MkdirAll(exportDir, 0755); err != nil {
				return fmt.Errorf("не удалось создать директорию %s: %w", exportDir, err)
			}
			filename = fmt.Sprintf("%s/%s_%s.json", exportDir, resource, time.Now().Format("20060102_150405"))
			if mainID != "" {
				filename = fmt.Sprintf("%s/%s_%s_%s.json", exportDir, resource, mainID, time.Now().Format("20060102_150405"))
			}
		}

		// Сохранение
		if err := client.SaveResponseToFile(data, filename, "json"); err != nil {
			return fmt.Errorf("ошибка экспорта в файл %s: %w", filename, err)
		}

		if !quiet {
			fmt.Printf("Данные экспортированы в %s\n", filename)
		}

		return nil
	},
}
