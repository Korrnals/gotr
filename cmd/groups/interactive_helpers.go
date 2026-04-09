package groups

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

// resolveGroupIDInteractive prompts the user to select a user group interactively.
func resolveGroupIDInteractive(ctx context.Context, cli client.ClientInterface) (int64, error) {
	projectID, err := resolveProjectIDInteractive(ctx, cli)
	if err != nil {
		return 0, err
	}

	groups, err := cli.GetGroups(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to get groups for project %d: %w", projectID, err)
	}
	if len(groups) == 0 {
		return 0, fmt.Errorf("no groups found in project %d", projectID)
	}

	return selectGroupID(ctx, groups)
}

// selectGroupID prompts for group selection and returns the chosen group ID.
func selectGroupID(ctx context.Context, groups data.GetGroupsResponse) (int64, error) {
	p := interactive.PrompterFromContext(ctx)
	options := make([]string, 0, len(groups))
	for i, group := range groups {
		options = append(options, fmt.Sprintf("[%d] ID: %d | %s", i+1, group.ID, group.Name))
	}

	idx, _, err := p.Select("Select group:", options)
	if err != nil {
		return 0, fmt.Errorf("failed to select group: %w", err)
	}

	return groups[idx].ID, nil
}
