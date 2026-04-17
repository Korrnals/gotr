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

	cols := []interactive.Column{
		{Header: "ID", MinWidth: 6},
		{Header: "Name"},
	}
	rows := make([][]string, len(labels))
	for i, l := range labels {
		rows[i] = []string{fmt.Sprintf("%d", l.ID), l.Name}
	}
	header, items := interactive.AlignedLabels(cols, rows)

	idx, err := interactive.Browse(ctx, p, interactive.BrowseConfig{
		Prompt: "Select label:",
		Header: header,
		Items:  items,
	})
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
	return interactive.SelectTest(ctx, p, tests, "")
}
