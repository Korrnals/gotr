package bdds

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
)

func resolveCaseIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)

	projectID, err := interactive.SelectProject(ctx, p, cli, "")
	if err != nil {
		return 0, err
	}

	suiteID, err := interactive.SelectSuiteForProject(ctx, p, cli, projectID, "")
	if err != nil {
		return 0, err
	}

	cases, err := cli.GetCases(ctx, projectID, suiteID, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to get cases for project %d suite %d: %w", projectID, suiteID, err)
	}
	if len(cases) == 0 {
		return 0, fmt.Errorf("no cases found in project %d suite %d", projectID, suiteID)
	}

	return selectCaseID(ctx, cases)
}

func selectCaseID(ctx context.Context, cases data.GetCasesResponse) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(cases))
	for i, c := range cases {
		options = append(options, fmt.Sprintf("[%d] ID: %d | %s", i+1, c.ID, c.Title))
	}

	idx, _, err := p.Select("Select case:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select case: %w", err)
	}

	return cases[idx].ID, nil
}
