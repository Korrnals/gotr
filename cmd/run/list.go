package run

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/service"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [project-id]",
	Short: "Получить список test runs проекта",
	Long: `Получает список всех test runs для указанного проекта.

В списке содержатся активные и завершённые runs с базовой информацией:
ID, название, описание, статистика тестов (passed/failed/blocked).

Если project-id не указан, будет предложен интерактивный выбор из списка проектов.

Примеры:
	# Получить список runs проекта (с интерактивным выбором)
	gotr run list

	# Получить список runs проекта (с явным ID)
	gotr run list 30

	# Сохранить в файл для дальнейшей обработки
	gotr run list 30 -o runs.json
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		httpClient := getClientSafe(cmd)
		if httpClient == nil {
			return fmt.Errorf("HTTP клиент не инициализирован")
		}

		svc := service.NewRunService(httpClient)

		var projectID int64
		var err error

		if len(args) > 0 {
			// Явно указан project-id
			projectID, err = strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("некорректный ID проекта: %w", err)
			}
		} else {
			// Интерактивный выбор проекта
			projectID, err = interactive.SelectProjectInteractively(httpClient)
			if err != nil {
				return err
			}
		}

		runs, err := svc.GetByProject(projectID)
		if err != nil {
			return fmt.Errorf("ошибка получения списка test runs: %w", err)
		}

		return svc.Output(cmd, runs)
	},
}
