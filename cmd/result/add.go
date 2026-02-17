package result

import (
	"fmt"
	"os"

	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newAddCmd создаёт команду 'result add'
// Эндпоинт: POST /add_result/{test_id}
func newAddCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [test-id]",
		Short: "Добавить результат для test",
		Long: `Добавляет результат выполнения для указанного test ID.

Статусы результатов (стандартные):
	1 — Passed
	2 — Blocked
	3 — Untested
	4 — Retest
	5 — Failed

Можно указать: комментарий, затраченное время, версию ПО,
дефекты (через запятую), назначение на пользователя.

Примеры:
	# Успешно пройденный тест
	gotr result add 12345 --status-id 1 --comment "All checks passed"

	# Не пройденный тест с дефектом
	gotr result add 12345 --status-id 5 --comment "Bug found" --defects "BUG-123"

	# С временем выполнения и версией
	gotr result add 12345 --status-id 1 --elapsed "2m 30s" --version "v2.0.1"

	# Переназначить на другого пользователя
	gotr result add 12345 --status-id 2 --assigned-to 10 \\
		--comment "Need re-test by another engineer"

	# Dry-run режим
	gotr result add 12345 --status-id 1 --comment "Test" --dry-run`,
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

			req, err := buildAddResultRequest(cmd)
			if err != nil {
				return err
			}

			// Проверяем dry-run режим
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := output.NewDryRunPrinter("result add")
				dr.PrintOperation(
					fmt.Sprintf("Add Result for Test %d", testID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/add_result/%d", testID),
					req,
				)
				return nil
			}

			result, err := svc.AddForTest(testID, req)
			if err != nil {
				return fmt.Errorf("ошибка добавления результата: %w", err)
			}

			svc.PrintSuccess(cmd, "Результат добавлен успешно:")
			return svc.Output(cmd, result)
		},
	}

	cmd.Flags().Int64("status-id", 0, "ID статуса результата (обязательный)")
	cmd.Flags().String("comment", "", "Комментарий к результату")
	cmd.Flags().String("version", "", "Версия ПО")
	cmd.Flags().String("elapsed", "", "Затраченное время (например: '1m 30s')")
	cmd.Flags().String("defects", "", "ID дефектов (через запятую)")
	cmd.Flags().Int64("assigned-to", 0, "ID пользователя для назначения")
	cmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")
	cmd.MarkFlagRequired("status-id")

	return cmd
}

// newAddCaseCmd создаёт команду 'result add-case'
// Эндпоинт: POST /add_result_for_case/{run_id}/{case_id}
func newAddCaseCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-case [run-id]",
		Short: "Добавить результат для кейса в run",
		Long: `Добавляет результат выполнения для указанного кейса в test run.

Отличие от 'add': здесь указывается run_id и case_id, а не test_id.
TestRail сам находит соответствующий test в run.

Примеры:
	# Добавить результат для кейса 98765 в run 12345
	gotr result add-case 12345 --case-id 98765 --status-id 1 \\
		--comment "Smoke test passed"

	# Указать дефект и время
	gotr result add-case 12345 --case-id 98765 --status-id 5 \\
		--defects "JIRA-456" --elapsed "5m"

	# Dry-run режим
	gotr result add-case 12345 --case-id 98765 --status-id 1 --dry-run`,
		Args: cobra.ExactArgs(1),
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

			caseID, _ := cmd.Flags().GetInt64("case-id")
			req, err := buildAddResultRequest(cmd)
			if err != nil {
				return err
			}

			// Проверяем dry-run режим
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := output.NewDryRunPrinter("result add-case")
				dr.PrintOperation(
					fmt.Sprintf("Add Result for Case %d in Run %d", caseID, runID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/add_result_for_case/%d/%d", runID, caseID),
					req,
				)
				return nil
			}

			result, err := svc.AddForCase(runID, caseID, req)
			if err != nil {
				return fmt.Errorf("ошибка добавления результата: %w", err)
			}

			svc.PrintSuccess(cmd, "Результат добавлен успешно:")
			return svc.Output(cmd, result)
		},
	}

	cmd.Flags().Int64("case-id", 0, "ID тест-кейса (обязательный)")
	cmd.Flags().Int64("status-id", 0, "ID статуса результата (обязательный)")
	cmd.Flags().String("comment", "", "Комментарий к результату")
	cmd.Flags().String("version", "", "Версия ПО")
	cmd.Flags().String("elapsed", "", "Затраченное время")
	cmd.Flags().String("defects", "", "ID дефектов (через запятую)")
	cmd.Flags().Int64("assigned-to", 0, "ID пользователя для назначения")
	cmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")
	cmd.MarkFlagRequired("case-id")
	cmd.MarkFlagRequired("status-id")

	return cmd
}

