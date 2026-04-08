package result

import (
	"context"
	"fmt"
	"sort"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/spf13/cobra"
)

// newGetCmd creates the 'result get' command.
// Endpoint: GET /get_results/{test_id}
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
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			svc := newResultServiceFromInterface(cli)

			var testID int64
			var err error
			if len(args) > 0 {
				testID, err = svc.ParseID(ctx, args, 0)
				if err != nil {
					return fmt.Errorf("invalid test ID: %w", err)
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("test_id required: gotr result get [test-id]")
				}
				runID, err := resolveResultRunID(ctx, cli)
				if err != nil {
					return err
				}
				testID, err = selectTestIDForRun(ctx, cli, runID)
				if err != nil {
					return err
				}
			}

			results, err := svc.GetForTest(ctx, testID)
			if err != nil {
				return fmt.Errorf("failed to get results: %w", err)
			}

			return output.OutputResultWithFlags(cmd, results)
		},
	}
}

// newGetCaseCmd creates the 'result get-case' command.
// Endpoint: GET /get_results_for_case/{run_id}/{case_id}
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
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := getClient(cmd)
			ctx := cmd.Context()
			if cli == nil {
				return fmt.Errorf("HTTP client not initialized")
			}

			svc := newResultServiceFromInterface(cli)

			var runID int64
			var caseID int64
			var err error

			if len(args) > 0 {
				runID, err = svc.ParseID(ctx, args, 0)
				if err != nil {
					return fmt.Errorf("invalid run ID: %w", err)
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("run_id required: gotr result get-case [run-id] [case-id]")
				}
				runID, err = resolveResultRunID(ctx, cli)
				if err != nil {
					return err
				}
			}

			if len(args) > 1 {
				caseID, err = svc.ParseID(ctx, args, 1)
				if err != nil {
					return fmt.Errorf("invalid case ID: %w", err)
				}
			} else {
				if !interactive.HasPrompterInContext(ctx) {
					return fmt.Errorf("case_id required: gotr result get-case [run-id] [case-id]")
				}
				caseID, err = selectCaseIDForRun(ctx, cli, runID)
				if err != nil {
					return err
				}
			}

			results, err := svc.GetForCase(ctx, runID, caseID)
			if err != nil {
				return fmt.Errorf("failed to get results: %w", err)
			}

			return output.OutputResultWithFlags(cmd, results)
		},
	}
}

func resolveResultRunID(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	projectID, err := interactive.SelectProject(ctx, p, cli, "")
	if err != nil {
		return 0, err
	}

	runs, err := cli.GetRuns(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get runs list: %w", err)
	}

	if len(runs) == 0 {
		return 0, fmt.Errorf("no test runs found in project %d", projectID)
	}

	return interactive.SelectRun(ctx, p, runs, "")
}

func selectTestIDForRun(ctx context.Context, cli client.ClientInterface, runID int64) (int64, error) {
	tests, err := cli.GetTests(ctx, runID, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get tests for run %d: %w", runID, err)
	}
	if len(tests) == 0 {
		return 0, fmt.Errorf("no tests found in run %d", runID)
	}

	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(tests))
	for i, test := range tests {
		options = append(options, fmt.Sprintf("[%d] ID: %d | Case: %d | %s", i+1, test.ID, test.CaseID, test.Title))
	}

	idx, _, err := p.Select("Select test:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select test: %w", err)
	}

	return tests[idx].ID, nil
}

func selectCaseIDForRun(ctx context.Context, cli client.ClientInterface, runID int64) (int64, error) {
	tests, err := cli.GetTests(ctx, runID, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get tests for run %d: %w", runID, err)
	}
	if len(tests) == 0 {
		return 0, fmt.Errorf("no tests found in run %d", runID)
	}

	byCase := make(map[int64]data.Test)
	for _, test := range tests {
		if _, exists := byCase[test.CaseID]; !exists {
			byCase[test.CaseID] = test
		}
	}

	caseIDs := make([]int64, 0, len(byCase))
	for caseID := range byCase {
		caseIDs = append(caseIDs, caseID)
	}
	sort.Slice(caseIDs, func(i, j int) bool { return caseIDs[i] < caseIDs[j] })

	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(caseIDs))
	for i, caseID := range caseIDs {
		test := byCase[caseID]
		options = append(options, fmt.Sprintf("[%d] Case: %d | Test: %d | %s", i+1, caseID, test.ID, test.Title))
	}

	idx, _, err := p.Select("Select case:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select case: %w", err)
	}

	return caseIDs[idx], nil
}

// Backward compatibility: exported vars for registration in result.go
var (
	getCmd     = newGetCmd(getClientSafe)
	getCaseCmd = newGetCaseCmd(getClientSafe)
)
