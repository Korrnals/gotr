package result

import (
	"fmt"
	"strconv"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// projectSelector интерфейс для выбора проекта (для тестирования)
type projectSelector interface {
	SelectProjectInteractively(httpClient client.ClientInterface) (int64, error)
}

// runSelector интерфейс для выбора run (для тестирования)
type runSelector interface {
	SelectRunInteractively(runs data.GetRunsResponse) (int64, error)
}

// defaultSelectors используется по умолчанию
type defaultSelectors struct{}

func (d *defaultSelectors) SelectProjectInteractively(httpClient client.ClientInterface) (int64, error) {
	return interactive.SelectProjectInteractively(httpClient)
}

func (d *defaultSelectors) SelectRunInteractively(runs data.GetRunsResponse) (int64, error) {
	return interactive.SelectRunInteractively(runs)
}

// selectors для интерактивного выбора (можно заменить в тестах)
var selectors projectSelector = &defaultSelectors{}
var runSelectors runSelector = &defaultSelectors{}

// newListCmd создаёт команду 'result list'
// Эндпоинт: GET /get_results_for_run/{run_id}
func newListCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
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
			cli := getClient(cmd)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			svc := newResultServiceFromInterface(cli)

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
				projectID, err := selectors.SelectProjectInteractively(cli)
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
				runID, err = runSelectors.SelectRunInteractively(runs)
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
}

// Обратная совместимость: глобальная переменная для использования в result.go
var listCmd = newListCmd(func(cmd *cobra.Command) client.ClientInterface { return getClientSafe(cmd) })