// newAddBulkCmd создаёт команду 'result add-bulk'
// Эндпоинт: POST /add_results/{run_id}
func newAddBulkCmd(getClient func(*cobra.Command) client.ClientInterface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-bulk [run-id]",
		Short: "Массовое добавление результатов",
		Long: `Добавляет несколько результатов одним запросом.

JSON файл должен содержать массив результатов:
[
  {
    "test_id": 12345,
    "status_id": 1,
    "comment": "Test passed"
  },
  {
    "case_id": 98765,
    "status_id": 5,
    "comment": "Test failed",
    "defects": "BUG-123"
  }
]

Поддерживаются оба формата: с test_id и с case_id.

Примеры:
	# Dry-run режим
	gotr result add-bulk 12345 --results-file results.json --dry-run`,
		Args: cobra.ExactArgs(1),
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

			resultsFile, _ := cmd.Flags().GetString("results-file")
			fileData, err := os.ReadFile(resultsFile)
			if err != nil {
				return fmt.Errorf("ошибка чтения файла: %w", err)
			}

			// Проверяем dry-run режим
			isDryRun, _ := cmd.Flags().GetBool("dry-run")
			if isDryRun {
				dr := output.NewDryRunPrinter("result add-bulk")
				dr.PrintOperation(
					fmt.Sprintf("Add Bulk Results for Run %d", runID),
					"POST",
					fmt.Sprintf("/index.php?/api/v2/add_results/%d", runID),
					string(fileData),
				)
				return nil
			}

			// Пытаемся распарсить и отправить
			results, err := svc.AddBulkResults(runID, fileData)
			if err != nil {
				return err
			}

			svc.PrintSuccess(cmd, "Результаты добавлены успешно:")
			return svc.Output(cmd, results)
		},
	}

	cmd.Flags().String("results-file", "", "JSON файл с результатами (обязательный)")
	cmd.Flags().Bool("dry-run", false, "Показать что будет выполнено без реальных изменений")
	cmd.MarkFlagRequired("results-file")

	return cmd
}

// buildAddResultRequest собирает запрос из флагов
func buildAddResultRequest(cmd *cobra.Command) (*data.AddResultRequest, error) {
	// Проверяем что status-id указан (обязательный параметр)
	if !cmd.Flags().Changed("status-id") {
		return nil, fmt.Errorf("--status-id обязателен (используйте: 1=Passed, 2=Blocked, 3=Untested, 4=Retest, 5=Failed)")
	}

	statusID, _ := cmd.Flags().GetInt64("status-id")
	comment, _ := cmd.Flags().GetString("comment")
	version, _ := cmd.Flags().GetString("version")
	elapsed, _ := cmd.Flags().GetString("elapsed")
	defects, _ := cmd.Flags().GetString("defects")
	assignedTo, _ := cmd.Flags().GetInt64("assigned-to")

	return &data.AddResultRequest{
		StatusID:   statusID,
		Comment:    comment,
		Version:    version,
		Elapsed:    elapsed,
		Defects:    defects,
		AssignedTo: assignedTo,
	}, nil
}

// Обратная совместимость: глобальные переменные для использования в result.go
var (
	addCmd     = newAddCmd(func(cmd *cobra.Command) client.ClientInterface { return getClientSafe(cmd) })
	addCaseCmd = newAddCaseCmd(func(cmd *cobra.Command) client.ClientInterface { return getClientSafe(cmd) })
	addBulkCmd = newAddBulkCmd(func(cmd *cobra.Command) client.ClientInterface { return getClientSafe(cmd) })
)
