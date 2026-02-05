package result

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/service"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [run-id]",
	Short: "Получить результаты для test run",
	Long: `Получает список результатов для указанного test run.

Если run-id не указан, будет предложен интерактивный выбор:
1. Выбор проекта из списка
2. Выбор test run из проекта

Примеры:
	# Получить результаты с интерактивным выбором run
	gotr result list

	# Получить результаты для конкретного run
	gotr result list 12345

	# Сохранить в файл
	gotr result list 12345 -o results.json
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		httpClient := getClientSafe(cmd)
		if httpClient == nil {
			return fmt.Errorf("HTTP клиент не инициализирован")
		}

		svc := service.NewResultService(httpClient)

		var runID int64
		var err error

		if len(args) > 0 {
			// Явно указан run-id
			runID, err = strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("некорректный ID run: %w", err)
			}
		} else {
			// Интерактивный выбор: проект → run
			projectID, err := interactive.SelectProjectInteractively(httpClient)
			if err != nil {
				return err
			}

			// Получаем список runs проекта
			runs, err := svc.GetRunsForProject(projectID)
			if err != nil {
				return fmt.Errorf("ошибка получения списка runs: %w", err)
			}

			if len(runs) == 0 {
				return fmt.Errorf("в проекте %d не найдено test runs", projectID)
			}

			// Выбираем run интерактивно
			runID, err = interactive.SelectRunInteractively(runs)
			if err != nil {
				return err
			}
		}

		results, err := svc.GetForRun(runID)
		if err != nil {
			return fmt.Errorf("ошибка получения результатов: %w", err)
		}

		return svc.Output(cmd, results)
	},
}
