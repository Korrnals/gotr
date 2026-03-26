package test

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
)

func resolveRunIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)

	// Выбираем проект
	projectID, err := interactive.SelectProject(ctx, p, cli, "")
	if err != nil {
		return 0, err
	}

	// Выбираем ран
	runs, err := cli.GetRuns(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get runs for project %d: %w", projectID, err)
	}
	if len(runs) == 0 {
		return 0, fmt.Errorf("no runs found for project %d", projectID)
	}

	runID, err := interactive.SelectRun(ctx, p, runs, "")
	if err != nil {
		return 0, err
	}

	return runID, nil
}

func resolveTestIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)

	// Выбираем проект
	projectID, err := interactive.SelectProject(ctx, p, cli, "")
	if err != nil {
		return 0, err
	}

	// Выбираем ран
	runs, err := cli.GetRuns(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get runs for project %d: %w", projectID, err)
	}
	if len(runs) == 0 {
		return 0, fmt.Errorf("no runs found for project %d", projectID)
	}

	runID, err := interactive.SelectRun(ctx, p, runs, "")
	if err != nil {
		return 0, err
	}

	// Получаем тесты для рана
	tests, err := cli.GetTests(ctx, runID, map[string]string{})
	if err != nil {
		return 0, fmt.Errorf("failed to get tests for run %d: %w", runID, err)
	}
	if len(tests) == 0 {
		return 0, fmt.Errorf("no tests found for run %d", runID)
	}

	// Возвращаем ID первого теста (при нескольких тестах можно расширить)
	return tests[0].ID, nil
}
