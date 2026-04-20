package reports

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
)

func resolveProjectIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	return interactive.SelectProject(ctx, p, cli, "")
}

func resolveReportTemplateIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	projectID, err := resolveProjectIDInteractive(ctx, cli)
	if err != nil {
		return 0, err
	}

	reports, err := cli.GetReports(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to list reports for project %d: %w", projectID, err)
	}
	if len(reports) == 0 {
		return 0, fmt.Errorf("no report templates found in project %d", projectID)
	}

	return selectReportTemplateID(ctx, reports)
}

func resolveCrossProjectReportTemplateIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	reports, err := cli.GetCrossProjectReports(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list cross-project reports: %w", err)
	}
	if len(reports) == 0 {
		return 0, fmt.Errorf("no cross-project report templates found")
	}

	return selectReportTemplateID(ctx, reports)
}

func selectReportTemplateID(ctx context.Context, reports data.GetReportsResponse) (int64, error) {
	p := interactive.PrompterFromContext(ctx)

	cols := []interactive.Column{
		{Header: "ID", MinWidth: 6},
		{Header: "Name"},
	}
	rows := make([][]string, len(reports))
	for i, r := range reports {
		rows[i] = []string{fmt.Sprintf("%d", r.ID), r.Name}
	}
	header, options := interactive.AlignedLabels(cols, rows)

	idx, err := interactive.Browse(ctx, p, interactive.BrowseConfig{
		Prompt: "Select report template:",
		Header: header,
		Items:  options,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to select report template: %w", err)
	}

	return reports[idx].ID, nil
}
