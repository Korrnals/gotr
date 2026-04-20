package datasets

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
)

// resolveProjectIDInteractive prompts the user to select a project interactively.
func resolveProjectIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	return interactive.SelectProject(ctx, p, cli, "")
}

// resolveDatasetIDInteractive prompts the user to select a dataset interactively.
func resolveDatasetIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	projectID, err := interactive.SelectProject(ctx, p, cli, "")
	if err != nil {
		return 0, err
	}

	datasets, err := cli.GetDatasets(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get datasets for project %d: %w", projectID, err)
	}
	if len(datasets) == 0 {
		return 0, fmt.Errorf("no datasets found in project %d", projectID)
	}

	return interactive.SelectDataset(ctx, p, datasets, "")
}