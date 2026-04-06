package configurations

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

// resolveGroupIDInteractive prompts the user to select a configuration group interactively.
func resolveGroupIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	projectID, err := resolveProjectIDInteractive(ctx, cli)
	if err != nil {
		return 0, err
	}

	groups, err := cli.GetConfigs(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get configuration groups for project %d: %w", projectID, err)
	}
	if len(groups) == 0 {
		return 0, fmt.Errorf("no configuration groups found in project %d", projectID)
	}

	return selectGroupID(ctx, groups)
}

// resolveConfigIDInteractive prompts the user to select a configuration interactively.
func resolveConfigIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	projectID, err := resolveProjectIDInteractive(ctx, cli)
	if err != nil {
		return 0, err
	}

	groups, err := cli.GetConfigs(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get configuration groups for project %d: %w", projectID, err)
	}
	if len(groups) == 0 {
		return 0, fmt.Errorf("no configuration groups found in project %d", projectID)
	}

	groupIdx, err := selectGroupIndex(ctx, groups)
	if err != nil {
		return 0, err
	}

	configs := groups[groupIdx].Configs
	if len(configs) == 0 {
		return 0, fmt.Errorf("no configurations found in group %d", groups[groupIdx].ID)
	}

	return selectConfigID(ctx, configs)
}

// selectGroupID prompts for group selection and returns the chosen group ID.
func selectGroupID(ctx context.Context, groups data.GetConfigsResponse) (int64, error) {
	idx, err := selectGroupIndex(ctx, groups)
	if err != nil {
		return 0, err
	}
	return groups[idx].ID, nil
}

// selectGroupIndex prompts for group selection and returns the chosen index.
func selectGroupIndex(ctx context.Context, groups data.GetConfigsResponse) (int, error) {
	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(groups))
	for i, group := range groups {
		options = append(options, fmt.Sprintf("[%d] ID: %d | %s", i+1, group.ID, group.Name))
	}

	idx, _, err := p.Select("Select configuration group:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select configuration group: %w", err)
	}

	return idx, nil
}

// selectConfigID prompts for configuration selection and returns the chosen config ID.
func selectConfigID(ctx context.Context, configs []data.Config) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(configs))
	for i, cfg := range configs {
		options = append(options, fmt.Sprintf("[%d] ID: %d | %s", i+1, cfg.ID, cfg.Name))
	}

	idx, _, err := p.Select("Select configuration:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select configuration: %w", err)
	}

	return configs[idx].ID, nil
}
