// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package labels

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
)

// resolveProjectIDInteractive selects a project interactively.
func resolveProjectIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	return interactive.SelectProject(ctx, p, cli, "")
}

// resolveLabelIDInteractive selects a label interactively: project → labels → select.
func resolveLabelIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	projectID, err := interactive.SelectProject(ctx, p, cli, "")
	if err != nil {
		return 0, err
	}
	labels, err := cli.GetLabels(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get labels: %w", err)
	}
	return selectLabelID(ctx, labels)
}

// selectLabelID lets the user choose a label from a list.
func selectLabelID(ctx context.Context, labels data.GetLabelsResponse) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	if len(labels) == 0 {
		return 0, fmt.Errorf("no labels found")
	}
	items := make([]string, len(labels))
	for i, l := range labels {
		items[i] = fmt.Sprintf("[%d] ID: %d | %s", i+1, l.ID, l.Name)
	}
	idx, _, err := p.Select("Select label:", items)
	if err != nil {
		return 0, fmt.Errorf("failed to select label: %w", err)
	}
	return labels[idx].ID, nil
}

// resolveTestIDInteractive selects a test interactively: project → run → tests → select.
func resolveTestIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	projectID, err := interactive.SelectProject(ctx, p, cli, "")
	if err != nil {
		return 0, err
	}
	runs, err := cli.GetRuns(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get runs: %w", err)
	}
	runID, err := interactive.SelectRun(ctx, p, runs, "")
	if err != nil {
		return 0, err
	}
	tests, err := cli.GetTests(ctx, runID, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get tests: %w", err)
	}
	return selectTestID(ctx, tests)
}

// selectTestID lets the user choose a test from a list.
func selectTestID(ctx context.Context, tests []data.Test) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	if len(tests) == 0 {
		return 0, fmt.Errorf("no tests found")
	}
	items := make([]string, len(tests))
	for i, t := range tests {
		items[i] = fmt.Sprintf("[%d] ID: %d | %s", i+1, t.ID, t.Title)
	}
	idx, _, err := p.Select("Select test:", items)
	if err != nil {
		return 0, fmt.Errorf("failed to select test: %w", err)
	}
	return tests[idx].ID, nil
}
