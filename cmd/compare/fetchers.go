package compare

import (
	"context"
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
)

// FetchFunc is the type for resource fetch functions used by simple compare commands.
type FetchFunc func(ctx context.Context, cli client.ClientInterface, projectID int64) ([]ItemInfo, error)

// --- Fetch functions for all simple resources ---

// fetchGroupItems fetches all groups for a project and returns them as ItemInfo slice.
func fetchGroupItems(ctx context.Context, cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	groups, err := cli.GetGroups(ctx, projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(groups))
	for _, g := range groups {
		items = append(items, ItemInfo{ID: g.ID, Name: g.Name})
	}
	return items, nil
}

// fetchLabelItems fetches all labels for a project and returns them as ItemInfo slice.
func fetchLabelItems(ctx context.Context, cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	labels, err := cli.GetLabels(ctx, projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(labels))
	for _, l := range labels {
		items = append(items, ItemInfo{ID: l.ID, Name: l.Name})
	}
	return items, nil
}

// fetchMilestoneItems fetches all milestones for a project and returns them as ItemInfo slice.
func fetchMilestoneItems(ctx context.Context, cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	milestones, err := cli.GetMilestones(ctx, projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(milestones))
	for _, m := range milestones {
		items = append(items, ItemInfo{ID: m.ID, Name: m.Name})
	}
	return items, nil
}

// fetchPlanItems fetches all plans for a project and returns them as ItemInfo slice.
func fetchPlanItems(ctx context.Context, cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	plans, err := cli.GetPlans(ctx, projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(plans))
	for _, p := range plans {
		items = append(items, ItemInfo{ID: p.ID, Name: p.Name})
	}
	return items, nil
}

// fetchRunItems fetches all runs for a project and returns them as ItemInfo slice.
func fetchRunItems(ctx context.Context, cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	runs, err := cli.GetRuns(ctx, projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(runs))
	for _, r := range runs {
		items = append(items, ItemInfo{ID: r.ID, Name: r.Name})
	}
	return items, nil
}

// fetchSharedStepItems fetches all shared steps for a project and returns them as ItemInfo slice.
func fetchSharedStepItems(ctx context.Context, cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	steps, err := cli.GetSharedSteps(ctx, projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(steps))
	for _, s := range steps {
		items = append(items, ItemInfo{ID: s.ID, Name: s.Title})
	}
	return items, nil
}

// fetchTemplateItems fetches all templates for a project and returns them as ItemInfo slice.
func fetchTemplateItems(ctx context.Context, cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	templates, err := cli.GetTemplates(ctx, projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(templates))
	for _, t := range templates {
		items = append(items, ItemInfo{ID: t.ID, Name: t.Name})
	}
	return items, nil
}

// fetchConfigurationItems fetches all configurations for a project and returns them as ItemInfo slice.
// Includes both config groups and individual configs (formatted as "Group / Config").
func fetchConfigurationItems(ctx context.Context, cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	groups, err := cli.GetConfigs(ctx, projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0)
	for _, group := range groups {
		items = append(items, ItemInfo{ID: group.ID, Name: group.Name})
		for _, config := range group.Configs {
			items = append(items, ItemInfo{
				ID:   config.ID,
				Name: fmt.Sprintf("%s / %s", group.Name, config.Name),
			})
		}
	}
	return items, nil
}

// fetchDatasetItems fetches all datasets for a project and returns them as ItemInfo slice.
func fetchDatasetItems(ctx context.Context, cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	datasets, err := cli.GetDatasets(ctx, projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(datasets))
	for _, d := range datasets {
		items = append(items, ItemInfo{ID: d.ID, Name: d.Name})
	}
	return items, nil
}
