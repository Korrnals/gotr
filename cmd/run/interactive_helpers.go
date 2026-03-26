package run

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
)

func resolveRunID(ctx context.Context, cli client.ClientInterface, currentArgs []string) (int64, error) {
	svc := newRunServiceFromInterface(cli)
	if len(currentArgs) > 0 {
		return svc.ParseID(ctx, currentArgs, 0)
	}

	p := interactive.PrompterFromContext(ctx)
	projectID, err := interactive.SelectProject(ctx, p, cli, "")
	if err != nil {
		return 0, err
	}

	runs, err := cli.GetRuns(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get runs for project %d: %w", projectID, err)
	}

	runID, err := interactive.SelectRun(ctx, p, runs, "")
	if err != nil {
		return 0, err
	}

	return runID, nil
}
