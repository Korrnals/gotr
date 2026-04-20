package bdds

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
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

	return interactive.SelectCase(ctx, p, cases, "")
}
