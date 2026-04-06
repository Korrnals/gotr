package datasets

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
)

// resolveProjectIDInteractive prompts the user to select a project interactively.
func resolveProjectIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	return interactive.SelectProject(ctx, p, cli, "")
}

// resolveDatasetIDInteractive prompts the user to select a dataset interactively.
func resolveDatasetIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	projectID, err := resolveProjectIDInteractive(ctx, cli)
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

	return selectDatasetID(ctx, datasets)
}

// selectDatasetID prompts for dataset selection and returns the chosen dataset ID.
func selectDatasetID(ctx context.Context, datasets data.GetDatasetsResponse) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(datasets))
	for i, dataset := range datasets {
		options = append(options, fmt.Sprintf("[%d] ID: %d | %s", i+1, dataset.ID, dataset.Name))
	}

	idx, _, err := p.Select("Select dataset:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select dataset: %w", err)
	}

	return datasets[idx].ID, nil
}