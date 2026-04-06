package tests

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
)

// resolveRunIDInteractive prompts the user to select a project, then a run.
func resolveRunIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)

	projectID, err := interactive.SelectProject(ctx, p, cli, "")
	if err != nil {
		return 0, err
	}

	runs, err := cli.GetRuns(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get runs for project %d: %w", projectID, err)
	}
	if len(runs) == 0 {
		return 0, fmt.Errorf("no runs found for project %d", projectID)
	}

	return interactive.SelectRun(ctx, p, runs, "")
}

// resolveTestIDInteractive prompts the user to select a run, then a test.
func resolveTestIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)

	runID, err := resolveRunIDInteractive(ctx, cli)
	if err != nil {
		return 0, err
	}

	tests, err := cli.GetTests(ctx, runID, map[string]string{})
	if err != nil {
		return 0, fmt.Errorf("failed to get tests for run %d: %w", runID, err)
	}
	if len(tests) == 0 {
		return 0, fmt.Errorf("no tests found for run %d", runID)
	}

	options := make([]string, 0, len(tests))
	for i, test := range tests {
		options = append(options, fmt.Sprintf("[%d] ID: %d | %s", i+1, test.ID, test.Title))
	}

	idx, _, err := p.Select("Select test:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select test: %w", err)
	}

	return tests[idx].ID, nil
}
