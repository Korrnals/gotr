package cmd

import (
	"fmt"
	embed "gotr/embedded"
	"gotr/internal/utils"
	"time"

	"github.com/spf13/cobra"
)

// getCmd — основная команда для GET-запросов
var getCmd = &cobra.Command{
    Use:   "get <endpoint> [id]",
    Short: "GET-запрос к TestRail API",
    Long: `Выполняет GET-запрос к указанному эндпоинту TestRail API.

Примеры:
  gotr get get_project 4                          # проект по ID
  gotr get get_project --project-id 4             # то же через флаг
  gotr get get_cases 1 --suite-id 5               # кейсы из сюиты
  gotr get get_cases --project-id 1 --suite-id 5  # всё через флаги
  gotr get get_projects -t table                  # таблица проектов
  gotr get get_project 4 -o project.json          # сохранить в файл
  gotr get get_cases 1 -j --suite-id 5            # через jq + фильтр`,

    Args: cobra.MinimumNArgs(1), // Требует минимум один аргумент — эндпоинт
	RunE: func(cmd *cobra.Command, args []string) error {
        client := GetClient(cmd)

        if len(args) == 0 {
            return fmt.Errorf("укажите эндпоинт")
        }
        endpoint := args[0]

        // Определяем основной ID (project_id или другой)
        var mainID string
        if pid, _ := cmd.Flags().GetString("project-id"); pid != "" {
            mainID = pid // флаг имеет приоритет
        } else if len(args) > 1 {
            mainID = args[1] // позиционный
        }

        // Формируем путь и query
        fullEndpoint, queryParams, err := buildRequestParams(endpoint, mainID, cmd)
        if err != nil {
            return err
        }

        utils.DebugPrint("{getCmd} - Финальный эндпоинт: %s", fullEndpoint)
        utils.DebugPrint("{getCmd} - Query-параметры: %v", queryParams)

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

        quiet, _ := cmd.Flags().GetBool("quiet")
        outputFormat, _ := cmd.Flags().GetString("type")

        // Сохранение
        if saveFile, _ := cmd.Flags().GetString("output"); saveFile != "" {
            if err := client.SaveResponseToFile(data, saveFile, outputFormat); err != nil {
                return fmt.Errorf("ошибка сохранения: %w", err)
            }
            if !quiet {
                fmt.Printf("Ответ сохранён в %s\n", saveFile)
            }
        }

        // jq
        jqEnabled, _ := cmd.Flags().GetBool("jq")
        jqFilter, _ := cmd.Flags().GetString("jq-filter")
        useJq := jqEnabled || jqFilter != ""
        // Подключение встроеной
        if useJq {
            if err := embed.RunEmbeddedJQ(data.RawBody, jqFilter); err != nil {
                return err
            }
            return nil
        }

        // Обычный вывод
        if !quiet {
            client.PrintResponseFromData(data, outputFormat)
        }

        return nil
    },
}

func init() {
	// Формат вывода
	getCmd.Flags().StringP("type", "t", "json", "Формат вывода: json, json-full, table")

	// Явный project_id (альтернатива позиционному)
	getCmd.Flags().StringP("project-id", "p", "", "ID проекта (альтернатива позиционному аргументу)")

	// Query-параметры
	getCmd.Flags().StringP("suite-id", "s", "", "ID тест-сюиты")
	getCmd.Flags().String("section-id", "", "ID секции")
	getCmd.Flags().String("milestone-id", "", "ID milestone")
	getCmd.Flags().String("assignedto-id", "", "ID назначенного пользователя")
	getCmd.Flags().String("status-id", "", "ID статуса")
	getCmd.Flags().String("priority-id", "", "ID приоритета")
	getCmd.Flags().String("type-id", "", "ID типа кейса")

    // Сохранение
    getCmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл (если указан)")

	// Автодополнение
	getCmd.ValidArgs = getValidGetEndpoints()
	getCmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "json-full", "table"}, cobra.ShellCompDirectiveNoFileComp
	})
}