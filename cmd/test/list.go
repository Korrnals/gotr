package test

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/service"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду для получения списка тестов
func newListCmd(getClient func(cmd *cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [run-id]",
		Short: "Получить список тестов в ране",
		Long: `Получает список всех тестов для указанного тест-рана.

Можно применять фильтры:
	--status-id      Фильтр по статусу (1=passed, 5=failed, etc.)
	--assigned-to    Фильтр по назначенному пользователю

Примеры:
	# Получить все тесты в ране
	gotr test list 100

	# Получить только failed тесты
	gotr test list 100 --status-id 5

	# Получить тесты, назначенные на пользователя
	gotr test list 100 --assigned-to 10

	# Сохранить в файл
	gotr test list 100 -o tests.json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			httpClient := getClient(cmd)
			if httpClient == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			svc := service.NewTestService(httpClient)
			runID, err := svc.ParseID(args, 0)
			if err != nil {
				return fmt.Errorf("некорректный ID рана: %w", err)
			}

			// Собираем фильтры
			filters := make(map[string]string)

			if cmd.Flags().Changed("status-id") {
				statusID, _ := cmd.Flags().GetInt64("status-id")
				filters["status_id"] = strconv.FormatInt(statusID, 10)
			}

			if cmd.Flags().Changed("assigned-to") {
				assignedTo, _ := cmd.Flags().GetInt64("assigned-to")
				filters["assignedto_id"] = strconv.FormatInt(assignedTo, 10)
			}

			tests, err := svc.GetForRun(runID, filters)
			if err != nil {
				return fmt.Errorf("ошибка получения списка тестов: %w", err)
			}

			// Проверяем нужно ли сохранить в файл
			output, _ := cmd.Flags().GetString("output")
			if output != "" {
				jsonBytes, err := json.MarshalIndent(tests, "", "  ")
				if err != nil {
					return fmt.Errorf("ошибка сериализации: %w", err)
				}
				if err := os.WriteFile(output, jsonBytes, 0644); err != nil {
					return fmt.Errorf("ошибка записи файла: %w", err)
				}
				svc.PrintSuccess(cmd, "Список тестов (%d) сохранён в %s", len(tests), output)
				return nil
			}

			return svc.Output(cmd, tests)
		},
	}

	cmd.Flags().StringP("output", "o", "", "Сохранить ответ в файл")
	cmd.Flags().BoolP("quiet", "q", false, "Тихий режим")
	cmd.Flags().Int64("status-id", 0, "Фильтр по ID статуса")
	cmd.Flags().Int64("assigned-to", 0, "Фильтр по ID назначенного пользователя")

	return cmd
}
