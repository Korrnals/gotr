package run

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/service"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [project-id]",
	Short: "Получить список test runs проекта",
	Long: `Получает список всех test runs для указанного проекта.

В списке содержатся активные и завершённые runs с базовой информацией:
ID, название, описание, статистика тестов (passed/failed/blocked).

Примеры:
	# Получить список runs проекта
	gotr run list 30

	# Сохранить в файл для дальнейшей обработки
	gotr run list 30 -o runs.json
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		httpClient := getClientSafe(cmd)
		if httpClient == nil {
			return fmt.Errorf("HTTP клиент не инициализирован")
		}

		svc := service.NewRunService(httpClient)
		projectID, err := svc.ParseID(args, 0)
		if err != nil {
			return fmt.Errorf("некорректный ID проекта: %w", err)
		}

		runs, err := svc.GetByProject(projectID)
		if err != nil {
			return fmt.Errorf("ошибка получения списка test runs: %w", err)
		}

		return svc.Output(cmd, runs)
	},
}
