package result

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// newGetCmd создаёт команду 'result get'
// Эндпоинт: GET /get_results/{test_id}
func newGetCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "get [test-id]",
		Short: "Получить результаты для test",
		Long: `Получает список результатов для указанного test ID.

Test — это экземпляр тест-кейса в конкретном test run.
Результаты показывают историю выполнения: статус, комментарии,
затраченное время, версию ПО, дефекты.

Примеры:
	# Получить результаты конкретного теста
	gotr result get 12345

	# Сохранить результаты в файл для анализа
	gotr result get 12345 -o test_results.json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			svc := newResultServiceFromInterface(cli)
			testID, err := svc.ParseID(args, 0)
			if err != nil {
				return fmt.Errorf("некорректный ID test: %w", err)
			}

			results, err := svc.GetForTest(testID)
			if err != nil {
				return fmt.Errorf("ошибка получения результатов: %w", err)
			}

			return svc.Output(cmd, results)
		},
	}
}

// newGetCaseCmd создаёт команду 'result get-case'
// Эндпоинт: GET /get_results_for_case/{run_id}/{case_id}
func newGetCaseCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	return &cobra.Command{
		Use:   "get-case [run-id] [case-id]",
		Short: "Получить результаты для кейса в run",
		Long: `Получает список результатов для указанного кейса в test run.

Удобно, когда нужно посмотреть историю выполнения конкретного кейса
без необходимости знать test_id. Используется комбинация run_id + case_id.

Примеры:
	# Получить результаты кейса 98765 в run 12345
	gotr result get-case 12345 98765

	# Сохранить в файл
	gotr result get-case 12345 98765 -o case_results.json
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			if cli == nil {
				return fmt.Errorf("HTTP клиент не инициализирован")
			}

			svc := newResultServiceFromInterface(cli)
			runID, err := svc.ParseID(args, 0)
			if err != nil {
				return fmt.Errorf("некорректный ID run: %w", err)
			}

			caseID, err := svc.ParseID(args, 1)
			if err != nil {
				return fmt.Errorf("некорректный ID case: %w", err)
			}

			results, err := svc.GetForCase(runID, caseID)
			if err != nil {
				return fmt.Errorf("ошибка получения результатов: %w", err)
			}

			return svc.Output(cmd, results)
		},
	}
}

// Обратная совместимость: глобальные переменные для использования в result.go
var (
	getCmd     = newGetCmd(func(cmd *cobra.Command) client.ClientInterface { return getClientSafe(cmd) })
	getCaseCmd = newGetCaseCmd(func(cmd *cobra.Command) client.ClientInterface { return getClientSafe(cmd) })
)
