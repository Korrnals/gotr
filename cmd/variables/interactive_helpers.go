package variables

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
)

// resolveDatasetIDInteractive prompts the user to select a project, then a dataset.
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

	return selectDatasetID(ctx, datasets)
}

// resolveVariableIDInteractive prompts the user to select a dataset, then a variable.
func resolveVariableIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	datasetID, err := resolveDatasetIDInteractive(ctx, cli)
	if err != nil {
		return 0, err
	}

	variables, err := cli.GetVariables(ctx, datasetID)
	if err != nil {
		return 0, fmt.Errorf("failed to get variables for dataset %d: %w", datasetID, err)
	}
	if len(variables) == 0 {
		return 0, fmt.Errorf("no variables found in dataset %d", datasetID)
	}

	return selectVariableID(ctx, variables)
}

// selectDatasetID presents a selection prompt for datasets.
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

// selectVariableID presents a selection prompt for variables.
func selectVariableID(ctx context.Context, variables data.GetVariablesResponse) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(variables))
	for i, variable := range variables {
		options = append(options, fmt.Sprintf("[%d] ID: %d | %s", i+1, variable.ID, variable.Name))
	}

	idx, _, err := p.Select("Select variable:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select variable: %w", err)
	}

	return variables[idx].ID, nil
}
