package compare

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/client"
)

// FetchFunc is the type for resource fetch functions used by simple compare commands.
type FetchFunc func(cli client.ClientInterface, projectID int64) ([]ItemInfo, error)

// --- Fetch functions for all simple resources ---

// fetchGroupItems fetches all groups for a project and returns them as ItemInfo slice.
func fetchGroupItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	groups, err := cli.GetGroups(projectID)
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
func fetchLabelItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	labels, err := cli.GetLabels(projectID)
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
func fetchMilestoneItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	milestones, err := cli.GetMilestones(projectID)
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
func fetchPlanItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	plans, err := cli.GetPlans(projectID)
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
func fetchRunItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	runs, err := cli.GetRuns(projectID)
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
func fetchSharedStepItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	steps, err := cli.GetSharedSteps(projectID)
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
func fetchTemplateItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	templates, err := cli.GetTemplates(projectID)
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
func fetchConfigurationItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	groups, err := cli.GetConfigs(projectID)
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
func fetchDatasetItems(cli client.ClientInterface, projectID int64) ([]ItemInfo, error) {
	datasets, err := cli.GetDatasets(projectID)
	if err != nil {
		return nil, err
	}

	items := make([]ItemInfo, 0, len(datasets))
	for _, d := range datasets {
		items = append(items, ItemInfo{ID: d.ID, Name: d.Name})
	}
	return items, nil
}
